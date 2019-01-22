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

func (v Char) WriteBfr(b bfr.Ctx) error { return quoteBuffer(string(v), b) }
func (v Str) WriteBfr(b bfr.Ctx) error  { return quoteBuffer(string(v), b) }
func (v Raw) WriteBfr(b bfr.Ctx) error  { return quoteBuffer(v.Char(), b) }
func (v UUID) WriteBfr(b bfr.Ctx) error { return quoteBuffer(v.Char(), b) }

func sglQuoteString(v string) string         { s, _ := lex.Quote(v, '\''); return s }
func dblQuoteBytes(v string) ([]byte, error) { s, err := lex.Quote(v, '"'); return []byte(s), err }
func quoteBuffer(v string, b bfr.Ctx) (err error) {
	if b.JSON {
		v, err = lex.Quote(v, '"')
	} else {
		v, err = lex.Quote(v, '\'')
	}
	if err != nil {
		return err
	}
	return b.Fmt(v)
}
