package lua

import (
	"crypto/rand"
	"fmt"
	"github.com/alttpo/sexp"
	"github.com/yuin/gopher-lua"
	"reflect"
	"strings"
	"testing"
)

func table() *lua.LTable {
	return &lua.LTable{
		Metatable: lua.LNil,
	}
}
func expectToken(s string) *lua.LTable {
	t := table()
	t.RawSetString("token", lua.LString(s))
	return t
}

func expectHex(v []byte) *lua.LTable {
	t := table()
	t.RawSetString("hex", lua.LString(v))
	return t
}

func expectList(children ...*lua.LTable) *lua.LTable {
	t := table()
	list := table()
	for _, c := range children {
		list.Append(c)
	}
	t.RawSetString("list", list)
	return t
}

func fmtLua(v lua.LValue) string {
	if v == nil {
		return ""
	}

	switch v.Type() {
	case lua.LTTable:
		tb := v.(*lua.LTable)
		sb := &strings.Builder{}
		sb.WriteRune('{')
		tb.ForEach(func(key lua.LValue, val lua.LValue) {
			sb.WriteString(fmtLua(key))
			sb.WriteRune('=')
			sb.WriteString(fmtLua(val))
			sb.WriteRune(',')
		})
		s := sb.String()
		return s[0:len(s)-1] + "}"
	case lua.LTString:
		st := string(v.(lua.LString))
		return fmt.Sprintf("%q", st)
	default:
		return v.String()
	}
}

func TestLuaDecoder(t *testing.T) {
	type test struct {
		name    string
		n       *sexp.Node
		nstr    string
		wantErr string
		wantN   lua.LValue
	}
	var cases = []test{
		{
			name:    "(a)",
			n:       sexp.MustList(sexp.MustToken("a")),
			wantErr: "",
			wantN:   expectList(expectToken("a")),
		},
		{
			name: "(a b c)",
			n: sexp.MustList(
				sexp.MustToken("a"),
				sexp.MustToken("b"),
				sexp.MustToken("c"),
			),
			wantErr: "",
			wantN: expectList(
				expectToken("a"),
				expectToken("b"),
				expectToken("c"),
			),
		},
		{
			name: "(a+1 b-2 c/3)",
			n: sexp.MustList(
				sexp.MustToken("a+1"),
				sexp.MustToken("b-2"),
				sexp.MustToken("c/3"),
			),
			wantErr: "",
			wantN: expectList(
				expectToken("a+1"),
				expectToken("b-2"),
				expectToken("c/3"),
			),
		},
		{
			name:    "(",
			nstr:    "(",
			wantErr: "eof",
			wantN:   expectList(),
		},
		{
			name:    ")",
			nstr:    ")",
			wantErr: "unexpected end of list",
			wantN:   lua.LNil,
		},
		{
			name:    "(#aa#)",
			nstr:    "(#aa#)",
			wantErr: "",
			wantN:   expectList(expectHex([]byte{0xaa})),
		},
		{
			name:    "(#aa bbc c#)",
			nstr:    "(#aa bbc c#)",
			wantErr: "",
			wantN:   expectList(expectHex([]byte{0xaa, 0xbb, 0xcc})),
		},
		func() test {
			large := make([]byte, 256)
			rand.Read(large)

			return test{
				name:  "large hex",
				n:     sexp.MustHexadecimal(large),
				wantN: expectHex(large),
			}
		}(),
	}

	l := lua.NewState(lua.Options{})
	defer l.Close()

	// load the tests.lua file:
	var err error
	err = l.DoFile("sexp.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			nstr := tt.nstr
			if tt.n != nil {
				nstr = tt.n.String()
			}

			err = l.CallByParam(
				lua.P{
					Fn:      l.GetGlobal("sexp_decode"),
					NRet:    3,
					Protect: true,
				},
				lua.LString(nstr),
			)
			if err != nil {
				t.Fatalf("glua error: %v", err)
			}

			n, i, perr := l.Get(-3), l.Get(-2), l.Get(-1)
			l.Pop(3)

			errStr := ""
			if perr != lua.LNil {
				errStr = string(perr.(*lua.LTable).RawGetString("err").(lua.LString))
			}

			if (errStr != "") != (tt.wantErr != "") {
				t.Fatalf("want err='%v' got '%v'", tt.wantErr, errStr)
			}

			_ = i

			if !reflect.DeepEqual(tt.wantN, n) {
				t.Fatalf("want %s\ngot  %s", fmtLua(tt.wantN), fmtLua(n))
			}
		})
	}
}

func TestLuaEncoder(t *testing.T) {
	type test struct {
		name    string
		arg     lua.LValue
		wantErr string
		wantN   string
	}
	var cases = []test{
		{
			name:    "(a)",
			arg:     expectList(expectToken("a")),
			wantErr: "",
			wantN:   "(a)",
		},
		{
			name: "(a b c)",
			arg: expectList(
				expectToken("a"),
				expectToken("b"),
				expectToken("c"),
			),
			wantErr: "",
			wantN:   "(a b c)",
		},
	}

	l := lua.NewState(lua.Options{})
	defer l.Close()

	// load the tests.lua file:
	var err error
	err = l.DoFile("sexp.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err = l.CallByParam(
				lua.P{
					Fn:      l.GetGlobal("sexp_encode"),
					NRet:    2,
					Protect: true,
				},
				tt.arg,
			)
			if err != nil {
				t.Fatalf("glua error: %v", err)
			}

			n, perr := l.Get(-2), l.Get(-1)
			l.Pop(2)

			errStr := ""
			if perr != lua.LNil {
				errStr = string(perr.(*lua.LTable).RawGetString("err").(lua.LString))
			}

			if (errStr != "") != (tt.wantErr != "") {
				t.Fatalf("want err='%v' got '%v'", tt.wantErr, errStr)
			}

			nstr := ""
			if nlstr, ok := n.(lua.LString); ok {
				nstr = string(nlstr)
			}

			if tt.wantN != nstr {
				t.Fatalf("want `%s`\ngot  `%s`", tt.wantN, nstr)
			}
		})
	}
}
