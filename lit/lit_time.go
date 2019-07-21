package lit

import (
	"time"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type (
	Time time.Time
	Span time.Duration
)

func (v Time) Typ() typ.Type { return typ.Time }
func (v Span) Typ() typ.Type { return typ.Span }

func (v Time) IsZero() bool { return v == ZeroTime }
func (v Span) IsZero() bool { return v == ZeroSpan }

func (v Time) Num() float64 { return float64(cor.UnixMilli(time.Time(v))) }
func (v Span) Num() float64 { return float64(cor.Milli(time.Duration(v))) }

func (v Time) Char() string { return cor.FormatTime(time.Time(v)) }
func (v Span) Char() string { return cor.FormatSpan(time.Duration(v)) }

func (v Time) Val() interface{} { return time.Time(v) }
func (v Span) Val() interface{} { return time.Duration(v) }

func (v Time) String() string { return sglQuoteString(v.Char()) }
func (v Span) String() string { return sglQuoteString(v.Char()) }

func (v Time) MarshalJSON() ([]byte, error) { return dblQuoteBytes(v.Char()) }
func (v Span) MarshalJSON() ([]byte, error) { return dblQuoteBytes(v.Char()) }
func (v *Span) UnmarshalText(b []byte) error {
	s, err := cor.ParseSpan(string(b))
	*v = Span(s)
	return err
}

func (v Time) WriteBfr(b *bfr.Ctx) error { return b.Quote(v.Char()) }
func (v Span) WriteBfr(b *bfr.Ctx) error { return b.Quote(v.Char()) }

func (v Span) Seconds() float64 {
	return time.Duration(v).Seconds()
}

func (v *Time) New() Proxy       { return new(Time) }
func (v *Time) Ptr() interface{} { return v }
func (v *Time) Assign(l Lit) error {
	l = Deopt(l)
	if b, ok := l.(Character); ok {
		if e, ok := b.Val().(time.Time); ok {
			*v = Time(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = ZeroTime
		return nil
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func (v *Span) New() Proxy       { return new(Span) }
func (v *Span) Ptr() interface{} { return v }
func (v *Span) Assign(l Lit) error {
	l = Deopt(l)
	if b, ok := l.(Character); ok {
		if e, ok := b.Val().(time.Duration); ok {
			*v = Span(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = 0
		return nil
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}
