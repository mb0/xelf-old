package exp_test

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func pos(o int) lex.Pos    { return lex.Pos{uint32(o), 1, uint16(o)} }
func src(p, e int) lex.Src { return lex.Src{Pos: pos(p), End: pos(e)} }

func TestParse(t *testing.T) {
	tests := []struct {
		raw  string
		want El
	}{
		{`void`, &Atom{typ.Void, src(0, 4)}},
		{`(void 1 2 3 'things')`, nil},
		{`()`, nil},
		{`null`, &Atom{lit.Nil, src(0, 4)}},
		{`1`, &Atom{lit.Num(1), src(0, 1)}},
		{`bool`, &Atom{typ.Bool, src(0, 4)}},
		{`name`, &Sym{Name: "name", Src: src(0, 4)}},
		{`(false)`, &Dyn{Src: src(0, 7), Els: []El{
			&Atom{lit.False, src(1, 6)}},
		}},
		{`(int 1)`, &Dyn{Src: src(0, 7), Els: []El{
			&Atom{typ.Int, src(1, 4)},
			&Atom{lit.Num(1), src(5, 6)},
		}}},
		{`(bool 1)`, &Dyn{Src: src(0, 8), Els: []El{
			&Atom{typ.Bool, src(1, 5)},
			&Atom{lit.Num(1), src(6, 7)},
		}}},
		{`(bool (() comment) 1)`, &Dyn{Src: src(0, 21), Els: []El{
			&Atom{typ.Bool, src(1, 5)},
			&Atom{lit.Num(1), src(19, 20)},
		}}},
		{`(rec :x :y int)`, &Atom{typ.Rec([]typ.Param{
			{"x", typ.Int},
			{"y", typ.Int},
		}), src(0, 15)}},
		{`('Hello ' $Name '!')`, &Dyn{Src: src(0, 20), Els: []El{
			&Atom{lit.Char("Hello "), src(1, 9)},
			&Sym{Name: "$Name", Src: src(10, 15)},
			&Atom{lit.Char("!"), src(16, 19)},
		}}},
		{`(a :b +c d)`, &Dyn{Src: src(0, 11), Els: []El{
			&Sym{Name: "a", Src: src(1, 2)},
			&Named{Name: ":b", Src: src(3, 5)},
			&Named{Name: "+c", Src: src(6, 8)},
			&Sym{Name: "d", Src: src(9, 10)},
		}}},
		{`((1 2) 1 2)`, &Dyn{Src: src(0, 11), Els: []El{
			&Dyn{Src: src(1, 6), Els: []El{
				&Atom{lit.Num(1), src(2, 3)},
				&Atom{lit.Num(2), src(4, 5)},
			}},
			&Atom{lit.Num(1), src(7, 8)},
			&Atom{lit.Num(2), src(9, 10)},
		}}},
		{`(1 (+z 3 4))`, &Dyn{Src: src(0, 12), Els: []El{
			&Atom{lit.Num(1), src(1, 2)},
			&Named{Name: "+z", Src: src(3, 11), El: &Dyn{Src: src(7, 10), Els: []El{
				&Atom{lit.Num(3), src(7, 8)},
				&Atom{lit.Num(4), src(9, 10)},
			}}},
		}}},
		{`(s (+m +a u :t))`, &Dyn{Src: src(0, 16), Els: []El{
			&Sym{Name: "s", Src: src(1, 2)},
			&Named{Name: "+m", Src: src(3, 15), El: &Dyn{Src: src(7, 14), Els: []El{
				&Named{Name: "+a", Src: src(7, 9)},
				&Sym{Name: "u", Src: src(10, 11)},
				&Named{Name: ":t", Src: src(12, 14)},
			}}},
		}}},
	}
	for _, test := range tests {
		got, err := Read(strings.NewReader(test.raw))
		if err != nil && err != ErrVoid {
			t.Errorf("%s parse err: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s want:\n%s\n\tgot:\n%s", test.raw, test.want, got)
			continue
		}
	}
}
