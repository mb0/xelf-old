package lit

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// MakeDict returns a new abstract dict literal with the given type or an error.
func MakeDict(t typ.Type) (*Dict, error) {
	return MakeDictCap(t, 0)
}

// MakeDictCap returns a new abstract dict literal with the given type and cap or an error.
func MakeDictCap(t typ.Type, cap int) (*Dict, error) {
	if t.Kind&typ.MaskElem != typ.KindDict {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, cap)
	return &Dict{t.Elem(), Keyr{list}}, nil
}

type (
	Dict struct {
		Elem typ.Type
		Keyr
	}
	proxyDict struct{ proxy }
)

func (a *Dict) Typ() typ.Type     { return typ.Dict(a.Elem) }
func (a *Dict) Element() typ.Type { return a.Elem }
func (a *Dict) Key(k string) (Lit, error) {
	l, err := a.Keyr.Key(k)
	if err != nil {
		return nil, err
	}
	if l == Nil {
		l = Null(a.Elem)
	}
	return l, nil
}
func (a *Dict) SetKey(key string, el Lit) (res Keyer, err error) {
	if el != nil {
		el, err = Convert(el, a.Elem, 0)
		if err != nil {
			return a, err
		}
	}
	res, err = a.Keyr.SetKey(key, el)
	if err != nil {
		return a, err
	}
	a.Keyr = *res.(*Keyr)
	return a, nil
}

func (p *proxyDict) Assign(l Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return cor.Errorf("%q not assignable to %q", l.Typ(), p.typ)
	}
	v, ok := p.elem(reflect.Map)
	if !ok {
		return ErrNotMap
	}
	b, ok := Deopt(l).(Keyer)
	if !ok || b.IsZero() { // a nil map
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	if v.IsNil() {
		v.Set(reflect.MakeMapWithSize(v.Type(), b.Len()))
	}
	return b.IterKey(func(k string, e Lit) error {
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
func (p *proxyDict) Element() typ.Type { return p.typ.Elem() }
func (p *proxyDict) Len() int {
	if v, ok := p.elem(reflect.Map); ok {
		return v.Len()
	}
	return 0
}
func (p *proxyDict) IsZero() bool { return p.Len() == 0 }
func (p *proxyDict) Key(k string) (Lit, error) {
	if v, ok := p.elem(reflect.Map); ok {
		return AdaptValue(v.MapIndex(reflect.ValueOf(k)))
	}
	return Null(p.typ.Elem()), nil
}
func (p *proxyDict) SetKey(k string, l Lit) (Keyer, error) {
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
	return p, cor.Errorf("nil keyer")
}

func (p *proxyDict) Keys() []string {
	if v, ok := p.elem(reflect.Map); ok {
		keys := v.MapKeys()
		res := make([]string, 0, len(keys))
		for _, key := range keys {
			res = append(res, key.String())
		}
	}
	return nil
}

func (p *proxyDict) IterKey(it func(string, Lit) error) error {
	if v, ok := p.elem(reflect.Map); ok {
		keys := v.MapKeys()
		for _, k := range keys {
			el, err := AdaptValue(v.MapIndex(k))
			if err != nil {
				return err
			}
			err = it(k.String(), el)
			if err != nil {
				if err == BreakIter {
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
	err := p.IterKey(func(k string, el Lit) error {
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

var _, _ Dictionary = &Dict{}, &proxyDict{}
