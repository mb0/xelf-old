package lit

import (
	"reflect"
	"testing"
	"time"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type myUUID [16]byte
type myTime time.Time
type myPoint struct {
	X, Y int
}

func TestAdapt(t *testing.T) {
	tests := []struct {
		val  interface{}
		want Lit
	}{
		{nil, Nil},
		{true, Bool(true)},
		{false, Bool(false)},
		{(*bool)(nil), Null(typ.Bool)},
		{cor.Bool(true), Some{Bool(true)}},
		{cor.Bool(false), Some{Bool(false)}},
		{cor.Any(false), Any{Bool(false)}},
		{1, Int(1)},
		{1.1, Real(1.1)},
		{"test", Str("test")},
		{cor.Str("test"), Some{Str("test")}},
		{[16]byte{}, ZeroUUID},
		{myUUID{}, ZeroUUID},
		{time.Time{}, ZeroTime},
		{myTime{}, ZeroTime},
		{[]int{1, 2}, &abstractArr{typ.Int, List{Int(1), Int(2)}}},
		{[]*int64{cor.Int(1), cor.Int(2)},
			&abstractArr{typ.Opt(typ.Int), List{Some{Int(1)}, Some{Int(2)}}},
		},
		{myPoint{1, 2}, &abstractObj{typ.Obj([]typ.Field{
			{Name: "X", Type: typ.Int},
			{Name: "Y", Type: typ.Int},
		}), Dict{[]Keyed{
			{Key: "x", Lit: Int(1)},
			{Key: "y", Lit: Int(2)},
		}}}},
		{(*myPoint)(nil), Null(typ.Obj([]typ.Field{
			{Name: "X", Type: typ.Int},
			{Name: "Y", Type: typ.Int},
		}))},
	}
	for _, test := range tests {
		got, err := Adapt(test.val)
		if err != nil {
			t.Errorf("adapt val %#v err: %v", test.val, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("adapt val want %#v got %#v", test.want, got)
		}
		got, err = Adapt(got)
		if err != nil {
			t.Errorf("adapt lit %#v err: %v", test.want, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("adapt lit want %#v got %#v", test.want, got)
		}
	}
}
