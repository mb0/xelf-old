package lit

import (
	"reflect"
	"testing"

	"github.com/mb0/xelf/typ"
)

func TestRead(t *testing.T) {
	tests := []struct {
		Lit
		str, out, jsn string
	}{
		{Null(typ.Any), `null`, ``, ``},
		{Bool(true), `true`, ``, ``},
		{Bool(false), `false`, ``, ``},
		{Num(0), `0`, ``, ``},
		{Num(23), `23`, ``, ``},
		{Num(-23), `-23`, ``, ``},
		{Num(0), `0.0`, `0`, `0`},
		{Num(-0.2), `-0.2`, ``, ``},
		{Char("test"), `"test"`, `'test'`, ``},
		{Char("test"), `'test'`, ``, `"test"`},
		{Char("te\"st"), `'te"st'`, ``, `"te\"st"`},
		{Char("te\"st"), `"te\"st"`, `'te"st'`, ``},
		{Char("te'st"), `'te\'st'`, ``, `"te'st"`},
		{Char("te'st"), `"te'st"`, `'te\'st'`, ``},
		{Char("te\\n\\\"st"), "`" + `te\n\"st` + "`", `'te\\n\\"st'`, `"te\\n\\\"st"`},
		{Char("2019-01-17"), `'2019-01-17'`, ``, `"2019-01-17"`},
		{Idxr{Num(1), Num(2), Num(3)}, `[1,2,3]`, `[1 2 3]`, ``},
		{Idxr{Num(1), Num(2), Num(3)}, `[1,2,3,]`, `[1 2 3]`, `[1,2,3]`},
		{Idxr{Num(1), Num(2), Num(3)}, `[1 2 3]`, ``, `[1,2,3]`},
		{&Keyr{List: []Keyed{{"a", Num(1)}, {"b", Num(2)}, {"c", Num(3)}}},
			`{"a":1,"b":2,"c":3}`,
			`{a:1 b:2 c:3}`, ``,
		},
		{&Keyr{List: []Keyed{{"a", Num(1)}, {"b", Num(2)}, {"c", Num(3)}}},
			`{"a":1,"b":2,"c":3,}`,
			`{a:1 b:2 c:3}`,
			`{"a":1,"b":2,"c":3}`,
		},
		{&Keyr{List: []Keyed{{"a", Num(1)}, {"b", Num(2)}, {"c", Num(3)}}},
			`{"a":1 "b":2 "c":3}`,
			`{a:1 b:2 c:3}`,
			`{"a":1,"b":2,"c":3}`,
		},
		{&Keyr{List: []Keyed{{"a", Num(1)}, {"b c", Num(2)}, {"+foo", Char("bar")}}},
			`{a:1, "b c":2 '+foo':'bar'}`,
			`{a:1 'b c':2 '+foo':'bar'}`,
			`{"a":1,"b c":2,"+foo":"bar"}`,
		},
	}
	for _, test := range tests {
		l, err := ParseString(test.str)
		if err != nil {
			t.Errorf("read %s err %v", test.str, err)
			continue
		}
		if !reflect.DeepEqual(test.Lit, l) {
			t.Errorf("%s want %+v got %+v", test.str, test.Lit, l)
		}
		want := strOr(test.out, test.str)
		if got := l.String(); want != got {
			t.Errorf("want xelf %s got %s", want, got)
		}
		buf, err := l.MarshalJSON()
		if err != nil {
			t.Errorf("marshal %s err %v", test.str, err)
			continue
		}
		want = strOr(test.jsn, test.str)
		if got := string(buf); want != got {
			t.Errorf("want xelf %s got %s", want, got)
		}
	}
}

func strOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
