package sexp

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		s io.RuneScanner
	}
	tests := []struct {
		name    string
		args    args
		wantN   *Node
		wantErr bool
	}{
		{
			name: "xpass: empty list",
			args: args{
				s: strings.NewReader("()"),
			},
			wantN: &Node{
				Kind:        KindList,
				OctetString: nil,
				List:        make([]*Node, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "xfail: mismatched end of list",
			args: args{
				s: strings.NewReader(")"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xfail: mismatched start of list",
			args: args{
				s: strings.NewReader("("),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xpass: list of one token",
			args: args{
				s: strings.NewReader("(abcdef)"),
			},
			wantN: &Node{
				Kind:        KindList,
				OctetString: nil,
				List: []*Node{
					{
						Kind:        KindToken,
						OctetString: []byte("abcdef"),
						List:        nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "xpass: list of one token with whitespace",
			args: args{
				s: strings.NewReader("( a-1*b+c:d=e/f_g.h\t\v\f )"),
			},
			wantN: &Node{
				Kind:        KindList,
				OctetString: nil,
				List: []*Node{
					{
						Kind:        KindToken,
						OctetString: []byte("a-1*b+c:d=e/f_g.h"),
						List:        nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "xpass: single token",
			args: args{
				s: strings.NewReader("abc"),
			},
			wantN: &Node{
				Kind:        KindToken,
				OctetString: []byte("abc"),
				List:        nil,
			},
			wantErr: false,
		},
		{
			name: "xpass: single token with trailing newline",
			args: args{
				s: strings.NewReader("abc\n"),
			},
			wantN: &Node{
				Kind:        KindToken,
				OctetString: []byte("abc"),
				List:        nil,
			},
			wantErr: false,
		},
		{
			name: "xfail: list containing newline",
			args: args{
				s: strings.NewReader("(abc\n)"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xpass: list of two tokens with tab whitespace",
			args: args{
				s: strings.NewReader("(abc\tdef)"),
			},
			wantN: &Node{
				Kind:        KindList,
				OctetString: nil,
				List: []*Node{
					{
						Kind:        KindToken,
						OctetString: []byte("abc"),
						List:        nil,
					},
					{
						Kind:        KindToken,
						OctetString: []byte("def"),
						List:        nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "xpass: hexadecimal",
			args: args{
				s: strings.NewReader("#616263#"),
			},
			wantN: &Node{
				Kind:        KindList,
				OctetString: nil,
				List: []*Node{
					{
						Kind:        KindToken,
						OctetString: []byte("abc"),
						List:        nil,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, err := Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotN, tt.wantN) {
				t.Errorf("Parse() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
