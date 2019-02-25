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
		names []string
		typ   typ.Type
		err   bool
	}{
		{strings.ToLower, nil, typ.Func([]typ.Field{
			{Type: typ.Str},
		}, typ.Str), false},
		{strings.Split, nil, typ.Func([]typ.Field{
			{Type: typ.Str},
			{Type: typ.Str},
		}, typ.Arr(typ.Str)), false},
		{time.Parse, nil, typ.Func([]typ.Field{
			{Type: typ.Str},
			{Type: typ.Str},
		}, typ.Time), true},
		{time.Time.Format, []string{"t", "format"}, typ.Func([]typ.Field{
			{Name: "t", Type: typ.Time},
			{Name: "format", Type: typ.Str},
		}, typ.Str), false},
	}
	for _, test := range tests {
		r, err := ReflectFunc(test.fun, test.names...)
		if err != nil {
			t.Errorf("reflect for %+v err: %v", test.fun, err)
			continue
		}
		if !test.typ.Equal(r.Sig) {
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
		r, err := ReflectFunc(test.fun, test.names...)
		if err != nil {
			t.Errorf("reflect for %+v err: %v", test.fun, err)
			continue
		}
		res, err := r.Resolve(&exp.Ctx{}, exp.StdEnv, &exp.Expr{Args: test.args})
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
