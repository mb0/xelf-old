package typ

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		typ Type
		raw string
		res string
	}{
		{Num, `num`, `~num`},
		{Int, `int`, ``},
		{Opt(Str), `str?`, ``},
		{Ref("a"), `@a`, ``},
		{Ref("a.b"), `@a.b`, ``},
		{Var(1), `@1`, ``},
		{Var(1, Num), `@1:num`, ``},
		{Var(0, Num, Str), `(@:alt num str)`, ``},
		{List(Var(1, Num, Str)), `(list|@1:alt num str)`, ``},
		{Alt(Num, Str), `(alt num str)`, `(~alt num str)`},
		{Opt(Ref("b")), `@b?`, ``},
		{Opt(Sch("a.b")), `~a.b?`, ``},
		{Opt(Enum("kind")), `(enum? 'kind')`, ``},
		{List(Any), `list`, ``},
		{List(Int), `list|int`, ``},
		{Keyr(Num), `keyr|num`, ``},
		{Cont(Num), `cont|num`, ``},
		{Opt(Rec([]Param{
			{Name: "Name", Type: Str},
		})), `(rec? :Name str)`, ``},
		{List(Rec([]Param{
			{Name: "Name", Type: Str},
		})), `(list|rec :Name str)`, ``},
		{Rec([]Param{
			{Name: "x", Type: Int},
			{Name: "y", Type: Int},
		}), `(rec :x :y int)`, ``},
		{Rec([]Param{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `(rec @Other :Name str)`, ``},
		{Obj("Foo"), `(obj 'Foo')`, ``},
		{Type{Kind: KindFunc, Info: &Info{Params: []Param{
			{Name: "text", Type: Str},
			{Name: "sub", Type: Str},
			{Type: Int},
		}}}, `(func :text :sub str int)`, ``},
		{Type{Kind: KindForm, Info: &Info{Ref: "_", Params: []Param{
			{Name: "a"},
			{Name: "b"},
			{Type: Void},
		}}}, `(form '_' :a :b : void)`, ``},
	}
	for _, test := range tests {
		raw := test.typ.String()
		want := test.raw
		if test.res != "" {
			want = test.res
		}
		if got := string(raw); got != want {
			t.Errorf("%s string want %s got %s", test.raw, want, got)
		}
		typ, err := ParseString(raw)
		if err != nil {
			t.Errorf("%s parse error: %v", test.raw, err)
			continue
		}
		if !typ.Equal(test.typ) {
			t.Errorf("%s parse want %v got %v", raw, test.typ, typ)
		}
		rawb, err := json.Marshal(test.typ)
		if err != nil {
			t.Errorf("%s marshal error: %v", test.raw, err)
			continue
		}
		want = fmt.Sprintf(`{"typ":"%s"}`, test.raw)
		if got := string(rawb); got != want {
			t.Errorf("%s marshal got %v", want, got)
		}
		err = json.Unmarshal([]byte(want), &typ)
		if err != nil {
			t.Errorf("%s unmarshal error: %v", want, err)
			continue
		}
		if !typ.Equal(test.typ) {
			t.Errorf("%s unmarshal want %v got %v", want, test.typ, typ)
		}
	}
}

func TestTypeSelfRef(t *testing.T) {
	a := Rec([]Param{{Name: "Ref"}})
	a.Params[0].Type = Opt(a)
	b := Rec([]Param{{Name: "C"}})
	c := Rec([]Param{{Name: "Ref"}})
	b.Params[0].Type = c
	c.Params[0].Type = List(b)
	tests := []struct {
		typ Type
		raw string
	}{
		{a, "(rec :Ref ~0?)"},
		{Opt(a), "(rec? :Ref ~0?)"},
		{b, "(rec :C (rec :Ref list|~1))"},
		{Opt(b), "(rec? :C (rec :Ref list|~1))"},
	}
	for _, test := range tests {
		raw := test.typ.String()
		if got := string(raw); got != test.raw {
			t.Errorf("%s string got %v", test.raw, got)
		}
		typ, err := ParseString(raw)
		if err != nil {
			t.Errorf("%s parse error: %v", test.raw, err)
			continue
		}
		if !typ.Equal(test.typ) {
			t.Errorf("%s parse want %+v got %+v", test.raw, test.typ, typ)
		}
	}
}
