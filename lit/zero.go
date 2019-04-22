package lit

import (
	"reflect"
	"time"

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
	if t.Kind&typ.FlagOpt != 0 {
		return Null(t)
	}
	switch t.Kind & typ.MaskRef {
	case typ.KindTyp:
		return typ.Void
	case typ.BaseNum:
		return Num(0)
	case typ.KindBool:
		return False
	case typ.KindInt:
		return Int(0)
	case typ.KindReal:
		return Real(0)
	case typ.BaseChar:
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
	case typ.BaseIdxr:
		return (Idxr)(nil)
	case typ.BaseKeyr:
		return &Keyr{}
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
func ZeroProxy(tt typ.Type) (res Assignable) {
	t, opt := tt.Deopt()
	switch t.Kind & typ.MaskRef {
	case typ.KindTyp:
		res = typProxy{&typ.Type{}}
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
	case typ.BaseIdxr:
		res = new(Idxr)
	case typ.BaseKeyr:
		res = &Keyr{}
	case typ.KindList:
		res, _ = MakeList(t, 0)
	case typ.KindDict:
		res, _ = MakeDict(t)
	case typ.KindRec, typ.KindObj:
		res, _ = MakeRec(t)
	}
	if res == nil {
		return &anyProxy{reflect.ValueOf(new(interface{})), Nil}
	}
	if opt {
		return SomeAssignable{res}
	}
	return res
}
