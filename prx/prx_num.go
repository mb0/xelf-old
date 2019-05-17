package prx

import (
	"reflect"
	"strconv"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
)

type proxyNum struct{ proxy }

func (p *proxyNum) New() lit.Proxy { return &proxyNum{p.new()} }
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

func (p *proxyNum) Assign(l lit.Lit) error {
	l = lit.Deopt(l)
	if b, ok := l.(lit.Numeric); ok {
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
	} else if p.typ.Equal(l.Typ()) { // leaves null
		if v := p.el(); v.IsValid() {
			switch v.Kind() {
			case reflect.Int64, reflect.Int, reflect.Int32:
				v.SetInt(0)
			case reflect.Float64, reflect.Float32:
				v.SetFloat(0)
			case reflect.Uint64, reflect.Uint, reflect.Uint32:
				v.SetUint(0)
			}
			return nil
		}
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), p.typ)
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
func (p *proxyNum) WriteBfr(b *bfr.Ctx) error    { return b.Fmt(p.String()) }
