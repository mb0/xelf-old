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

func (Time) Typ() typ.Type { return typ.Time }
func (Span) Typ() typ.Type { return typ.Span }

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

func (v Time) WriteBfr(b bfr.Ctx) error { return quoteBuffer(v.Char(), b) }
func (v Span) WriteBfr(b bfr.Ctx) error { return quoteBuffer(v.Char(), b) }

func (v Span) Seconds() float64 {
	return time.Duration(v).Seconds()
}

func (v *Time) Assign(l Lit) error {
	if b, ok := l.(Charer); ok {
		if e, ok := b.Val().(time.Time); ok {
			*v = Time(e)
			return nil
		}
	}
	return ErrNotAssignable
}
func (v *Span) Assign(l Lit) error {
	if b, ok := l.(Charer); ok {
		if e, ok := b.Val().(time.Duration); ok {
			*v = Span(e)
			return nil
		}
	}
	return ErrNotAssignable
}
