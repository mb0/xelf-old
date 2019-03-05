package utl

import (
	"strings"
	"testing"
	"time"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func TestReflectFunc(t *testing.T) {
	tests := []struct {
		fun   interface{}
		name  string
		names []string
		typ   exp.Sig
		err   bool
	}{
		{strings.ToLower, "lower", nil, exp.FuncSig("lower", []typ.Param{
			{Type: typ.Str},
			{Type: typ.Str},
		}), false},
		{strings.Split, "", nil, exp.AnonSig([]typ.Param{
			{Type: typ.Str},
			{Type: typ.Str},
			{Type: typ.Arr(typ.Str)},
		}), false},
		{time.Parse, "", nil, exp.AnonSig([]typ.Param{
			{Type: typ.Str},
			{Type: typ.Str},
			{Type: typ.Time},
		}), true},
		{time.Time.Format, "", []string{"t", "format"}, exp.AnonSig([]typ.Param{
			{Name: "t", Type: typ.Time},
			{Name: "format", Type: typ.Str},
			{Type: typ.Str},
		}), false},
	}
	for _, test := range tests {
		r, err := ReflectFunc(test.name, test.fun, test.names...)
		if err != nil {
			t.Errorf("reflect for %+v err: %v", test.fun, err)
			continue
		}
		if !test.typ.Equal(r.Sig.Info) {
			t.Errorf("for %T want param %s got %s", test.fun, test.typ, r.Sig)
		}
		b := r.Body.(*ReflectBody)
		if test.err != b.err {
			t.Errorf("for %T want err %v got %v", test.fun, test.err, b.err)
		}
	}
}

func TestFuncResolver(t *testing.T) {
	tests := []struct {
		fun   interface{}
		names []string
		args  []exp.El
		want  string
		err   error
	}{
		{strings.ToLower, nil, []exp.El{
			lit.Str("HELLO"),
		}, `'hello'`, nil},
		{time.Time.Format, []string{"t", "format"}, []exp.El{
			exp.Tag{Name: ":format", Args: []exp.El{
				lit.Char(`2006-02-01`),
			}},
		}, `'0001-01-01'`, nil},
	}
	for _, test := range tests {
		r, err := ReflectFunc("", test.fun, test.names...)
		if err != nil {
			t.Errorf("reflect for %+v err: %v", test.fun, err)
			continue
		}
		c := &exp.Ctx{Exec: true}
		res, err := r.Resolve(c, exp.StdEnv, &exp.Expr{r, test.args}, typ.Void)
		if err != nil {
			if test.err == nil || test.err.Error() != err.Error() {
				t.Errorf("for %T want err %v got %v", test.fun, test.err, err)
			}
			continue
		}
		if test.want != "" {
			if got := res.String(); test.want != got {
				t.Errorf("for %T want res %s got %v", test.fun, test.want, got)

			}
		}
	}
}
