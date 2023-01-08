package sexp

import (
	"io"
	"reflect"
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
		// TODO: Add test cases.
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
