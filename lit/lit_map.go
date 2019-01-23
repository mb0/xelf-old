package lit

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// MakeMap returns a new abstract map literal with the given type or an error.
func MakeMap(t typ.Type) (Map, error) {
	return MakeMapCap(t, 0)
}

// MakeMapCap returns a new abstract map literal with the given type and cap or an error.
func MakeMapCap(t typ.Type, cap int) (Map, error) {
	if t.Kind&typ.MaskElem != typ.KindMap {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, cap)
	return &abstrMap{t.Next(), Dict{list}}, nil
}

type (
	abstrMap struct {
		elem typ.Type
		Dict
	}
	proxyMap struct{ proxy }
)

func (a *abstrMap) Typ() typ.Type  { return typ.Map(a.elem) }
func (a *abstrMap) Elem() typ.Type { return a.elem }
func (a *abstrMap) Key(k string) (Lit, error) {
	l, err := a.Dict.Key(k)
	if err != nil {
		return nil, err
	}
	if l == Nil {
		l = Null(a.elem)
	}
	return l, nil
}
func (a *abstrMap) SetKey(key string, el Lit) (err error) {
	if el != nil {
		el, err = Convert(el, a.elem, 0)
		if err != nil {
			return err
		}
	}
	return a.Dict.SetKey(key, el)
}

func (p *proxyMap) Assign(l Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return ErrNotAssignable
	}
	v, ok := p.elem(reflect.Map)
	if !ok {
		return ErrNotAssignable
	}
	b, ok := l.(Keyer)
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
func (p *proxyMap) Elem() typ.Type { return p.typ.Next() }
func (p *proxyMap) Len() int {
	if v, ok := p.elem(reflect.Map); ok {
		return v.Len()
	}
	return 0
}
func (p *proxyMap) IsZero() bool { return p.Len() == 0 }
func (p *proxyMap) Key(k string) (Lit, error) {
	if v, ok := p.elem(reflect.Map); ok {
		return AdaptValue(v.MapIndex(reflect.ValueOf(k)))
	}
	return Null(p.typ.Next()), nil
}
func (p *proxyMap) SetKey(k string, l Lit) error {
	if v, ok := p.elem(reflect.Map); ok {
		ev := reflect.New(v.Type().Elem())
		err := AssignToValue(l, ev)
		if err != nil {
			return err
		}
		v.SetMapIndex(reflect.ValueOf(k), ev.Elem())
		return nil
	}
	return ErrNilKeyer
}
func (p *proxyMap) IterKey(it func(string, Lit) error) error {
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

func (p *proxyMap) String() string               { return bfr.String(p) }
func (p *proxyMap) MarshalJSON() ([]byte, error) { return bfr.JSON(p) }
func (p *proxyMap) WriteBfr(b bfr.Ctx) error {
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
