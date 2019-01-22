package lit

import (
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
	}
	return Null(t)
}
