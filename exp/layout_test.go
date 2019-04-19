package exp

import (
	"reflect"
	"testing"

	"github.com/mb0/xelf/typ"
)

func TestLayout(t *testing.T) {
	tests := []struct {
		sig  string
		raw  string
		args []string
	}{
		{"(form '_' +plain + any)",
			"(_ a b c)", []string{
				"(a b c)",
			},
		},
		{"(form '_' +a +plain + any)",
			"(_ a b c)", []string{
				"(a)",
				"(b c)",
			},
		},
		{"(form '_' +a +tags + any)",
			"(_ a :b 1 :c 2)", []string{
				"(a)",
				"(:b 1 :c 2)",
			},
		},
		{"(form '_' +a? +tags + any)",
			"(_ :b 1 :c 2)", []string{
				"()",
				"(:b 1 :c 2)",
			},
		},
		{"(form '_' +args + any)",
			"(_ a :b 1 :c 2)", []string{
				"(a :b 1 :c 2)",
			},
		},
		{"(form '_' +decls + any)",
			"(_ +a +b 1 +c 2)", []string{
				"(+a +b 1 +c 2)",
			},
		},
		{"(form '_' +decls + any)",
			"(_ (+a +b 1) +c 2)", []string{
				"((+a +b 1) +c 2)",
			},
		},
	}
	for _, test := range tests {
		form, err := typ.ParseString(test.sig)
		if err != nil {
			t.Errorf("parse sig %s err: %v", test.sig, err)
			continue
		}
		ps := form.Params[:len(form.Params)-1]
		el, err := ParseString(test.raw)
		if err != nil {
			t.Errorf("parse raw %s err: %v", test.raw, err)
			continue
		}
		l, err := LayoutArgs(ps, el.(Dyn)[1:])
		if err != nil {
			t.Errorf("layout %s %s %s err: %v", test.sig, test.raw, ps, err)
			continue
		}
		res := make([]string, 0, len(l.args))
		for _, args := range l.args {
			res = append(res, Dyn(args).String())
		}
		if !reflect.DeepEqual(res, test.args) {
			t.Errorf("want %s got %s", test.args, res)
		}
	}
}
