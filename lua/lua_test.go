package lua

import (
	"github.com/alttpo/sexp"
	"github.com/yuin/gopher-lua"
	"reflect"
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
	type args struct {
		n *sexp.Node
	}
	type test struct {
		name    string
		args    args
		wantErr string
		wantN   interface{}
	}
	var cases = []test{
		{
			name: "(a)",
			args: args{
				n: sexp.List(sexp.MustToken("a")),
			},
			wantErr: "",
			wantN:   expectList(expectToken("a")),
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

			err = l.CallByParam(
				lua.P{
					Fn:      l.GetGlobal("sexp_parse"),
					NRet:    4,
					Protect: true,
				},
				lua.LString(tt.args.n.String()),
			)
			if err != nil {
				t.Fatalf("glua error: %v", err)
			}

			n, i, eol, perr := l.Get(-4), l.Get(-3), l.Get(-2), l.Get(-1)
			l.Pop(4)

			errStr := ""
			if perr != lua.LNil {
				errStr = string(perr.(*lua.LTable).RawGetString("err").(lua.LString))
			}

			if (errStr != "") != (tt.wantErr != "") {
				t.Fatalf("want err='%v' got '%v'", tt.wantErr, errStr)
			}

			_, _, _ = i, eol, perr

			if !reflect.DeepEqual(tt.wantN, n) {
				t.Fatalf("want n=%#v got %#v", tt.wantN, n)
			}
		})
	}
}
