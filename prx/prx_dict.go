package prx

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
)

type proxyDict struct{ proxy }

func (p *proxyDict) New() lit.Proxy { return &proxyDict{p.new()} }
func (p *proxyDict) Assign(l lit.Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return cor.Errorf("%q not assignable to %q", l.Typ(), p.typ)
	}
	v, ok := p.elem(reflect.Map)
	if !ok {
		return ErrNotMap
	}
	b, ok := lit.Deopt(l).(lit.Keyer)
	if !ok || b.IsZero() { // a nil map
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	if v.IsNil() {
		v.Set(reflect.MakeMapWithSize(v.Type(), b.Len()))
	}
	return b.IterKey(func(k string, e lit.Lit) error {
		fp := reflect.New(v.Type().Elem())
		fl, err := ProxyValue(fp)
		if err != nil {
			return err
		}
		err = fl.Assign(e)
		if err != nil {
			return err
		}
		v.SetMapIndex(reflect.ValueOf(k), fp.Elem())
		return nil
	})
}

func (p *proxyDict) Element() (lit.Proxy, error) {
	if v, ok := p.elem(reflect.Map); ok {
		rt := v.Type().Elem()
		fp := reflect.New(rt)
		return ProxyValue(fp)
	}
	return nil, ErrNotMap
}

func (p *proxyDict) Len() int {
	if v, ok := p.elem(reflect.Map); ok {
		return v.Len()
	}
	return 0
}
func (p *proxyDict) IsZero() bool { return p.Len() == 0 }
func (p *proxyDict) Key(k string) (lit.Lit, error) {
	if v, ok := p.elem(reflect.Map); ok {
		return AdaptValue(v.MapIndex(reflect.ValueOf(k)))
	}
	return lit.Null(p.typ.Elem()), nil
}
func (p *proxyDict) SetKey(k string, l lit.Lit) (lit.Keyer, error) {
	if v, ok := p.elem(reflect.Map); ok {
		ev := reflect.New(v.Type().Elem())
		err := AssignToValue(l, ev)
		if err != nil {
			return p, err
		}
		if v.IsNil() {
			v = reflect.MakeMap(v.Type())
			p.val.Elem().Set(v)
		}
		v.SetMapIndex(reflect.ValueOf(k), ev.Elem())
		return p, nil
	}
	return p, cor.Errorf("not a map keyer")
}

func (p *proxyDict) Keys() (res []string) {
	if v, ok := p.elem(reflect.Map); ok {
		keys := v.MapKeys()
		res = make([]string, 0, len(keys))
		for _, key := range keys {
			res = append(res, key.String())
		}
	}
	return res
}

func (p *proxyDict) IterKey(it func(string, lit.Lit) error) error {
	if v, ok := p.elem(reflect.Map); ok {
		keys := v.MapKeys()
		for _, k := range keys {
			el, err := AdaptValue(v.MapIndex(k))
			if err != nil {
				return err
			}
			err = it(k.String(), el)
			if err != nil {
				if err == lit.BreakIter {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func (p *proxyDict) String() string               { return bfr.String(p) }
func (p *proxyDict) MarshalJSON() ([]byte, error) { return bfr.JSON(p) }
func (p *proxyDict) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('{')
	i := 0
	err := p.IterKey(func(k string, el lit.Lit) error {
		if i > 0 {
			writeSep(b)
		}
		i++
		writeKey(b, k)
		return writeLit(b, el)
	})
	if err != nil {
		return err
	}
	return b.WriteByte('}')
}

func writeSep(b *bfr.Ctx) error {
	if b.JSON {
		return b.WriteByte(',')
	}
	return b.WriteByte(' ')
}

func writeLit(b *bfr.Ctx, e lit.Lit) error {
	if e == nil {
		return b.Fmt("null")
	}
	return e.WriteBfr(b)
}
func writeKey(b *bfr.Ctx, key string) (err error) {
	if !b.JSON && cor.IsName(key) {
		b.WriteString(key)
		return b.WriteByte(':')
	}
	if b.JSON {
		key, err = cor.Quote(key, '"')
	} else {
		key, err = cor.Quote(key, '\'')
	}
	if err != nil {
		return err
	}
	b.WriteString(key)
	return b.WriteByte(':')
}

var _ lit.Dictionary = &proxyDict{}
