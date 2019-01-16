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

func (v Time) IsZero() bool { return time.Time(v) == time.Time{} }
func (v Span) IsZero() bool { return v == 0 }

func (v Time) Num() float64 { return float64(cor.UnixMilli(time.Time(v))) }
func (v Span) Num() float64 { return float64(v / 1000000) }

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
