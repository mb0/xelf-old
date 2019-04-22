package lit

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// ErrUnconv is the default conversion error.
var ErrUnconv = cor.StrError("cannot convert literal")

// Convert converts l to the dst type and returns the result or an error.
// Cmp is used if not CmpNone, otherwise Compare is called with the type of l and dst.
func Convert(l Lit, dst typ.Type, cmp typ.Cmp) (_ Lit, err error) {
	if l == nil {
		return nil, cor.Errorf("%v, is nil %T", ErrUnconv, l)
	}
	if cmp == typ.CmpNone {
		cmp = typ.Compare(l.Typ(), dst)
	}
	if cmp < typ.LvlCheck {
		return nil, cor.Errorf("%v, incompatible %s to %s", ErrUnconv, l.Typ(), dst)
	}
	if cmp&typ.BitUnwrap != 0 {
		o, ok := l.(Opter)
		if !ok {
			return nil, cor.Errorf("%v, not an opter %T", ErrUnconv, l)
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
		return checkList(l.(Idxr), dst)
	case typ.CmpCheckDict:
		return checkDict(l.(*Keyr), dst)
	case typ.CmpConvArr, typ.CmpCheckArr:
		return checkArr(l, dst)
	case typ.CmpConvMap, typ.CmpCheckMap:
		return checkMap(l, dst)
	case typ.CmpConvRec, typ.CmpCheckRec:
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
		if v, ok := l.(Numeric); ok {
			return Num(v.Num()), nil
		}
	case typ.BaseChar:
		if v, ok := l.(Character); ok {
			return Char(v.Char()), nil
		}
	}
	return nil, cor.Errorf("%v %T to base", ErrUnconv, l)
}
func convList(l Lit) (Lit, error) {
	if v, ok := l.(Indexer); ok {
		res := make(Idxr, v.Len())
		for i := range res {
			el, err := v.Idx(i)
			if err != nil {
				return nil, err
			}
			res[i] = el
		}
		return res, nil
	}
	return nil, cor.Errorf("%v %T to list", ErrUnconv, l)
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
		return &Keyr{res}, nil
	}
	return nil, cor.Errorf("%v %T to dict", ErrUnconv, l)
}
func checkAny(l Lit, to typ.Type) (Lit, error) {
	if v, ok := l.(Opter); ok {
		vl := v.Some()
		if vl == nil {
			return Zero(to), nil
		}
		vt, opt := to.Deopt()
		vl, err := Convert(vl, vt, 0)
		if err != nil {
			return nil, err
		}
		if opt {
			vl = Some{vl}
		}
		return vl, nil
	}
	return nil, cor.Errorf("%v %T to %s", ErrUnconv, l, to)
}
func checkSpec(l Lit, to typ.Type) (res Lit, err error) {
	switch v := l.(type) {
	case Numeric:
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
	case Character:
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
		return nil, cor.Errorf("%v %T to %s", ErrUnconv, l, to)
	}
	return res, nil
}
func checkList(l Idxr, to typ.Type) (res Indexer, err error) {
	switch to.Kind & typ.MaskElem {
	case typ.KindList:
		res, err = MakeList(to, len(l))
	case typ.KindRec:
		res, err = MakeRec(to)
	default:
		return nil, cor.Errorf("%v %T to %s", ErrUnconv, l, to)
	}
	if err != nil {
		return nil, err
	}
	for i, e := range l {
		_, err := res.SetIdx(i, e)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
func checkDict(l *Keyr, to typ.Type) (res Keyer, err error) {
	switch to.Kind & typ.MaskElem {
	case typ.KindDict:
		res, err = MakeDictCap(to, l.Len())
	case typ.KindRec:
		res, err = MakeRec(to)
	default:
		return nil, cor.Errorf("%v %T to %s", ErrUnconv, l, to)
	}
	if l != nil {
		for _, e := range l.List {
			_, err := res.SetKey(e.Key, e.Lit)
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}
func checkArr(l Lit, to typ.Type) (Appender, error) {
	if v, ok := l.(Appender); ok {
		res, err := MakeList(to, v.Len())
		if err != nil {
			return nil, err
		}
		err = v.IterIdx(func(i int, e Lit) error {
			_, err := res.SetIdx(i, e)
			return err
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, cor.Errorf("%v %T to %s", ErrUnconv, l, to)
}
func checkMap(l Lit, to typ.Type) (Dictionary, error) {
	if v, ok := l.(Dictionary); ok {
		res, err := MakeDictCap(to, v.Len())
		if err != nil {
			return nil, err
		}
		err = v.IterKey(func(k string, e Lit) error {
			_, err := res.SetKey(k, e)
			return err
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, cor.Errorf("%v %T to %s", ErrUnconv, l, to)
}
func checkObj(l Lit, to typ.Type) (Lit, error) {
	if v, ok := l.(Opter); ok {
		l = v.Some()
		if l == nil {
			return Null(to), nil
		}
	}
	if v, ok := l.(Record); ok {
		res, err := MakeRec(to)
		if err != nil {
			return nil, err
		}
		err = v.IterKey(func(k string, e Lit) error {
			_, err := res.SetKey(k, e)
			return err
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, cor.Errorf("%v %s to %s", ErrUnconv, l.Typ(), to)
}
