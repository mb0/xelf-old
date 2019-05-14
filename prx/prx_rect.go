package prx

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
)

type proxyRec struct {
	proxy
	idx [][]int
}

func (p *proxyRec) Assign(l lit.Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return cor.Errorf("%q not assignable to %q", l.Typ(), p.typ)
	}
	b, ok := lit.Deopt(l).(lit.Keyer)
	if !ok || b.IsZero() { // a nil rec?
		v := p.val.Elem()
		v.Set(reflect.New(v.Type().Elem()))
		return nil
	}
	v, ok := p.elem(reflect.Struct)
	if !ok {
		return ErrNotStruct
	}
	return b.IterKey(func(k string, e lit.Lit) error {
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

func (p *proxyRec) Len() int {
	if p.typ.Info != nil {
		return len(p.typ.Params)
	}
	return 0
}
func (p *proxyRec) IsZero() bool {
	v := p.el()
	return !v.IsValid() || v.Kind() == reflect.Ptr && v.IsNil() || p.typ.Info.IsZero()
}
func (p *proxyRec) Keys() []string {
	if p.typ.HasParams() {
		res := make([]string, 0, len(p.typ.Params))
		for _, f := range p.typ.Params {
			res = append(res, f.Key())
		}
		return res
	}
	return nil
}
func (p *proxyRec) Idx(i int) (lit.Lit, error) {
	f, err := p.typ.ParamByIdx(i)
	if err != nil {
		return nil, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		res, err := ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
		if err != nil {
			return nil, err
		}
		return lit.Convert(res, f.Type, 0)
	}
	return lit.Null(f.Type), nil
}
func (p *proxyRec) Key(k string) (lit.Lit, error) {
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
		return lit.Convert(res, f.Type, 0)
	}
	return lit.Null(f.Type), nil
}
func (p *proxyRec) SetIdx(i int, l lit.Lit) (lit.Indexer, error) {
	_, err := p.typ.ParamByIdx(i)
	if err != nil {
		return p, err
	}
	if v, ok := p.elem(reflect.Struct); ok {
		return p, AssignToValue(l, v.FieldByIndex(p.idx[i]).Addr())
	}
	return p, ErrNotStruct
}
func (p *proxyRec) SetKey(k string, l lit.Lit) (lit.Keyer, error) {
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
func (p *proxyRec) IterIdx(it func(int, lit.Lit) error) (err error) {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Params {
			var el lit.Lit
			el, err = ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
			if err != nil {
				return err
			}
			el, err = lit.Convert(el, f.Type, 0)
			if err != nil {
				return err
			}
			err = it(i, el)
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
func (p *proxyRec) IterKey(it func(string, lit.Lit) error) (err error) {
	if v, ok := p.elem(reflect.Struct); ok && p.typ.Info != nil {
		for i, f := range p.typ.Params {
			var el lit.Lit
			el, err = ProxyValue(v.FieldByIndex(p.idx[i]).Addr())
			if err != nil {
				return err
			}
			el, err = lit.Convert(el, f.Type, 0)
			if err != nil {
				return err
			}
			err = it(f.Key(), el)
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
func (p *proxyRec) String() string               { return bfr.String(p) }
func (p *proxyRec) MarshalJSON() ([]byte, error) { return bfr.JSON(p) }
func (p *proxyRec) WriteBfr(b *bfr.Ctx) error {
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
