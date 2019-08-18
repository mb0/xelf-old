package typ

import "testing"

func TestUnify(t *testing.T) {
	tests := []struct {
		a, b, w Type
	}{
		{Int, Int, Int},
		{Num, Int, Int},
		{Int, Num, Int},
		{Num, Num, Num},
		{Int, Real, Num},
		{Var(1), Int, Int},
		{Var(1, Num), Int, Int},
		{Int, Var(1, Num), Int},
		{Int, Var(1, Num, Char), Int},
		{Time, Var(1, Num, Char), Time},
		{List(Var(1, Num)), List(Int), List(Int)},
		{Int, Var(1), Int},
		{Var(1), Var(2), Var(2)},
		{Var(1), Var(2, Num), Var(2, Num)},
		{Var(1, Num), Var(2), Var(2, Num)},
		{Var(1), Var(1), Var(1)},
		{Var(1, Num), Var(1), Var(1, Num)},
		{Var(1), Var(1, Num), Var(1, Num)},
		{Alt(Num), Int, Int},
		{Alt(Any), Int, Int},
		{Alt(Any, Num), Int, Any},
		{Alt(Num, Int), Int, Int},
		{Alt(Num, Int), Real, Num},
		{Alt(Num, Int), Num, Int},
		{Int, Alt(Num, Int), Int},
		{Real, Alt(Num, Int), Num},
		{List(Int), List(Any), List(Any)},
		{List(Any), List(Int), List(Any)},
		{List(Int), List(Int), List(Int)},
		{List(Real), List(Int), List(Num)},
		{Cont(Any), List(Any), List(Any)},
		{Cont(Int), List(Var(1)), List(Int)},
		{Var(1, Cont(Int)), Dict(Var(2)), Dict(Int)},
		{Alt(Char, Str, Raw), UUID, Char},
		{Alt(Char, Str), Time, Char},
	}
	for _, test := range tests {
		c := new(Ctx)
		a, m := c.inst(test.a, nil, nil)
		b, m := c.inst(test.b, m, nil)
		r := c.New()
		_, err := Unify(c, r, a)
		if err != nil {
			t.Errorf("unify error: %v", err)
			continue
		}
		_, err = Unify(c, r, b)
		if err != nil {
			t.Errorf("unify error for %s %s: %v", r, test.b, err)
			continue
		}
		got := c.Apply(r)
		if !got.Equal(test.w) {
			t.Errorf("unify %s for %s %s want %s got %s\n%v",
				r, a, b, test.w, got, c.binds)
		}
		c = new(Ctx)
		a, m = c.inst(test.a, nil, nil)
		b, m = c.inst(test.b, m, nil)
		r, err = Unify(c, a, b)
		if err != nil {
			t.Errorf("unify ab error: %v", err)
			continue
		}
		got = c.Apply(r)
		if !got.Equal(test.w) {
			t.Errorf("unify ab %s for %s %s want %s got %s\n%v",
				r, a, b, test.w, got, c.binds)
		}
	}
}

func TestUnifyError(t *testing.T) {
	tests := []struct {
		a, b Type
	}{
		{Num, Char},
		{Var(1, Char), Int},
		{Int, Var(1, Char)},
		{Alt(Num, Int), Char},
		{List(Alt(Num)), List(Char)},
	}
	for _, test := range tests {
		c := new(Ctx)
		a, m := c.inst(test.a, nil, nil)
		b, m := c.inst(test.b, m, nil)
		r := c.New()
		_, err := Unify(c, r, a)
		if err != nil {
			t.Errorf("unify a error for %s %s: %+v", a, b, err)
			continue
		}
		_, err = Unify(c, r, b)
		if err == nil {
			got := c.Apply(r)
			t.Errorf("unify b want error for %s %s got %s", a, b, got)
		}
	}
}
