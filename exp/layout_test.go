package exp

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mb0/xelf/typ"
)

func TestLayout(t *testing.T) {
	tests := []struct {
		sig  string
		raw  string
		args []string
	}{
		{"<form _ plain; any>",
			"(_ a b c)", []string{
				"(a b c)",
			},
		},
		{"<form _ a; plain; any>",
			"(_ a b c)", []string{
				"(a)",
				"(b c)",
			},
		},
		{"<form _ a; tags; any>",
			"(_ a b:1 c:2)", []string{
				"(a)",
				"(b:1 c:2)",
			},
		},
		{"<form _ a?; tags; any>",
			"(_ b:1 c:2)", []string{
				"()",
				"(b:1 c:2)",
			},
		},
		{"<form _ args; any>",
			"(_ a b:1 c:2)", []string{
				"(a b:1 c:2)",
			},
		},
	}
	for _, test := range tests {
		form, err := typ.Read(strings.NewReader(test.sig))
		if err != nil {
			t.Errorf("parse sig %s err: %v", test.sig, err)
			continue
		}
		el, err := Read(strings.NewReader(test.raw))
		if err != nil {
			t.Errorf("parse raw %s err: %v", test.raw, err)
			continue
		}
		l, err := FormLayout(form, el.(*Dyn).Els[1:])
		if err != nil {
			t.Errorf("layout %s %s %s err: %v", test.sig, test.raw, form, err)
			continue
		}
		res := make([]string, 0, len(l.Groups))
		for _, args := range l.Groups {
			res = append(res, (&Dyn{Els: args}).String())
		}
		if !reflect.DeepEqual(res, test.args) {
			t.Errorf("want %s got %s", test.args, res)
		}
	}
}
