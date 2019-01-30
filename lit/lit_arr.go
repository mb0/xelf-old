package lit

import (
	"reflect"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// MakeArr returns a new abstract arr literal with the given type and len or an error.
func MakeArr(t typ.Type, len int) (Arr, error) {
	return MakeArrCap(t, len, len)
}

// MakeArrCap returns a new abstract arr literal with the given type, len and cap or an error.
func MakeArrCap(t typ.Type, len, cap int) (Arr, error) {
	if t.Kind&typ.MaskElem != typ.KindArr {
		return nil, typ.ErrInvalid
	}
	res := abstrArr{t.Next(), make(List, len, cap)}
	for i := range res.List {
		res.List[i] = Null(res.elem)
	}
	return &res, nil
}

type (
	abstrArr struct {
		elem typ.Type
		List
	}
	proxyArr struct{ proxy }
)

func (a abstrArr) Typ() typ.Type  { return typ.Arr(a.elem) }
func (a abstrArr) Elem() typ.Type { return a.elem }
func (a abstrArr) SetIdx(i int, el Lit) (err error) {
	if el == nil {
		el = Null(a.elem)
	} else {
		el, err = Convert(el, a.elem, 0)
		if err != nil {
			return err
		}
	}
	return a.List.SetIdx(i, el)
}

func (a abstrArr) Append(ls ...Lit) (Appender, error) {
	for _, e := range ls {
		e, err := Convert(e, a.elem, 0)
		if err != nil {
			return nil, err
		}
		a.List = append(a.List, e)
	}
	return a, nil
}

func (p *proxyArr) Assign(l Lit) error {
	if l == nil || !p.typ.Equal(l.Typ()) {
		return ErrNotAssignable
	}
	b, ok := l.(Idxer)
	if !ok || b.IsZero() { // a nil obj?
		p.val.Set(reflect.Zero(p.val.Type()))
		return nil
	}
	v, ok := p.elem(reflect.Slice)
	if !ok {
		return ErrNotAssignable
	}
	v = v.Slice(0, 0)
	err := b.IterIdx(func(i int, e Lit) error {
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

func (p *proxyArr) Append(ls ...Lit) (Appender, error) {
	v, ok := p.elem(reflect.Slice)
	if !ok {
		return nil, ErrNotAssignable
	}
	rt := v.Type().Elem()
	for _, e := range ls {
		if e == nil {
			return nil, ErrNotAssignable
		}
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

func (p *proxyArr) Elem() typ.Type { return p.typ.Next() }
func (p *proxyArr) Len() int {
	if v, ok := p.elem(reflect.Slice); ok {
		return v.Len()
	}
	return 0
}
func (p *proxyArr) IsZero() bool { return p.Len() == 0 }
func (p *proxyArr) Idx(i int) (Lit, error) {
	if v, ok := p.elem(reflect.Slice); ok {
		if i >= 0 && i < v.Len() {
			return ProxyValue(v.Index(i).Addr())
		}
	}
	return nil, ErrIdxBounds
}
func (p *proxyArr) SetIdx(i int, l Lit) error {
	if v, ok := p.elem(reflect.Slice); ok {
		if i >= 0 && i < v.Len() {
			return AssignToValue(l, v.Index(i).Addr())
		}
	}
	return ErrIdxBounds
}
func (p *proxyArr) IterIdx(it func(int, Lit) error) error {
	if v, ok := p.elem(reflect.Slice); ok {
		for i, n := 0, v.Len(); i < n; i++ {
			el, err := ProxyValue(v.Index(i).Addr())
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

func (p *proxyArr) String() string               { return bfr.String(p) }
func (p *proxyArr) MarshalJSON() ([]byte, error) { return bfr.JSON(p) }
func (p *proxyArr) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('[')
	err := p.IterIdx(func(i int, el Lit) error {
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
