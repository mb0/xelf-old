package lit

import (
	"reflect"
	"testing"
)

func TestSetPath(t *testing.T) {
	tests := []struct {
		lit  Lit
		path string
		el   Lit
		res  Lit
	}{
		{&Dict{}, "a.1.b", True, &Dict{List: []Keyed{
			{"a", &List{Data: []Lit{Nil, &Dict{List: []Keyed{
				{"b", True},
			}}}}},
		}}},
	}
	for _, test := range tests {
		p, err := ReadPath(test.path)
		if err != nil {
			t.Errorf("parse path %s error: %v", test.path, err)
			continue
		}
		l, err := SetPath(test.lit, p, test.el, true)
		if err != nil {
			t.Errorf("set path %s error: %v", test.path, err)
			continue
		}
		if !reflect.DeepEqual(test.res, l) {
			t.Errorf("want %s got %s", test.res, l)
		}
	}
}
