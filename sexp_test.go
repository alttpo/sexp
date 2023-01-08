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
			name: "empty list",
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
			name: "mismatched end of list",
			args: args{
				s: strings.NewReader(")"),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "mismatched start of list",
			args: args{
				s: strings.NewReader("("),
			},
			wantN:   nil,
			wantErr: true,
		},
		{
			name: "single token",
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
