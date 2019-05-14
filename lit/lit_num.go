package lit

import (
	"strconv"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type (
	Num  float64
	Bool bool
	Int  int64
	Real float64
)

func (v Num) Typ() typ.Type  { return typ.Num }
func (v Bool) Typ() typ.Type { return typ.Bool }
func (v Int) Typ() typ.Type  { return typ.Int }
func (v Real) Typ() typ.Type { return typ.Real }

func (v Num) IsZero() bool  { return v == 0 }
func (v Bool) IsZero() bool { return v == false }
func (v Int) IsZero() bool  { return v == 0 }
func (v Real) IsZero() bool { return v == 0 }

func (v Num) Num() float64  { return float64(v) }
func (v Bool) Num() float64 { return boolToFloat(bool(v)) }
func (v Int) Num() float64  { return float64(v) }
func (v Real) Num() float64 { return float64(v) }

func (v Num) Val() interface{}  { return float64(v) }
func (v Bool) Val() interface{} { return bool(v) }
func (v Int) Val() interface{}  { return int64(v) }
func (v Real) Val() interface{} { return float64(v) }

func (v Num) String() string  { return floatToString(float64(v)) }
func (v Bool) String() string { return strconv.FormatBool(bool(v)) }
func (v Int) String() string  { return strconv.FormatInt(int64(v), 10) }
func (v Real) String() string { return floatToString(float64(v)) }

func (v Num) MarshalJSON() ([]byte, error)  { return floatToBytes(float64(v)), nil }
func (v Bool) MarshalJSON() ([]byte, error) { return strconv.AppendBool(nil, bool(v)), nil }
func (v Int) MarshalJSON() ([]byte, error)  { return strconv.AppendInt(nil, int64(v), 10), nil }
func (v Real) MarshalJSON() ([]byte, error) { return floatToBytes(float64(v)), nil }

func (v Num) WriteBfr(b *bfr.Ctx) error  { return b.Fmt(v.String()) }
func (v Bool) WriteBfr(b *bfr.Ctx) error { return b.Fmt(v.String()) }
func (v Int) WriteBfr(b *bfr.Ctx) error  { return b.Fmt(v.String()) }
func (v Real) WriteBfr(b *bfr.Ctx) error { return b.Fmt(v.String()) }

func (v *Bool) Ptr() interface{} { return v }
func (v *Bool) Assign(l Lit) error {
	l = Deopt(l)
	if b, ok := l.(Numeric); ok {
		if e, ok := b.Val().(bool); ok {
			*v = Bool(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = false
		return nil
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func (v *Int) Ptr() interface{} { return v }
func (v *Int) Assign(l Lit) error {
	l = Deopt(l)
	if b, ok := l.(Numeric); ok {
		if e, ok := b.Val().(int64); ok {
			*v = Int(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = 0
		return nil
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func (v *Real) Ptr() interface{} { return v }
func (v *Real) Assign(l Lit) error {
	l = Deopt(l)
	if b, ok := l.(Numeric); ok {
		if e, ok := b.Val().(float64); ok {
			*v = Real(e)
			return nil
		}
	} else if v.Typ().Equal(l.Typ()) { // leaves null
		*v = 0
		return nil
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), v.Typ())
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
func floatToString(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func floatToBytes(v float64) []byte  { return strconv.AppendFloat(nil, v, 'g', -1, 64) }
