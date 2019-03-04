package typ_test

import (
	"encoding/json"
	"testing"

	"github.com/mb0/xelf/exp"
	. "github.com/mb0/xelf/typ"
)

func TestString(t *testing.T) {
	tests := []struct {
		typ Type
		raw string
	}{
		{Int, `int`},
		{Opt(Str), `str?`},
		{Ref("a"), `@a`},
		{Opt(Ref("b")), `@b?`},
		{Opt(Enum("kind")), `(enum? 'kind')`},
		{Opt(Obj([]Param{
			{Name: "Name", Type: Str},
		})), `(obj? +Name str)`},
		{Obj([]Param{
			{Name: "x", Type: Int},
			{Name: "y", Type: Int},
		}), `(obj +x +y int)`},
		{Obj([]Param{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `(obj + @Other +Name str)`},
		{Rec("Foo"), `(rec 'Foo')`},
		{Type{Kind: ExpFunc, Info: &Info{Params: []Param{
			{Name: "text", Type: Str},
			{Name: "sub", Type: Str},
			{Type: Int},
		}}}, `(func +text +sub str + int)`},
		{Type{Kind: ExpForm, Info: &Info{Ref: "_", Params: []Param{
			{Name: "a"},
			{Name: "b"},
			{Type: Void},
		}}}, `(form '_' +a +b + void)`},
	}
	for _, test := range tests {
		raw := test.typ.String()
		if got := string(raw); got != test.raw {
			t.Errorf("%s string got %v", test.raw, got)
		}
		typ, err := exp.ParseTypeString(raw)
		if err != nil {
			t.Errorf("%s parse error: %v", test.raw, err)
			continue
		}
		if !typ.Equal(test.typ) {
			t.Errorf("%s parse want %v got %v", test.raw, test.typ, typ)
		}
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		typ Type
		raw string
	}{
		{Int, `{"typ":"int"}`},
		{Opt(Str), `{"typ":"str?"}`},
		{Ref("a"), `{"typ":"ref","ref":"a"}`},
		{Opt(Ref("b")), `{"typ":"ref?","ref":"b"}`},
		{Opt(Enum("kind")), `{"typ":"enum?","ref":"kind"}`},
		{Opt(Obj([]Param{
			{Name: "Name", Type: Str},
		})), `{"typ":"obj?","params":[{"name":"Name","typ":"str"}]}`},
		{Obj([]Param{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `{"typ":"obj","params":[{"typ":"ref","ref":"Other"},{"name":"Name","typ":"str"}]}`},
		{Rec("Foo"), `{"typ":"rec","ref":"Foo"}`},
	}
	for _, test := range tests {
		raw, err := json.Marshal(test.typ)
		if err != nil {
			t.Errorf("%s marshal error: %v", test.raw, err)
			continue
		}
		if got := string(raw); got != test.raw {
			t.Errorf("%s marshal got %v", test.raw, got)
		}
		var typ Type
		err = json.Unmarshal([]byte(test.raw), &typ)
		if err != nil {
			t.Errorf("%s unmarshal error: %v", test.raw, err)
			continue
		}
		if !typ.Equal(test.typ) {
			t.Errorf("%s unmarshal want %v got %v", test.raw, test.typ, typ)
		}
	}
}

func TestTypeSelfRef(t *testing.T) {
	a := Obj([]Param{{Name: "Ref"}})
	a.Params[0].Type = Opt(a)
	b := Obj([]Param{{Name: "C"}})
	c := Obj([]Param{{Name: "Ref"}})
	b.Params[0].Type = c
	c.Params[0].Type = Arr(b)
	tests := []struct {
		typ Type
		raw string
	}{
		{a, "(obj +Ref @0?)"},
		{Opt(a), "(obj? +Ref @0?)"},
		{b, "(obj +C (obj +Ref arr|@1))"},
		{Opt(b), "(obj? +C (obj +Ref arr|@1))"},
	}
	for _, test := range tests {
		raw := test.typ.String()
		if got := string(raw); got != test.raw {
			t.Errorf("%s string got %v", test.raw, got)
		}
		typ, err := exp.ParseTypeString(raw)
		if err != nil {
			t.Errorf("%s parse error: %v", test.raw, err)
			continue
		}
		if !typ.Equal(test.typ) {
			t.Errorf("%s parse want %+v got %+v", test.raw, test.typ, typ)
		}
	}
}
