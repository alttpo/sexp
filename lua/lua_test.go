package lua

import (
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
func expectList(children ...*lua.LTable) *lua.LTable {
	t := table()
	list := table()
	for _, c := range children {
		list.Append(c)
	}
	t.RawSetString("list", list)
	return t
}

func TestLuaParser(t *testing.T) {
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
			n:       sexp.List(sexp.MustToken("a")),
			wantErr: "",
			wantN:   expectList(expectToken("a")),
		},
		{
			name: "(a b c)",
			n: sexp.List(
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
			n: sexp.List(
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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			l := lua.NewState(lua.Options{})
			defer l.Close()

			// load the tests.lua file:
			var err error
			err = l.DoFile("sexp.lua")
			if err != nil {
				t.Fatal(err)
			}

			nstr := tt.nstr
			if tt.n != nil {
				nstr = tt.n.String()
			}

			err = l.CallByParam(
				lua.P{
					Fn:      l.GetGlobal("sexp_parse"),
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

			_, _ = i, perr

			if !reflect.DeepEqual(tt.wantN, n) {
				t.Fatalf("want %s\ngot  %s", fmtLua(tt.wantN), fmtLua(n))
			}
		})
	}
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
