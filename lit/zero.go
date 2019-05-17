package lit

import (
	"reflect"
	"time"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var (
	Nil      = Null(typ.Any)
	False    = Bool(false)
	True     = Bool(true)
	ZeroUUID = UUID([16]byte{})
	ZeroTime = Time(time.Time{})
	ZeroSpan = Span(0)
)

// Zero returns the zero literal for the given type t.
func Zero(t typ.Type) Lit {
	if t.Kind&typ.KindOpt != 0 {
		return Null(t)
	}
	switch t.Kind & typ.MaskRef {
	case typ.KindTyp:
		return typ.Void
	case typ.KindNum:
		return Num(0)
	case typ.KindBool:
		return False
	case typ.KindInt:
		return Int(0)
	case typ.KindReal:
		return Real(0)
	case typ.KindChar:
		return Char("")
	case typ.KindStr:
		return Str("")
	case typ.KindRaw:
		return Raw(nil)
	case typ.KindUUID:
		return ZeroUUID
	case typ.KindTime:
		return ZeroTime
	case typ.KindSpan:
		return ZeroSpan
	case typ.KindIdxr:
		return &List{}
	case typ.KindKeyr:
		return &Dict{}
	case typ.KindList:
		a, _ := MakeList(t, 0)
		return a
	case typ.KindDict:
		a, _ := MakeDict(t)
		return a
	case typ.KindRec, typ.KindObj:
		a, _ := MakeRec(t)
		return a
	}
	return Null(t)
}

// ZeroProxy returns an assignable zero literal for the given type t.
func ZeroProxy(tt typ.Type) (res Proxy) {
	t, opt := tt.Deopt()
	switch t.Kind & typ.MaskRef {
	case typ.KindTyp:
		res = TypProxy{&typ.Type{}}
	case typ.KindBool:
		res = new(Bool)
	case typ.KindInt:
		res = new(Int)
	case typ.KindReal:
		res = new(Real)
	case typ.KindStr:
		res = new(Str)
	case typ.KindRaw:
		res = new(Raw)
	case typ.KindUUID:
		res = new(UUID)
	case typ.KindTime:
		res = new(Time)
	case typ.KindSpan:
		res = new(Span)
	case typ.KindIdxr:
		res = &List{}
	case typ.KindKeyr:
		res = &Dict{}
	case typ.KindList:
		res, _ = MakeList(t, 0)
	case typ.KindDict:
		res, _ = MakeDict(t)
	case typ.KindRec, typ.KindObj:
		res, _ = MakeRec(t)
	}
	if res == nil {
		return &AnyProxy{reflect.ValueOf(new(interface{})), Nil}
	}
	if opt {
		return SomeProxy{res}
	}
	return res
}

type TypProxy struct {
	*typ.Type
}

func (p TypProxy) New() Proxy       { return TypProxy{} }
func (p TypProxy) Ptr() interface{} { return p.Type }
func (p TypProxy) Assign(l Lit) error {
	if t, ok := l.(typ.Type); ok {
		*p.Type = t
		return nil
	}
	return cor.Errorf("%q not assignable to %q", l.Typ(), typ.Typ)
}

type AnyProxy struct {
	Val reflect.Value
	Lit
}

func (p *AnyProxy) New() Proxy       { return &AnyProxy{Val: reflect.New(p.Val.Type().Elem())} }
func (p *AnyProxy) Ptr() interface{} { return p.Val.Interface() }
func (p *AnyProxy) Assign(l Lit) error {
	p.Lit = l
	var v interface{}
	switch x := l.(type) {
	case Character:
		v = x.Val()
	case Numeric:
		v = x.Val()
	case Proxy:
		v = x.Ptr()
		p.Val.Elem().Set(reflect.ValueOf(v).Elem())
		return nil
	default:
		v = x
	}
	p.Val.Elem().Set(reflect.ValueOf(v))
	return nil
}
