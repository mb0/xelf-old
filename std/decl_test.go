package std

import (
	"strings"
	"testing"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/typ"
)

func TestInferFn(t *testing.T) {
	tests := []struct {
		raw string
		sig typ.Type
	}{
		{"(fn (add _ 1))", typ.Func("", []typ.Param{
			{Type: typ.Var(2, typ.Num)},
			{Type: typ.Var(2, typ.Num)},
		})},
	}
	for _, test := range tests {
		x, err := exp.Read(strings.NewReader(test.raw))
		if err != nil {
			t.Errorf("parse %s error: %v", test.raw, err)
			continue
		}
		l, err := exp.Eval(Std, x)
		if err != nil {
			t.Errorf("exec %s error: %v", x, err)
			continue
		}
		s, ok := l.(*exp.Atom).Lit.(*exp.Spec)
		if !ok {
			t.Errorf("for %s want spec got %T %[2]s", x, s)
			continue
		}
		if !test.sig.Equal(s.Type) {
			t.Errorf("for %s want %s got %s", x, test.sig, s.Type)
		}
	}
}
