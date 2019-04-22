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
		{Int, Var(1), Int},
		{Var(1), Var(2), Var(2)},
		{Var(1), Var(1), Var(1)},
		{Alt(Num, Int), Int, Int},
		{Alt(Num, Int), Real, Num},
		{Int, Alt(Num, Int), Int},
		{Real, Alt(Num, Int), Num},
		{Arr(Int), List, Arr(Int)},
		{List, Arr(Int), Arr(Int)},
		{Arr(Int), Arr(Int), Arr(Int)},
		{Arr(Real), Arr(Int), Arr(Num)},
		{Alt(Char, Str, Raw), UUID, Char},
		{Alt(Char, Str), Time, Char},
		{Alt(Num, Int), Time, Num},
	}
	for _, test := range tests {
		c := new(Ctx)
		a, m := c.inst(test.a, nil)
		b, m := c.inst(test.b, m)
		w, m := c.inst(test.w, m)
		r := c.New()
		err := Unify(c, r, a)
		if err != nil {
			t.Errorf("unify error: %v", err)
			continue
		}
		err = Unify(c, r, b)
		if err != nil {
			t.Errorf("unify error: %v", err)
			continue
		}

		got := c.Apply(r)
		if !got.Equal(w) {
			t.Errorf("unify %s for %s %s want %s got %s\n%v",
				r, a, b, w, got, c.binds)
		}
	}
}
func TestUnifyError(t *testing.T) {
	tests := []struct {
		a, b Type
	}{
		{Num, Char},
		{Alt(Num, Int), Char},
		{Arr(Alt(Num)), Arr(Char)},
	}
	for _, test := range tests {
		c := new(Ctx)
		a, m := c.inst(test.a, nil)
		b, m := c.inst(test.b, m)
		r := c.New()
		err := Unify(c, r, a)
		if err != nil {
			t.Errorf("unify a error for %s %s: %+v", a, b, err)
			continue
		}
		err = Unify(c, r, b)
		if err == nil {
			got := c.Apply(r)
			t.Errorf("unify b want error for %s %s got %s", a, b, got)
		}
	}
}
