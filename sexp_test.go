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
			wantN:   MustList(),
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
			wantN:   MustList(MustToken("abcdef")),
			wantErr: false,
		},
		{
			name: "xpass: list of one token with whitespace",
			args: args{
				s: strings.NewReader("( a-1*b+c:d=e/f_g.h\t\v\f )"),
			},
			wantN:   MustList(MustToken("a-1*b+c:d=e/f_g.h")),
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
				Kind:        KindHexadecimal,
				OctetString: []byte("abc"),
				List:        nil,
			},
			wantErr: false,
		},
		{
			name: "xpass: hexadecimal with whitespace",
			args: args{
				s: strings.NewReader("#61 6 26 3 #"),
			},
			wantN: &Node{
				Kind:        KindHexadecimal,
				OctetString: []byte("abc"),
				List:        nil,
			},
			wantErr: false,
		},
		{
			name: "xfail: hexadecimal with newline",
			args: args{
				s: strings.NewReader("#61\n62 63 #"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xfail: hexadecimal without termination",
			args: args{
				s: strings.NewReader("#61|YyWj"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xfail: hexadecimal but eof",
			args: args{
				s: strings.NewReader("#61"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xpass: hexadecimal with length prefix",
			args: args{
				s: strings.NewReader("3#616263#"),
			},
			wantN: &Node{
				Kind:        KindHexadecimal,
				OctetString: []byte("abc"),
				List:        nil,
			},
			wantErr: false,
		},
		{
			name: "xfail: hexadecimal with wrong length prefix",
			args: args{
				s: strings.NewReader("4#616263#"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "xpass: list of two tokens with embedded lists",
			args: args{
				s: strings.NewReader("(abc (def ghi z/a))"),
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
						Kind:        KindList,
						OctetString: nil,
						List: []*Node{
							{
								Kind:        KindToken,
								OctetString: []byte("def"),
								List:        nil,
							},
							{
								Kind:        KindToken,
								OctetString: []byte("ghi"),
								List:        nil,
							},
							{
								Kind:        KindToken,
								OctetString: []byte("z/a"),
								List:        nil,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "xpass: list of empty lists",
			args: args{
				s: strings.NewReader("(() () () ( () () ))"),
			},
			wantN: &Node{
				Kind:        KindList,
				OctetString: nil,
				List: []*Node{
					{
						Kind:        KindList,
						OctetString: nil,
						List:        []*Node{},
					},
					{
						Kind:        KindList,
						OctetString: nil,
						List:        []*Node{},
					},
					{
						Kind:        KindList,
						OctetString: nil,
						List:        []*Node{},
					},
					{
						Kind:        KindList,
						OctetString: nil,
						List: []*Node{
							{
								Kind:        KindList,
								OctetString: nil,
								List:        []*Node{},
							},
							{
								Kind:        KindList,
								OctetString: nil,
								List:        []*Node{},
							},
						},
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

func TestNode_String(t *testing.T) {
	tests := []struct {
		name   string
		fields *Node
		want   string
	}{
		{
			name:   "()",
			fields: MustList(),
			want:   "()",
		},
		{
			name: "(abc)",
			fields: MustList(
				MustToken("abc"),
			),
			want: "(abc)",
		},
		{
			name: "(abc def)",
			fields: MustList(
				MustToken("abc"),
				MustToken("def"),
			),
			want: "(abc def)",
		},
		{
			name: "(abc def (g z/a *))",
			fields: MustList(
				MustToken("abc"),
				MustToken("def"),
				MustList(
					MustToken("g"),
					MustToken("z/a"),
					MustToken("*"),
				),
			),
			want: "(abc def (g z/a *))",
		},
		{
			name: "(abc #616263#)",
			fields: MustList(
				MustToken("abc"),
				MustHexadecimal([]byte("abc")),
			),
			want: "(abc #616263#)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := tt.fields
			if got := n.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMulti(t *testing.T) {
	type args struct {
		s io.RuneScanner
	}
	tests := []struct {
		name    string
		args    args
		wantN   []*Node
		wantErr bool
	}{
		{
			name: "xpass: consecutive lists",
			args: args{
				s: strings.NewReader("(a b)(c d)"),
			},
			wantN: []*Node{
				MustList(
					MustToken("a"),
					MustToken("b"),
				),
				MustList(
					MustToken("c"),
					MustToken("d"),
				),
			},
			wantErr: false,
		},
		{
			name: "xpass: consecutive list and token",
			args: args{
				s: strings.NewReader("(a b)c"),
			},
			wantN: []*Node{
				MustList(
					MustToken("a"),
					MustToken("b"),
				),
				MustToken("c"),
			},
			wantErr: false,
		},
		{
			name: "xpass: consecutive list and token",
			args: args{
				s: strings.NewReader("(a b)c d#6162#"),
			},
			wantN: []*Node{
				MustList(
					MustToken("a"),
					MustToken("b"),
				),
				MustToken("c"),
				MustToken("d"),
				MustHexadecimal([]byte("ab")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN := make([]*Node, len(tt.wantN))
			for i := range tt.wantN {
				var err error
				gotN[i], err = Parse(tt.args.s)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if !reflect.DeepEqual(gotN, tt.wantN) {
				t.Errorf("Parse() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
