// Package std provides built-in expression resolvers.
package std

import "github.com/mb0/xelf/exp"

// Std is an environment that includes both the Core and Decl lookup function.
var Std = exp.Builtin{Core, Decl}

var core = make(formMap, 32)
var decl = make(formMap, 8)

// Core is a resolver lookup for all standard forms, not involving declarations or functions.
//
// Logic forms:
//    and, or, bool, not, if
// Arithmetic forms:
//    add, mul, sub, div, rem, abs, neg, min, max
// Comparison forms:
//    eq, ne, equal, in, ni, lt, le, gt, ge
// Other forms:
//    len, with, dyn, con, cat, apd, set
func Core(sym string) *exp.Spec {
	if f, ok := core[sym]; ok {
		return f
	}
	return nil
}

// Decl is a resolver lookup for the declaration and container forms.
//
// Declaration forms:
//   let and fn
// Container forms:
//    fst, lst, nth, filter, map, fold, foldr
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
