package typ

import (
	"encoding/json"
	"testing"
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
		{Opt(Obj([]Field{
			{Name: "Name", Type: Str},
		})), `(obj? +Name str)`},
		{Obj([]Field{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `(obj + @Other +Name str)`},
		{Rec("Foo", []Field{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `(rec 'Foo' + @Other +Name str)`},
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
		{Opt(Obj([]Field{
			{Name: "Name", Type: Str},
		})), `{"typ":"obj?","fields":[{"name":"Name","typ":"str"}]}`},
		{Obj([]Field{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `{"typ":"obj","fields":[{"typ":"ref","ref":"Other"},{"name":"Name","typ":"str"}]}`},
		{Rec("Foo", []Field{
			{Type: Ref("Other")},
			{Name: "Name", Type: Str},
		}), `{"typ":"rec","ref":"Foo","fields":[{"typ":"ref","ref":"Other"},{"name":"Name","typ":"str"}]}`},
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
