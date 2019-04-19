package exp

import (
	"reflect"
	"testing"

	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func TestParse(t *testing.T) {
	tests := []struct {
		raw  string
		want El
	}{
		{`void`, typ.Void},
		{`(void 1 2 3 'things')`, typ.Void},
		{`()`, typ.Void},
		{`null`, lit.Nil},
		{`1`, lit.Num(1)},
		{`bool`, typ.Bool},
		{`name`, &Sym{Name: "name", Pos: lex.Offset(0)}},
		{`(false)`, Dyn{lit.False}},
		{`(int 1)`, Dyn{typ.Int, lit.Num(1)}},
		{`(bool 1)`, Dyn{typ.Bool, lit.Num(1)}},
		{`(bool (() comment) 1)`, Dyn{typ.Bool, lit.Num(1)}},
		{`(obj +x +y int)`, typ.Obj([]typ.Param{
			{"x", typ.Int},
			{"y", typ.Int},
		})},
		{`('Hello ' $Name '!')`, Dyn{
			lit.Char("Hello "),
			&Sym{Name: "$Name", Pos: lex.Offset(10)},
			lit.Char("!"),
		}},
		{`(a :b +c d)`, Dyn{
			&Sym{Name: "a", Pos: lex.Offset(1)},
			&Named{Name: ":b", Pos: lex.Offset(3)},
			&Named{Name: "+c", Pos: lex.Offset(6)},
			&Sym{Name: "d", Pos: lex.Offset(9)},
		}},
		{`((1 2) 1 2)`, Dyn{
			Dyn{lit.Num(1), lit.Num(2)},
			lit.Num(1), lit.Num(2),
		}},
		{`(1 (+z 3 4))`, Dyn{lit.Num(1),
			&Named{Name: "+z", Pos: lex.Offset(4), El: Dyn{
				lit.Num(3),
				lit.Num(4),
			}},
		}},
		{`(s (+m +a u :t))`, Dyn{
			&Sym{Name: "s", Pos: lex.Offset(1)},
			&Named{Name: "+m", Pos: lex.Offset(4), El: Dyn{
				&Named{Name: "+a", Pos: lex.Offset(7)},
				&Sym{Name: "u", Pos: lex.Offset(10)},
				&Named{Name: ":t", Pos: lex.Offset(12)},
			}},
		}},
	}
	for _, test := range tests {
		got, err := ParseString(test.raw)
		if err != nil {
			t.Errorf("%s parse err: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s want:\n%s\n\tgot:\n%#v", test.raw, test.want, got)
			continue
		}
	}
}
