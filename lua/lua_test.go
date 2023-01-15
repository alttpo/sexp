package lua

import (
	"github.com/alttpo/sexp"
	"github.com/yuin/gopher-lua"
	"reflect"
	"testing"
)

func TestLuaParser(t *testing.T) {
	type args struct {
		n *sexp.Node
	}
	type test struct {
		name    string
		args    args
		wantErr bool
		wantN   interface{}
	}
	var cases = []test{
		{
			name: "(a)",
			args: args{
				n: sexp.List(sexp.MustToken("a")),
			},
			wantErr: false,
			wantN: func() lua.LValue {
				tbl := &lua.LTable{
					Metatable: lua.LNil,
				}
				tbl.Append(lua.LString("a"))
				return tbl
			}(),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			l := lua.NewState(lua.Options{
				SkipOpenLibs: true,
			})
			defer l.Close()

			// load the tests.lua file:
			var err error
			err = l.DoFile("tests.lua")
			if err != nil {
				t.Fatal(err)
			}

			err = l.CallByParam(
				lua.P{
					Fn:      l.GetGlobal("sexp_parse"),
					NRet:    1,
					Protect: true,
				},
				lua.LString(tt.args.n.String()),
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr %v got %v", tt.wantErr, err)
			}

			ret := l.Get(-1)
			l.Pop(1)

			if !reflect.DeepEqual(tt.wantN, ret) {
				t.Fatalf("want %#v got %#v", tt.wantN, ret)
			}
		})
	}
}
