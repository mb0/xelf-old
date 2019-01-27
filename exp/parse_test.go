package exp

import (
	"reflect"
	"testing"

	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func TestParse(t *testing.T) {
	tests := []struct {
		raw  string
		want El
	}{
		{`void`, nil},
		{`(void any 1 2 3 'things')`, nil},
		{`null`, lit.Nil},
		{`()`, Dyn{}},
		{`1`, lit.Num(1)},
		{`bool`, typ.Bool},
		{`name`, &Ref{Sym: Sym{Name: "name"}}},
		{`(false)`, lit.False},
		{`(bool 1)`, Dyn{typ.Bool, lit.Num(1)}},
		{`(obj +x +y int)`, typ.Obj([]typ.Field{
			{Name: "x", Type: typ.Int},
			{Name: "y", Type: typ.Int},
		})},
		{`('Hello ' $Name '!')`, Dyn{
			lit.Char("Hello "),
			&Ref{Sym: Sym{Name: "$Name"}},
			lit.Char("!"),
		}},
		{`(a :b +c d)`, &Expr{
			Sym: Sym{Name: "a"},
			Args: []El{
				Tag{Name: ":b"},
				Decl{Name: "+c", Args: []El{
					&Ref{Sym: Sym{Name: "d"}},
				}},
			},
		}},
		{`((1 2) 1 2)`, Dyn{
			Dyn{lit.Num(1), lit.Num(2)},
			lit.Num(1), lit.Num(2),
		}},
	}
	for _, test := range tests {
		got, err := ParseString(test.raw)
		if err != nil {
			t.Errorf("%s parse err: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s want:\n%#v\n\tgot:\n%#v", test.raw, test.want, got)
			continue
		}
	}
}
