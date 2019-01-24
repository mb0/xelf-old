package lit

import (
	"reflect"
	"strconv"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

type (
	Num      float64
	Bool     bool
	Int      int64
	Real     float64
	proxyNum struct{ proxy }
)

func (Num) Typ() typ.Type  { return typ.Num }
func (Bool) Typ() typ.Type { return typ.Bool }
func (Int) Typ() typ.Type  { return typ.Int }
func (Real) Typ() typ.Type { return typ.Real }

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

func (v Num) WriteBfr(b bfr.Ctx) error  { return b.Fmt(v.String()) }
func (v Bool) WriteBfr(b bfr.Ctx) error { return b.Fmt(v.String()) }
func (v Int) WriteBfr(b bfr.Ctx) error  { return b.Fmt(v.String()) }
func (v Real) WriteBfr(b bfr.Ctx) error { return b.Fmt(v.String()) }

func (v *Bool) Ptr() interface{} { return v }
func (v *Bool) Assign(l Lit) error {
	if b, ok := l.(Numer); ok {
		if e, ok := b.Val().(bool); ok {
			*v = Bool(e)
			return nil
		}
	}
	return ErrNotAssignable
}

func (v *Int) Ptr() interface{} { return v }
func (v *Int) Assign(l Lit) error {
	if b, ok := l.(Numer); ok {
		if e, ok := b.Val().(int64); ok {
			*v = Int(e)
			return nil
		}
	}
	return ErrNotAssignable
}

func (v *Real) Ptr() interface{} { return v }
func (v *Real) Assign(l Lit) error {
	if b, ok := l.(Numer); ok {
		if e, ok := b.Val().(float64); ok {
			*v = Real(e)
			return nil
		}
	}
	return ErrNotAssignable
}

func (p *proxyNum) Val() interface{} {
	if v := p.el(); v.IsValid() {
		switch v.Kind() {
		case reflect.Int64, reflect.Int, reflect.Int32:
			return v.Int()
		case reflect.Float64, reflect.Float32:
			return v.Float()
		case reflect.Uint64, reflect.Uint, reflect.Uint32:
			return int64(v.Uint())
		}
	}
	return nil
}

func (p *proxyNum) Assign(l Lit) error {
	if b, ok := l.(Numer); ok {
		if v := p.el(); v.IsValid() {
			switch v.Kind() {
			case reflect.Int64, reflect.Int, reflect.Int32:
				if e, ok := b.Val().(int64); ok {
					v.SetInt(e)
					return nil
				}
			case reflect.Float64, reflect.Float32:
				if e, ok := b.Val().(float64); ok {
					v.SetFloat(e)
					return nil
				}
			case reflect.Uint64, reflect.Uint, reflect.Uint32:
				if e, ok := b.Val().(int64); ok {
					v.SetUint(uint64(e))
					return nil
				}
			}
		}
	}
	return ErrNotAssignable
}

func (p *proxyNum) IsZero() bool {
	switch v := p.Val().(type) {
	case int64:
		return v == 0
	case float64:
		return v == 0
	}
	return true
}

func (p *proxyNum) Num() float64 {
	switch v := p.Val().(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	}
	return 0
}

func (p *proxyNum) String() string {
	switch v := p.Val().(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	}
	return ""
}
func (p *proxyNum) MarshalJSON() ([]byte, error) { return []byte(p.String()), nil }
func (p *proxyNum) WriteBfr(b bfr.Ctx) error     { return b.Fmt(p.String()) }

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
func floatToString(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func floatToBytes(v float64) []byte  { return strconv.AppendFloat(nil, v, 'g', -1, 64) }
