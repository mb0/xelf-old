package typ

import (
	"testing"
)

func TestChoose(t *testing.T) {
	tests := []struct {
		alt  Type
		want Type
	}{
		{Alt(Int), Int},
		{Alt(Int, Span), Int},
		{Alt(Int, Real), Num},
		{Alt(Int, Real, Span), Num},
		{Alt(Num, Int), Int},
		{Alt(Any, Num, Int), Any},
		{Alt(Num, Int, Span), Num},
		{Alt(Num, Int, Real, Span), Num},
		{Alt(Span, Real), Num},
		{Alt(Str, Int), Any},
		// TODO what do we want for parameter types?
		{Alt(List(Int), List(Num)), List(Num)},
		{Alt(List(Int), List(Real)), List(Num)},
		{Alt(List(Any), List(Num)), List(Any)},
		{Alt(List(Num), List(Int)), List(Num)},
		{Alt(List(Str), List(Int)), List(Any)},
	}
	for _, test := range tests {
		got, _ := Choose(test.alt)
		if !test.want.Equal(got) {
			t.Errorf("for %s want %s got %s %#v", test.alt, test.want, got, got.Info)
		}
	}
}
