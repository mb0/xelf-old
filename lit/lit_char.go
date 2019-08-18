package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type (
	Char string
	Str  string
	Raw  []byte
	UUID [16]byte
)

func (v Char) Typ() typ.Type { return typ.Char }
func (v Str) Typ() typ.Type  { return typ.Str }
func (v Raw) Typ() typ.Type  { return typ.Raw }
func (v UUID) Typ() typ.Type { return typ.UUID }

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
func (v *UUID) UnmarshalText(b []byte) (err error) {
	*v, err = cor.ParseUUID(string(b))
	return err
}

func (v Char) WriteBfr(b *bfr.Ctx) error { return b.Quote(string(v)) }
func (v Str) WriteBfr(b *bfr.Ctx) error  { return b.Quote(string(v)) }
func (v Raw) WriteBfr(b *bfr.Ctx) error  { return b.Quote(v.Char()) }
func (v UUID) WriteBfr(b *bfr.Ctx) error { return b.Quote(v.Char()) }

func (v Str) Len() int  { return len(v) }
func (v Char) Len() int { return len(v) }
func (v Raw) Len() int  { return len(v) }

func (v *Str) New() Proxy       { return new(Str) }
func (v *Str) Ptr() interface{} { return v }
func (v *Str) Assign(l Lit) error {
	if o := Deopt(l); o == nil {
		ot, _ := l.Typ().Deopt()
		if v.Typ().Equal(ot) {
			*v = ""
			return nil
		}
	} else if b, ok := o.(Character); ok {
		if e, ok := b.Val().(string); ok {
			*v = Str(e)
			return nil
		}
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func (v *Raw) New() Proxy       { return new(Raw) }
func (v *Raw) Ptr() interface{} { return v }
func (v *Raw) Assign(l Lit) error {
	if o := Deopt(l); o == nil {
		ot, _ := l.Typ().Deopt()
		if v.Typ().Equal(ot) {
			*v = nil
			return nil
		}
	} else if b, ok := o.(Character); ok {
		if e, ok := b.Val().([]byte); ok {
			*v = Raw(e)
			return nil
		}
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func (v *UUID) New() Proxy       { return new(UUID) }
func (v *UUID) Ptr() interface{} { return v }
func (v *UUID) Assign(l Lit) error {
	if o := Deopt(l); o == nil {
		ot, _ := l.Typ().Deopt()
		if v.Typ().Equal(ot) {
			*v = ZeroUUID
			return nil
		}
	} else if b, ok := o.(Character); ok {
		if e, ok := b.Val().([16]byte); ok {
			*v = UUID(e)
			return nil
		}
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func sglQuoteString(v string) string         { s, _ := cor.Quote(v, '\''); return s }
func dblQuoteBytes(v string) ([]byte, error) { s, err := cor.Quote(v, '"'); return []byte(s), err }
