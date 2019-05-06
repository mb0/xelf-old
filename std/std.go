package std

import (
	"github.com/mb0/xelf/exp"
)

var Std = exp.Builtin{Core, Decl}

var core = make(formMap, 32)
var decl = make(formMap, 8)

func Core(sym string) *exp.Spec {
	if f, ok := core[sym]; ok {
		return f
	}
	return nil
}

func Decl(sym string) *exp.Spec {
	if f, ok := decl[sym]; ok {
		return f
	}
	return nil
}

type formMap map[string]*exp.Spec

func (m formMap) impl(sig string, r exp.ReslReqFunc) *exp.Spec {
	f := exp.ImplementReq(sig, false, r)
	m[f.Ref] = f
	return f
}

func (m formMap) implResl(sig string, r exp.ReslReqFunc) *exp.Spec {
	f := exp.ImplementReq(sig, true, r)
	m[f.Ref] = f
	return f
}
