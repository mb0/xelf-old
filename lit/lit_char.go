package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

type (
	Char string
	Str  string
	Raw  []byte
	UUID [16]byte
)

func (Char) Typ() typ.Type { return typ.Char }
func (Str) Typ() typ.Type  { return typ.Str }
func (Raw) Typ() typ.Type  { return typ.Raw }
func (UUID) Typ() typ.Type { return typ.UUID }

func (v Char) IsZero() bool { return v == "" }
func (v Str) IsZero() bool  { return v == "" }
func (v Raw) IsZero() bool  { return len(v) == 0 }
func (v UUID) IsZero() bool { return v == ZeroUUID }

func (v Char) Char() string { return string(v) }
func (v Str) Char() string  { return string(v) }
func (v Raw) Char() string  { return cor.FormatRaw(v) }
func (v UUID) Char() string { return cor.FormatUUID(v) }

func (v Char) Val() interface{} { return string(v) }
func (v Str) Val() interface{}  { return string(v) }
func (v Raw) Val() interface{}  { return []byte(v) }
func (v UUID) Val() interface{} { return [16]byte(v) }

func (v Char) String() string { return sglQuoteString(string(v)) }
func (v Str) String() string  { return sglQuoteString(string(v)) }
func (v Raw) String() string  { return sglQuoteString(v.Char()) }
func (v UUID) String() string { return sglQuoteString(v.Char()) }

func (v Char) MarshalJSON() ([]byte, error) { return dblQuoteBytes(string(v)) }
func (v Str) MarshalJSON() ([]byte, error)  { return dblQuoteBytes(string(v)) }
func (v Raw) MarshalJSON() ([]byte, error)  { return dblQuoteBytes(v.Char()) }
func (v UUID) MarshalJSON() ([]byte, error) { return dblQuoteBytes(v.Char()) }

func (v Char) WriteBfr(b bfr.Ctx) error { return b.Quote(string(v)) }
func (v Str) WriteBfr(b bfr.Ctx) error  { return b.Quote(string(v)) }
func (v Raw) WriteBfr(b bfr.Ctx) error  { return b.Quote(v.Char()) }
func (v UUID) WriteBfr(b bfr.Ctx) error { return b.Quote(v.Char()) }

func (v *Str) Ptr() interface{} { return v }
func (v *Str) Assign(l Lit) error {
	if b, ok := l.(Charer); ok {
		if e, ok := b.Val().(string); ok {
			*v = Str(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = ""
		return nil
	}
	return ErrAssign(l.Typ(), v.Typ())
}

func (v *Raw) Ptr() interface{} { return v }
func (v *Raw) Assign(l Lit) error {
	if b, ok := l.(Charer); ok {
		if e, ok := b.Val().([]byte); ok {
			*v = Raw(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = nil
		return nil
	}
	return ErrAssign(l.Typ(), v.Typ())
}

func (v *UUID) Ptr() interface{} { return v }
func (v *UUID) Assign(l Lit) error {
	if b, ok := l.(Charer); ok {
		if e, ok := b.Val().([16]byte); ok {
			*v = UUID(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = ZeroUUID
		return nil
	}
	return ErrAssign(l.Typ(), v.Typ())
}

func sglQuoteString(v string) string         { s, _ := lex.Quote(v, '\''); return s }
func dblQuoteBytes(v string) ([]byte, error) { s, err := lex.Quote(v, '"'); return []byte(s), err }
