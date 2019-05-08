package typ

import "testing"

func TestCommon(t *testing.T) {
	tests := []struct {
		a, b Type
		want Type
	}{
		{Int, Int, Int},
		{Int, Opt(Int), Int},
		{Int, Real, Num},
		{Int, Span, Int},
		{Str, Int, Any},
		{List(Any), List(Int), List(Any)},
		{Cont(Any), List(Any), List(Any)},
		{List(Int), List(Int), List(Int)},
		{List(Int), List(Real), List(Num)},
		{List(Int), List(Real), List(Num)},
		{List(Int), List(Span), List(Int)},
		{Dict(Int), Dict(Span), Dict(Int)},
		{Str, Var(0), Str},
		{Sym, Expr, Expr},
		{Sym, Typ, Expr},
		{Alt(Int), Int, Alt(Int)},
		{Alt(Int), Alt(Int), Alt(Int)},
		{Alt(Int), Alt(Real), Alt(Int, Real)},
		{Alt(Int, Real), Str, Alt(Int, Real, Str)},
	}
	for _, test := range tests {
		got, _, _ := Common(test.a, test.b)
		if !test.want.Equal(got) {
			t.Errorf("for %s,%s want %s got %s", test.a, test.b, test.want, got)
		}
		got, _, _ = Common(test.b, test.a)
		if !test.want.Equal(got) {
			t.Errorf("for %s,%s want %s got %s", test.b, test.a, test.want, got)
		}
	}
}
