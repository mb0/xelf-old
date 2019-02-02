package gen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mb0/xelf/typ"
)

func TestWriteGoType(t *testing.T) {
	tests := []struct {
		t       typ.Type
		s       string
		imports []string
	}{
		{typ.Any, "interface{}", nil},
		{typ.List, "[]interface{}", nil},
		{typ.Dict, "map[string]interface{}", nil},
		{typ.Bool, "bool", nil},
		{typ.Span, "time.Duration", []string{"time"}},
		{typ.Arr(typ.Time), "[]time.Time", []string{"time"}},
		{typ.Obj([]typ.Field{
			{Name: "Foo", Type: typ.Str},
			{Name: "Bar?", Type: typ.Int},
			{Name: "Spam", Type: typ.Opt(typ.Int)},
		}), `struct {
	Foo string ` + "`json:\"foo\"`" + `
	Bar int64 ` + "`json:\"bar,omitempty\"`" + `
	Spam *int64 ` + "`json:\"spam\"`\n}", nil},
	}
	for _, test := range tests {
		var b strings.Builder
		c := &Ctx{B: &b}
		err := WriteGoType(c, test.t)
		if err != nil {
			t.Errorf("test %s error: %v", test.s, err)
			continue
		}
		res := b.String()
		if res != test.s {
			t.Errorf("test %s got %s", test.s, res)
		}
		if !reflect.DeepEqual(c.Imports.List, test.imports) {
			t.Errorf("test %s want imports %v got %v", test.s, test.imports, c.Imports)
		}
	}
}
