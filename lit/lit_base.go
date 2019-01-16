package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// Null represents a typed zero value
type Null typ.Type

func (n Null) Typ() typ.Type { return typ.Type(n) }
func (Null) IsZero() bool    { return true }
func (Null) Some() Lit       { return nil }

func (Null) String() string               { return "null" }
func (Null) MarshalJSON() ([]byte, error) { return []byte("null"), nil }
func (Null) WriteBfr(b bfr.Ctx) error     { _, err := b.WriteString("null"); return err }

// Some represents non-null option
type Some struct{ Lit }

func (s Some) Typ() typ.Type { return typ.Opt(s.Lit.Typ()) }
func (_ Some) IsZero() bool  { return false }
func (s Some) Some() Lit     { return s.Lit }

// Any represents a non-null, any-typed literal
type Any struct{ Lit }

func (Any) Typ() typ.Type { return typ.Any }
func (a Any) Some() Lit   { return a.Lit }
