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
	_, err := idxField(a.typ, i)
	if err != nil {
		return nil, err
	}
	return a.Dict.List[i].Lit, nil
}
func (a *abstrObj) Key(key string) (Lit, error) {
	_, _, err := keyField(a.typ, key)
	if err != nil {
		return nil, err
	}
	return a.Dict.Key(key)
}
func (a *abstrObj) SetIdx(i int, el Lit) error {
	f, err := idxField(a.typ, i)
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
	f, _, err := keyField(a.typ, key)
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
		_, i, err := keyField(p.typ, k)
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
	f, err := idxField(p.typ, i)
	if err != nil {
		return nil, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		res, err := AdaptValue(v.FieldByIndex(p.idx[i]))
		if err != nil {
			return nil, err
		}
		return Convert(res, f.Type, 0)
	}
	return Null(f.Type), nil
}
func (p *proxyObj) Key(k string) (Lit, error) {
	f, i, err := keyField(p.typ, k)
	if err != nil {
		return nil, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		res, err := AdaptValue(v.FieldByIndex(p.idx[i]))
		if err != nil {
			return nil, err
		}
		return Convert(res, f.Type, 0)
	}
	return Null(f.Type), nil
}
func (p *proxyObj) SetIdx(i int, l Lit) error {
	_, err := idxField(p.typ, i)
	if err != nil {
		return err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		return AssignToValue(l, v.FieldByIndex(p.idx[i]).Addr())
	}
	return ErrObjProxyVal
}
func (p *proxyObj) SetKey(k string, l Lit) error {
	_, i, err := keyField(p.typ, k)
	if err != nil {
		return err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		return AssignToValue(l, v.FieldByIndex(p.idx[i]).Addr())
	}
	return ErrObjProxyVal
}
func (p *proxyObj) IterIdx(it func(int, Lit) error) error {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Fields {
			el, err := AdaptValue(v.FieldByIndex(p.idx[i]))
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
func (p *proxyObj) IterKey(it func(string, Lit) error) error {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Fields {
			el, err := AdaptValue(v.FieldByIndex(p.idx[i]))
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

func idxField(t typ.Type, i int) (*typ.Field, error) {
	if n := t.Info; n != nil {
		if i >= 0 && i < len(n.Fields) {
			return &n.Fields[i], nil
		}
	}
	return nil, ErrIdxBounds
}
func keyField(t typ.Type, k string) (*typ.Field, int, error) {
	if n := t.Info; n != nil {
		for i, f := range n.Fields {
			if f.Key() == k {
				return &f, i, nil
			}
		}
	}
	return nil, 0, typ.ErrFieldName
}
