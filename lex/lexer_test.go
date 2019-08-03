package lex

import (
	"reflect"
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		raw  string
		want *Tree
		err  string
	}{
		{"0.12", &Tree{Token{Tok: Number, Src: src(0, 4), Raw: "0.12"}, nil}, ""},
		{"[0 0]", &Tree{Token{Tok: '[', Src: src(0, 5)}, []*Tree{
			{Token{Tok: Number, Src: src(1, 1), Raw: "0"}, nil},
			{Token{Tok: Number, Src: src(3, 1), Raw: "0"}, nil},
		}}, ""},
		{"[00]", nil, "number zero must be separated by whitespace"},
	}
	for _, test := range tests {
		got, err := Read(strings.NewReader(test.raw))
		if test.err != "" {
			if err == nil {
				t.Errorf("expect error %s got nil", test.err)
				continue
			}
			if !strings.Contains(err.Error(), test.err) {
				t.Errorf("want error %s got %v", test.err, err)
			}
		} else {
			if err != nil {
				t.Errorf("scan %s: %v", test.raw, err)
				continue
			}
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want tree %s got %s\n\t%#[1]v\n\t%#[2]v", test.want, got)
			}
		}
	}
}

func src(o, l uint32) Src {
	return Src{
		Pos{Off: o, Line: 1, Col: uint16(o)},
		Pos{Off: o + l, Line: 1, Col: uint16(o + l)},
	}
}
