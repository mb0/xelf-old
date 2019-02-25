package exp

import (
	"reflect"
	"testing"

	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func TestStdFail(t *testing.T) {
	x, err := ParseString(`(fail 'oops')`)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	c := &Ctx{Exec: true}
	_, err = c.Resolve(StdEnv, x)
	if err == nil {
		t.Fatalf("want err got nothing")
	}
	c.Exec = false
	_, err = c.Resolve(StdEnv, x)
	if err != ErrUnres {
		t.Fatalf("want err unres got %v", err)
	}
}
func TestStdResolve(t *testing.T) {
	tests := []struct {
		raw  string
		want El
	}{
		{`any`, typ.Any},
		{`bool`, typ.Bool},
		{`void`, typ.Void},
		{`raw`, typ.Raw},
		{`null`, lit.Nil},
		{`true`, lit.True},
		{`(void anything)`, typ.Void},
		{`(true)`, lit.True},
		{`(bool)`, lit.False},
		{`(bool 1)`, lit.True},
		{`(bool 0)`, lit.False},
		{`(raw)`, lit.Raw(nil)},
		{`7`, lit.Num(7)},
		{`(7)`, lit.Num(7)},
		{`()`, typ.Void},
		{`(int 7)`, lit.Int(7)},
		{`(real 7)`, lit.Real(7)},
		{`'abc'`, lit.Char("abc")},
		{`(str)`, lit.Str("")},
		{`(str 'abc')`, lit.Str("abc")},
		{`(raw 'abc')`, lit.Raw("abc")},
		{`(time)`, lit.ZeroTime},
		{`(time null)`, lit.ZeroTime},
		{`(or)`, lit.False},
		{`(or 0)`, lit.False},
		{`(or 1)`, lit.True},
		{`(or 1 (fail))`, lit.True},
		{`(or 0 1)`, lit.True},
		{`(or 1 2 3)`, lit.True},
		{`(and)`, lit.True},
		{`(and 0)`, lit.False},
		{`(and 1)`, lit.True},
		{`(and 1 0)`, lit.False},
		{`(and 0 (fail))`, lit.False},
		{`(and 1 2 3)`, lit.True},
		{`(true 2 3)`, lit.True},
		{`((bool 1) 2 3)`, lit.True},
		{`(not)`, lit.True},
		{`(not 0)`, lit.True},
		{`(not 1)`, lit.False},
		{`(not 0 (fail))`, lit.True},
		{`(not 1 0)`, lit.True},
		{`(not 0 1)`, lit.True},
		{`(not 1 2 3)`, lit.False},
		{`(eq 1 1)`, lit.True},
		{`(eq (int 1) 1)`, lit.True},
		{`(equal (int 1) 1)`, lit.False},
		{`(equal (int 1) (int 1))`, lit.True},
		{`(ne 1 1)`, lit.False},
		{`(ne 0 1)`, lit.True},
		{`(ne 1 1 1)`, lit.False},
		{`(ne 1 1 2)`, lit.True},
		{`(ne 0 1 2)`, lit.True},
		{`(lt 0 1 2)`, lit.True},
		{`(lt 2 1 0)`, lit.False},
		{`(lt 0 0 2)`, lit.False},
		{`(ge 0 1 2)`, lit.False},
		{`(ge 2 1 0)`, lit.True},
		{`(ge 0 0 2)`, lit.False},
		{`(ge 2 0 0)`, lit.True},
		{`(gt 0 1 2)`, lit.False},
		{`(gt 2 1 0)`, lit.True},
		{`(gt 0 0 2)`, lit.False},
		{`(gt 2 0 0)`, lit.False},
		{`(le 0 1 2)`, lit.True},
		{`(le 2 1 0)`, lit.False},
		{`(le 0 0 2)`, lit.True},
		{`(le 2 0 0)`, lit.False},
		{`(add)`, lit.Num(0)},
		{`(add 1 2)`, lit.Num(3)},
		{`(add 1 2 3)`, lit.Num(6)},
		{`(add -5 2 3)`, lit.Num(0)},
		{`(add (int? 1) 2 3)`, lit.Some{lit.Int(6)}},
		{`(1 2 3)`, lit.Num(6)},
		{`(add (int 1) 2 3)`, lit.Int(6)},
		{`(add (real 1) 2 3)`, lit.Real(6)},
		{`(abs 1)`, lit.Num(1)},
		{`(abs -1)`, lit.Num(1)},
		{`(abs (int -1))`, lit.Int(1)},
		{`(min 1 2 3)`, lit.Num(1)},
		{`(min 3 2 1)`, lit.Num(1)},
		{`(max 1 2 3)`, lit.Num(3)},
		{`(max 3 2 1)`, lit.Num(3)},
		{`(cat 'a' 'b' 'c')`, lit.Str("abc")},
		{`(cat (raw 'a') 'b' 'c')`, lit.Raw("abc")},
		{`(cat [1] [2] [3])`, lit.List{lit.Num(1), lit.Num(2), lit.Num(3)}},
		{`(apd [] 1 2 3)`, lit.List{lit.Num(1), lit.Num(2), lit.Num(3)}},
		{`([] 1 2 3)`, lit.List{lit.Num(1), lit.Num(2), lit.Num(3)}},
		{`(list (arr|int 1 2 3))`, lit.List{lit.Int(1), lit.Int(2), lit.Int(3)}},
		{`(set {} +x +y 3)`, &lit.Dict{List: []lit.Keyed{
			{"x", lit.Num(3)},
			{"y", lit.Num(3)},
		}}},
		{`({} +x +y 3)`, &lit.Dict{List: []lit.Keyed{
			{"x", lit.Num(3)},
			{"y", lit.Num(3)},
		}}},
		{`(dict (map|int +x +y 3))`, &lit.Dict{List: []lit.Keyed{
			{"x", lit.Int(3)},
			{"y", lit.Int(3)},
		}}},
		{`((real 1) 2 3)`, lit.Real(6)},
		{`(if 1 2)`, lit.Num(2)},
		{`(if 1 2 (fail))`, lit.Num(2)},
		{`(if 1 2 (fail) 3)`, lit.Num(2)},
		{`(if 0 1 2 3)`, lit.Num(3)},
		{`(if 0 1 0 2 3)`, lit.Num(3)},
		{`(if 0 1 0 2)`, lit.Num(0)},
		{`(if 0 (fail) 2)`, lit.Num(2)},
		{`(if 0 (fail))`, lit.Nil},
		{`(if 1 'a')`, lit.Char("a")},
		{`(if 0 'a' 'b')`, lit.Char("b")},
		{`(if 0 'a')`, lit.Char("")},
		{`(let +a (int 1))`, lit.Int(1)},
		{`(let +a 1 +b 2 +c (int (add a b)))`, lit.Int(3)},
		{`(if (let +a (int 1)) a)`, lit.Int(1)},
		{`(with +a 1 +b 2 +c (add a b) (add a b c))`, lit.Num(6)},
		{`(reduce 'hello' +e +i ['alice' 'bob'] (cat _ (if i ',') ' ' e))`,
			lit.Str("hello alice, bob"),
		},
		{`(with +a int @a)`, typ.Int},
		{`(with +a (obj +b int) @a.b)`, typ.Int},
		{`(with +f (fn + int 1) (f))`, lit.Int(1)},
		{`(with +f (fn +a int + int (add $a 1)) (f 1))`, lit.Int(2)},
	}
	for _, test := range tests {
		x, err := ParseString(test.raw)
		if err != nil {
			t.Errorf("%s parse err: %v", test.raw, err)
			continue
		}
		c := &Ctx{Exec: true}
		r, err := c.Resolve(NewScope(StdEnv), x)
		if err != nil {
			t.Errorf("%s resolve err: %v\n%v", test.raw, err, c.Unres)
			continue
		}
		if !reflect.DeepEqual(r, test.want) {
			t.Errorf("%s want %#v got %#v", test.raw, test.want, r)
		}
	}
}

func TestStdResolvePart(t *testing.T) {
	tests := []struct {
		raw  string
		want string
		typ  string
	}{
		{`(or x)`, `(bool x)`, "bool"},
		{`(or 0 x)`, `(bool x)`, "bool"},
		{`(or 1 x)`, `true`, "bool"},
		{`(and x)`, `(bool x)`, "bool"},
		{`(and 0 x)`, `false`, "bool"},
		{`(and 1 x)`, `(bool x)`, "bool"},
		{`(not x)`, `(not x)`, "bool"},
		{`(if 1 x)`, `x`, "void"},
		{`(if 0 1 x)`, `x`, "void"},
		{`(eq 1 x)`, `(eq 1 x)`, "bool"},
		{`(eq 1 x 1)`, `(eq 1 x)`, "bool"},
		{`(eq 1 1 x)`, `(eq 1 x)`, "bool"},
		{`(eq x 1 1)`, `(eq x 1)`, "bool"},
		{`(eq a b 1)`, `(eq a b 1)`, "bool"},
		{`(lt 0 1 x)`, `(lt 1 x)`, "bool"},
		{`(lt 0 x 2)`, `(lt 0 x 2)`, "bool"},
		{`(lt x 1 2)`, `(lt x 1)`, "bool"},
		{`(add x 2 3)`, `(add x 5)`, "void"},
		{`(add 1 x 3)`, `(add 4 x)`, "num"},
		{`(add 1 2 x)`, `(add 3 x)`, "num"},
		{`(sub x 2 3)`, `(sub x 5)`, "void"},
		{`(sub 1 x 3)`, `(sub -2 x)`, "num"},
		{`(sub 1 2 x)`, `(sub -1 x)`, "num"},
		{`(mul x 2 3)`, `(mul x 6)`, "void"},
		{`(mul 6 x 3)`, `(mul 18 x)`, "num"},
		{`(mul 6 2 x)`, `(mul 12 x)`, "num"},
		{`(div x 2 3)`, `(div x 6)`, "void"},
		{`(div 6 x 3)`, `(div 2 x)`, "num"},
		{`(div 6 2 x)`, `(div 3 x)`, "num"},
		{`(1 2 x)`, `(add 3 x)`, "num"},
		{`(not (bool x))`, `(not x)`, "bool"},
		{`(not (not x))`, `(bool x)`, "bool"},
		{`(not (not (not x)))`, `(not x)`, "bool"},
		{`(not (not (not (not x))))`, `(bool x)`, "bool"},
		{`(bool (bool x))`, `(bool x)`, "bool"},
		{`(bool (not x))`, `(not x)`, "bool"},
		{`(bool (not (bool x)))`, `(not x)`, "bool"},
		{`(bool (not (bool (not x))))`, `(bool x)`, "bool"},
	}
	for _, test := range tests {
		x, err := ParseString(test.raw)
		if err != nil {
			t.Errorf("%s parse err: %v", test.raw, err)
			continue
		}
		c := &Ctx{Exec: true, Part: true}
		r, err := c.Resolve(NewScope(StdEnv), x)
		if err != nil && err != ErrUnres {
			t.Errorf("%s resolve err expect ErrUnres, got: %v\n%v", test.raw, err, c.Unres)
			continue
		}
		if got := r.String(); got != test.want {
			t.Errorf("%s want %s got %s", test.raw, test.want, got)
		}
		if got, _ := elType(r); got.String() != test.typ {
			t.Errorf("%s want %s got %s", test.raw, test.typ, got)
		}
	}
}
