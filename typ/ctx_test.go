package typ_test

import (
	"testing"

	"github.com/mb0/xelf/typ"
)

func TestCtx(t *testing.T) {
	s := typ.Func("", []typ.Param{
		{Type: typ.Func("", []typ.Param{
			{Type: typ.Var(1)},
			{Type: typ.Bool},
		})},
		{Type: typ.List(typ.Var(1))},
		{Type: typ.List(typ.Var(1))},
	})
	want := `(func (func @1 bool) list|@1 list|@1)`
	if got := s.String(); got != want {
		t.Errorf("want %s\ngot %s", want, got)
	}
	c := new(typ.Ctx)
	c.Bind(typ.VarKind(5), typ.Any)
	s = c.Inst(s)
	want = `(func (func @6 bool) list|@6 list|@6)`
	if got := s.String(); got != want {
		t.Errorf("want inst %s\ngot %s", want, got)
	}
	free := c.Free(s, nil)
	if len(free) != 1 || free[0] != typ.VarKind(6) {
		t.Errorf("want free [6] got %s", free)
	}
	bound := c.Bound(s, nil)
	if len(bound) != 0 {
		t.Errorf("want bound [] got %s", bound)
	}
	c.Bind(typ.VarKind(6), typ.Int)
	free = c.Free(s, nil)
	if len(free) != 0 {
		t.Errorf("want free [] got %s", free)
	}
	bound = c.Bound(s, nil)
	if len(bound) != 1 || bound[0] != typ.VarKind(6) {
		t.Errorf("want bound [6] got %s", bound)
	}
	a := c.Apply(s)
	want = `(func (func int bool) list|int list|int)`
	if got := a.String(); got != want {
		t.Errorf("want %s\ngot %s", want, got)
	}
}
