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
		{`void`, typ.Void},
		{`(void 1 2 3 'things')`, typ.Void},
		{`()`, typ.Void},
		{`null`, lit.Nil},
		{`1`, lit.Num(1)},
		{`bool`, &Sym{Name: "bool"}},
		{`name`, &Sym{Name: "name"}},
		{`(false)`, Dyn{lit.False}},
		{`(int 1)`, Dyn{&Sym{Name: "int"}, lit.Num(1)}},
		{`(bool 1)`, Dyn{&Sym{Name: "bool"}, lit.Num(1)}},
		{`(bool (() comment) 1)`, Dyn{&Sym{Name: "bool"}, lit.Num(1)}},
		{`(obj +x +y int)`, Dyn{&Sym{Name: "obj"},
			Decl{Name: "+x"},
			Decl{Name: "+y", Args: []El{&Sym{Name: "int"}}},
		}},
		{`('Hello ' $Name '!')`, Dyn{
			lit.Char("Hello "),
			&Sym{Name: "$Name"},
			lit.Char("!"),
		}},
		{`(a :b +c d)`, Dyn{
			&Sym{Name: "a"},
			Tag{Name: ":b"},
			Decl{Name: "+c", Args: []El{
				&Sym{Name: "d"},
			}},
		}},
		{`((1 2) 1 2)`, Dyn{
			Dyn{lit.Num(1), lit.Num(2)},
			lit.Num(1), lit.Num(2),
		}},
		{`(1 (+z 3 4))`, Dyn{
			lit.Num(1),
			Decl{Name: "+z", Args: []El{
				lit.Num(3),
				lit.Num(4),
			}},
		}},
		{`(s (+m +a u :t))`, Dyn{
			&Sym{Name: "s"},
			Decl{Name: "+m", Args: []El{
				Decl{Name: "+a", Args: []El{
					&Sym{Name: "u"},
					Tag{Name: ":t"},
				}},
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
			t.Errorf("%s want:\n%s\n\tgot:\n%s", test.raw, test.want, got)
			continue
		}
	}
}
