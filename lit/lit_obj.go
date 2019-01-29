package lit

import (
	"errors"
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// MakeObj return a new abstract obj literal with the given type or an error.
func MakeObj(t typ.Type) (Obj, error) {
	if t.Kind&typ.MaskElem != typ.KindObj || t.Info == nil || len(t.Fields) == 0 {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, len(t.Fields))
	for _, f := range t.Fields {
		list = append(list, Keyed{f.Key(), Null(f.Type)})
	}
	return &abstrObj{t, Dict{list}}, nil
}

type (
	abstrObj struct {
		typ typ.Type
		Dict
	}
	proxyObj struct {
		proxy
		idx [][]int
	}
)

func (a *abstrObj) Typ() typ.Type { return a.typ }
func (a *abstrObj) IsZero() bool {
	if a.Dict.IsZero() {
		return true
	}
	for _, k := range a.List {
		if !k.Lit.IsZero() {
			return false
		}
	}
	return true
}
func (a *abstrObj) Idx(i int) (Lit, error) {
	_, err := a.typ.FieldByIdx(i)
	if err != nil {
		return nil, err
	}
	return a.Dict.List[i].Lit, nil
}
func (a *abstrObj) Key(key string) (Lit, error) {
	_, _, err := a.typ.FieldByKey(key)
	if err != nil {
		return nil, err
	}
	return a.Dict.Key(key)
}
func (a *abstrObj) SetIdx(i int, el Lit) error {
	f, err := a.typ.FieldByIdx(i)
	if err != nil {
		return err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return err
		}
	}
	a.Dict.List[i].Lit = el
	return nil
}
func (a *abstrObj) SetKey(key string, el Lit) error {
	f, _, err := a.typ.FieldByKey(key)
	if err != nil {
		return err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return err
		}
	}
	return a.Dict.SetKey(key, el)
}
func (a *abstrObj) IterIdx(it func(int, Lit) error) error {
	for i, el := range a.Dict.List {
		if err := it(i, el.Lit); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (a *abstrObj) String() string               { return bfr.String(a) }
func (a *abstrObj) MarshalJSON() ([]byte, error) { return bfr.JSON(a) }
func (a *abstrObj) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('{')
	n := 0
	for i, f := range a.typ.Fields {
		el, err := a.Idx(i)
		if err != nil {
			return err
		}
		if f.Opt() && el.IsZero() {
			continue
		}
		if n++; n > 1 {
			writeSep(b)
		}
		writeKey(b, f.Key())
		err = writeLit(b, el)
		if err != nil {
			return err
		}
	}
	return b.WriteByte('}')
}

var ErrObjProxyVal = errors.New("unexpected obj proxy value")

func (p *proxyObj) Assign(l Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return ErrNotAssignable
	}
	b, ok := l.(Keyer)
	if !ok || b.IsZero() { // a nil obj?
		v := p.val.Elem()
		v.Set(reflect.New(v.Type().Elem()))
		return nil
	}
	v, ok := p.elem(reflect.Struct)
	if !ok {
		return ErrObjProxyVal
	}
	return b.IterKey(func(k string, e Lit) error {
		_, i, err := p.typ.FieldByKey(k)
		if err != nil {
			return err
		}
		idx := p.idx[i]
		if len(idx) == 0 {
			return errors.New("no field index")
		}
		fv := v.FieldByIndex(idx)
		fl, err := ProxyValue(fv.Addr())
		if err != nil {
			return err
		}
		return fl.Assign(e)
	})
}

func (p *proxyObj) Len() int {
	if p.typ.Info != nil {
		return len(p.typ.Fields)
	}
	return 0
}
func (p *proxyObj) IsZero() bool {
	v := p.el()
	return !v.IsValid() || v.Kind() == reflect.Ptr && v.IsNil() || p.typ.Info.IsZero()
}
func (p *proxyObj) Keys() []string {
	if p.typ.Info != nil {
		res := make([]string, 0, len(p.typ.Fields))
		for _, f := range p.typ.Fields {
			res = append(res, f.Key())
		}
		return res
	}
	return nil
}
func (p *proxyObj) Idx(i int) (Lit, error) {
	f, err := p.typ.FieldByIdx(i)
	if err != nil {
		return nil, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		res, err := ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
		if err != nil {
			return nil, err
		}
		return Convert(res, f.Type, 0)
	}
	return Null(f.Type), nil
}
func (p *proxyObj) Key(k string) (Lit, error) {
	f, i, err := p.typ.FieldByKey(k)
	if err != nil {
		return nil, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		res, err := ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
		if err != nil {
			return nil, err
		}
		return Convert(res, f.Type, 0)
	}
	return Null(f.Type), nil
}
func (p *proxyObj) SetIdx(i int, l Lit) error {
	_, err := p.typ.FieldByIdx(i)
	if err != nil {
		return err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		return AssignToValue(l, v.FieldByIndex(p.idx[i]).Addr())
	}
	return ErrObjProxyVal
}
func (p *proxyObj) SetKey(k string, l Lit) error {
	_, i, err := p.typ.FieldByKey(k)
	if err != nil {
		return err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		return AssignToValue(l, v.FieldByIndex(p.idx[i]).Addr())
	}
	return ErrObjProxyVal
}
func (p *proxyObj) IterIdx(it func(int, Lit) error) (err error) {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Fields {
			var el Lit
			el, err = ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
			if err != nil {
				return err
			}
			el, err = Convert(el, f.Type, 0)
			if err != nil {
				return err
			}
			err = it(i, el)
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
func (p *proxyObj) IterKey(it func(string, Lit) error) (err error) {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Fields {
			var el Lit
			el, err = ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
			if err != nil {
				return err
			}
			el, err = Convert(el, f.Type, 0)
			if err != nil {
				return err
			}
			err = it(f.Key(), el)
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
func (p *proxyObj) String() string               { return bfr.String(p) }
func (p *proxyObj) MarshalJSON() ([]byte, error) { return bfr.JSON(p) }
func (p *proxyObj) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('{')
	n := 0
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Fields {
			el, err := ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
			if err != nil {
				return err
			}
			if f.Opt() && el.IsZero() {
				continue
			}
			if n++; n > 1 {
				writeSep(b)
			}
			writeKey(b, f.Key())
			err = writeLit(b, el)
			if err != nil {
				return err
			}
		}
	}
	return b.WriteByte('}')
}
