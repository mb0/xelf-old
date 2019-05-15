package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// Null represents a typed zero value.
type Null typ.Type

func (n Null) Typ() typ.Type { return typ.Type(n) }
func (n Null) IsZero() bool  { return true }
func (n Null) Some() Lit     { return nil }

func (n Null) String() string               { return "null" }
func (n Null) MarshalJSON() ([]byte, error) { return []byte("null"), nil }
func (n Null) WriteBfr(b *bfr.Ctx) error    { _, err := b.WriteString("null"); return err }

// Some represents non-null option.
type Some struct{ Lit }

func (s Some) Typ() typ.Type { return typ.Opt(s.Lit.Typ()) }
func (s Some) Some() Lit     { return s.Lit }

// SomeAssignable represents non-null assignable option.
type SomeAssignable struct{ Assignable }

func (s SomeAssignable) Typ() typ.Type { return typ.Opt(s.Assignable.Typ()) }
func (s SomeAssignable) Some() Lit     { return s.Assignable }

// Any represents a non-null, any-typed literal.
type Any struct{ Lit }

func (a Any) Typ() typ.Type { return typ.Any }
func (a Any) Some() Lit     { return a.Lit }

// BitsInt represents a bits int constant
type BitsInt struct {
	Type typ.Type
	Int
}

func (c BitsInt) Typ() typ.Type { return c.Type }

// EnumStr represents a enum str constant
type EnumStr struct {
	Type typ.Type
	Str
}

func (c EnumStr) Typ() typ.Type { return c.Type }
