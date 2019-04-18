package lit

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// MakeObj return a new abstract obj literal with the given type or an error.
func MakeObj(t typ.Type) (*DictObj, error) {
	if t.Kind&typ.MaskElem != typ.KindObj || t.Info == nil || len(t.Params) == 0 {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, len(t.Params))
	for _, f := range t.Params {
		list = append(list, Keyed{f.Key(), ZeroProxy(f.Type)})
	}
	return &DictObj{t, Dict{list}}, nil
}

// ObjFromKeyed creates a new abstract obj literal from the given list of keyed literals.
func ObjFromKeyed(list []Keyed) *DictObj {
	fs := make([]typ.Param, 0, len(list))
	for _, d := range list {
		fs = append(fs, typ.Param{d.Key, d.Lit.Typ()})
	}
	return &DictObj{typ.Obj(fs), Dict{list}}
}

type (
	DictObj struct {
		Type typ.Type
		Dict
	}
	proxyObj struct {
		proxy
		idx [][]int
	}
)

func (a *DictObj) Typ() typ.Type { return a.Type }
func (a *DictObj) IsZero() bool {
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
func (a *DictObj) Idx(i int) (Lit, error) {
	_, err := a.Type.ParamByIdx(i)
	if err != nil {
		return nil, err
	}
	return a.Dict.List[i].Lit, nil
}
func (a *DictObj) Key(key string) (Lit, error) {
	_, _, err := a.Type.ParamByKey(key)
	if err != nil {
		return nil, err
	}
	return a.Dict.Key(key)
}
func (a *DictObj) SetIdx(i int, el Lit) (Idxer, error) {
	f, err := a.Type.ParamByIdx(i)
	if err != nil {
		return a, err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return a, err
		}
	}
	a.Dict.List[i].Lit = el
	return a, nil
}
func (a *DictObj) SetKey(key string, el Lit) (Keyer, error) {
	f, _, err := a.Type.ParamByKey(key)
	if err != nil {
		return a, err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return a, err
		}
	}
	res, err := a.Dict.SetKey(key, el)
	if err != nil {
		return a, err
	}
	a.Dict = *res.(*Dict)
	return a, nil
}
func (a *DictObj) IterIdx(it func(int, Lit) error) error {
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
func (a *DictObj) String() string               { return bfr.String(a) }
func (a *DictObj) MarshalJSON() ([]byte, error) { return bfr.JSON(a) }
func (a *DictObj) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('{')
	n := 0
	for i, f := range a.Type.Params {
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

func (p *proxyObj) Assign(l Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return cor.Errorf("%q not assignable to %q", l.Typ(), p.typ)
	}
	b, ok := Deopt(l).(Keyer)
	if !ok || b.IsZero() { // a nil obj?
		v := p.val.Elem()
		v.Set(reflect.New(v.Type().Elem()))
		return nil
	}
	v, ok := p.elem(reflect.Struct)
	if !ok {
		return ErrNotStruct
	}
	return b.IterKey(func(k string, e Lit) error {
		_, i, err := p.typ.ParamByKey(k)
		if err != nil {
			return err
		}
		idx := p.idx[i]
		if len(idx) == 0 {
			return cor.Error("no field index")
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
		return len(p.typ.Params)
	}
	return 0
}
func (p *proxyObj) IsZero() bool {
	v := p.el()
	return !v.IsValid() || v.Kind() == reflect.Ptr && v.IsNil() || p.typ.Info.IsZero()
}
func (p *proxyObj) Keys() []string {
	if p.typ.Info != nil {
		res := make([]string, 0, len(p.typ.Params))
		for _, f := range p.typ.Params {
			res = append(res, f.Key())
		}
		return res
	}
	return nil
}
func (p *proxyObj) Idx(i int) (Lit, error) {
	f, err := p.typ.ParamByIdx(i)
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
	f, i, err := p.typ.ParamByKey(k)
	if err != nil {
		return nil, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		v = fieldByIndex(v, p.idx[i])
		res, err := ProxyValue(v.Addr())
		if err != nil {
			return nil, err
		}
		return Convert(res, f.Type, 0)
	}
	return Null(f.Type), nil
}
func (p *proxyObj) SetIdx(i int, l Lit) (Idxer, error) {
	_, err := p.typ.ParamByIdx(i)
	if err != nil {
		return p, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		return p, AssignToValue(l, v.FieldByIndex(p.idx[i]).Addr())
	}
	return p, ErrNotStruct
}
func (p *proxyObj) SetKey(k string, l Lit) (Keyer, error) {
	_, i, err := p.typ.ParamByKey(k)
	if err != nil {
		return p, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		v = fieldByIndex(v, p.idx[i])
		return p, AssignToValue(l, v.Addr())
	}
	return p, ErrNotStruct
}
func fieldByIndex(v reflect.Value, idx []int) reflect.Value {
	for _, x := range idx {
		if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(x)
	}
	return v
}
func (p *proxyObj) IterIdx(it func(int, Lit) error) (err error) {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Params {
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
		for i, f := range p.typ.Params {
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
func (p *proxyObj) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('{')
	n := 0
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Params {
			v := fieldByIndex(v, p.idx[i])
			el, err := ProxyValue(v.Addr())
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
