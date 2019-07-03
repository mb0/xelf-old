package prx

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
)

type proxyList struct{ proxy }

func (p *proxyList) New() lit.Proxy { return &proxyList{p.new()} }
func (p *proxyList) Assign(l lit.Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return cor.Errorf("%q not assignable to %q", l.Typ(), p.typ)
	}
	b, ok := l.(lit.Indexer)
	if !ok || b.IsZero() { // a nil list?
		if p.val.CanAddr() {
			p.val.Set(reflect.Zero(p.val.Type()))
		} else {
			p.val = reflect.New(p.val.Type().Elem())
		}
		return nil
	}
	v, ok := p.elem(reflect.Slice)
	if !ok {
		return ErrNotSlice
	}
	v = v.Slice(0, 0)
	err := b.IterIdx(func(i int, e lit.Lit) error {
		fp := reflect.New(v.Type().Elem())
		fl, err := ProxyValue(fp)
		if err != nil {
			return err
		}
		err = fl.Assign(e)
		if err != nil {
			return err
		}
		v = reflect.Append(v, fp.Elem())
		return nil
	})
	if err != nil {
		return err
	}
	pv := p.val.Elem()
	if pv.Kind() == reflect.Ptr {
		pv = pv.Elem()
	}
	pv.Set(v)
	return nil
}

func (p *proxyList) Append(ls ...lit.Lit) (lit.Appender, error) {
	v, ok := p.elem(reflect.Slice)
	if !ok {
		return nil, ErrNotSlice
	}
	rt := v.Type().Elem()
	for _, e := range ls {
		fp := reflect.New(rt)
		err := AssignToValue(e, fp)
		if err != nil {
			return nil, err
		}
		v = reflect.Append(v, fp.Elem())
	}
	res := *p
	res.val = reflect.New(v.Type())
	res.val.Set(v)
	return &res, nil
}

func (p *proxyList) Element() (lit.Proxy, error) {
	if v, ok := p.elem(reflect.Slice); ok {
		rt := v.Type().Elem()
		fp := reflect.New(rt)
		return ProxyValue(fp)
	}
	return nil, ErrNotSlice
}

func (p *proxyList) Len() int {
	if v, ok := p.elem(reflect.Slice); ok {
		return v.Len()
	}
	return 0
}
func (p *proxyList) IsZero() bool { return p.Len() == 0 }
func (p *proxyList) Idx(i int) (lit.Lit, error) {
	if v, ok := p.elem(reflect.Slice); ok {
		if i >= 0 && i < v.Len() {
			return ProxyValue(v.Index(i).Addr())
		}
	}
	return nil, lit.ErrIdxBounds
}
func (p *proxyList) SetIdx(i int, l lit.Lit) (lit.Indexer, error) {
	if v, ok := p.elem(reflect.Slice); ok {
		if i >= 0 && i < v.Len() {
			return p, AssignToValue(l, v.Index(i).Addr())
		}
	}
	return p, lit.ErrIdxBounds
}

func (p *proxyList) IterIdx(it func(int, lit.Lit) error) error {
	if v, ok := p.elem(reflect.Slice); ok {
		for i, n := 0, v.Len(); i < n; i++ {
			el, err := ProxyValue(v.Index(i).Addr())
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

func (p *proxyList) String() string               { return bfr.String(p) }
func (p *proxyList) MarshalJSON() ([]byte, error) { return bfr.JSON(p) }
func (p *proxyList) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('[')
	err := p.IterIdx(func(i int, el lit.Lit) error {
		if i > 0 {
			writeSep(b)
		}
		return writeLit(b, el)
	})
	if err != nil {
		return err
	}
	return b.WriteByte(']')
}

var _ lit.Appender = &proxyList{}
