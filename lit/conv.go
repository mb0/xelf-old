package lit

import (
	"errors"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// ErrUnconv is the default conversion error.
var ErrUnconv = errors.New("cannot convert literal")

// Convert converts l to the dst type and returns the result or an error.
// Cmp is used if not CmpNone, otherwise Compare is called with the type of l and dst.
func Convert(l Lit, dst typ.Type, cmp typ.Cmp) (_ Lit, err error) {
	if l == nil {
		return nil, ErrUnconv
	}
	if cmp == typ.CmpNone {
		cmp = typ.Compare(l.Typ(), dst)
	}
	if cmp < typ.LvlCheck {
		return nil, ErrUnconv
	}
	if cmp&typ.BitUnwrap != 0 {
		o, ok := l.(Opter)
		if !ok {
			return nil, ErrUnconv
		}
		l, cmp = o.Some(), cmp&^typ.BitUnwrap
		if l == nil {
			t, _ := o.Typ().Deopt()
			l = Zero(t)
		}
	}
	switch cmp &^ typ.BitWrap {
	case typ.CmpSame, typ.CmpInfer:
	case typ.CmpCompAny:
		return Any{l}, nil
	case typ.CmpCompBase:
		l, err = convBase(l, dst)
	case typ.CmpCompList:
		return convList(l)
	case typ.CmpCompDict:
		return convDict(l)
	case typ.CmpCheckRef:
		return nil, typ.ErrInvalid
	case typ.CmpCheckAny:
		return checkAny(l, dst)
	case typ.CmpCompSpec, typ.CmpCheckSpec:
		l, err = checkSpec(l, dst)
	case typ.CmpCheckList:
		return checkList(l.(List), dst)
	case typ.CmpCheckDict:
		return checkDict(l.(*Dict), dst)
	case typ.CmpConvArr, typ.CmpCheckArr:
		return checkArr(l, dst)
	case typ.CmpConvMap, typ.CmpCheckMap:
		return checkMap(l, dst)
	case typ.CmpConvObj, typ.CmpCheckObj:
		l, err = checkObj(l, dst)
	}
	if err != nil {
		return nil, err
	}
	if cmp&typ.BitWrap != 0 {
		return Some{l}, nil
	}
	return l, nil
}

func convBase(l Lit, to typ.Type) (Lit, error) {
	switch to.Kind & typ.MaskBase {
	case typ.BaseNum:
		if v, ok := l.(Numer); ok {
			return Num(v.Num()), nil
		}
	case typ.BaseChar:
		if v, ok := l.(Charer); ok {
			return Char(v.Char()), nil
		}
	}
	return nil, ErrUnconv
}
func convList(l Lit) (Lit, error) {
	if v, ok := l.(Idxer); ok {
		res := make(List, v.Len())
		for i := range res {
			el, err := v.Idx(i)
			if err != nil {
				return nil, err
			}
			res[i] = el
		}
		return res, nil
	}
	return nil, ErrUnconv
}
func convDict(l Lit) (Lit, error) {
	if v, ok := l.(Keyer); ok {
		res := make([]Keyed, v.Len())
		for i, k := range v.Keys() {
			el, err := v.Key(k)
			if err != nil {
				return nil, err
			}
			res[i] = Keyed{k, el}
		}
		return &Dict{res}, nil
	}
	return nil, ErrUnconv
}
func checkAny(l Lit, to typ.Type) (Lit, error) {
	if v, ok := l.(Opter); ok {
		to, _ = to.Deopt()
		return Convert(v.Some(), to, 0)
	}
	return nil, ErrUnconv
}
func checkSpec(l Lit, to typ.Type) (res Lit, err error) {
	switch v := l.(type) {
	case Numer:
		n := v.Num()
		switch to.Kind & typ.MaskElem {
		case typ.KindBool:
			res = Bool(n != 0)
		case typ.KindInt:
			res = Int(n)
		case typ.KindReal:
			res = Real(n)
		case typ.KindSpan:
			res = Span(cor.MilliSpan(int64(n)))
		case typ.KindTime:
			res = Time(cor.UnixMilliTime(int64(n)))
		}
	case Charer:
		s := v.Char()
		switch to.Kind & typ.MaskElem {
		case typ.KindStr:
			res = Str(s)
		case typ.KindRaw:
			d, err := cor.ParseRaw(s)
			if err != nil {
				return nil, err
			}
			res = Raw(d)
		case typ.KindUUID:
			d, err := cor.ParseUUID(s)
			if err != nil {
				return nil, err
			}
			res = UUID(d)
		case typ.KindSpan:
			d, err := cor.ParseSpan(s)
			if err != nil {
				return nil, err
			}
			res = Span(d)
		case typ.KindTime:
			d, err := cor.ParseTime(s)
			if err != nil {
				return nil, err
			}
			res = Time(d)
		}
	}
	if res == nil {
		return nil, ErrUnconv
	}
	return res, nil
}
func checkList(l List, to typ.Type) (res Idxer, err error) {
	switch to.Kind & typ.MaskElem {
	case typ.KindArr:
		res, err = MakeArr(to, len(l))
	case typ.KindObj:
		res, err = MakeObj(to)
	default:
		return nil, ErrUnconv
	}
	if err != nil {
		return nil, err
	}
	for i, e := range l {
		err := res.SetIdx(i, e)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
func checkDict(l *Dict, to typ.Type) (res Keyer, err error) {
	switch to.Kind & typ.MaskElem {
	case typ.KindMap:
		res, err = MakeMapCap(to, l.Len())
	case typ.KindObj:
		res, err = MakeObj(to)
	default:
		return nil, ErrUnconv
	}
	if l != nil {
		for _, e := range l.List {
			err := res.SetKey(e.Key, e.Lit)
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}
func checkArr(l Lit, to typ.Type) (Arr, error) {
	if v, ok := l.(Arr); ok {
		res, err := MakeArr(to, v.Len())
		if err != nil {
			return nil, err
		}
		err = v.IterIdx(func(i int, e Lit) error {
			return res.SetIdx(i, e)
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, ErrUnconv
}
func checkMap(l Lit, to typ.Type) (Map, error) {
	if v, ok := l.(Map); ok {
		res, err := MakeMapCap(to, v.Len())
		if err != nil {
			return nil, err
		}
		err = v.IterKey(func(k string, e Lit) error {
			return res.SetKey(k, e)
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, ErrUnconv
}
func checkObj(l Lit, to typ.Type) (Obj, error) {
	if v, ok := l.(Obj); ok {
		res, err := MakeObj(to)
		if err != nil {
			return nil, err
		}
		err = v.IterKey(func(k string, e Lit) error {
			return res.SetKey(k, e)
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, ErrUnconv
}
