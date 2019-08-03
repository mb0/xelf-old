// Package std provides built-in expression resolvers.
package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/typ"
)

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

func (m formMap) add(s *exp.Spec) *exp.Spec {
	m[s.Ref] = s
	return s
}

type CallCtx struct {
	*exp.Ctx
	Env  exp.Env
	Call *exp.Call
	*exp.Layout
	Hint typ.Type
}
type Evaler func(CallCtx) (exp.El, error)

func DefaultResl(x CallCtx) (exp.El, error) {
	err := x.Layout.Resolve(x.Ctx.WithExec(false), x.Env, x.Hint)
	x.Call.Type = x.Layout.Sig
	return x.Call, err
}

func SpecDX(sig string, x Evaler) *exp.Spec     { return FullSpec(sig, DefaultResl, x, nil) }
func SpecDXX(sig string, x Evaler) *exp.Spec    { return FullSpec(sig, DefaultResl, x, x) }
func SpecX(sig string, x Evaler) *exp.Spec      { return FullSpec(sig, nil, x, nil) }
func SpecXX(sig string, x Evaler) *exp.Spec     { return FullSpec(sig, nil, x, x) }
func SpecRX(sig string, r, x Evaler) *exp.Spec  { return FullSpec(sig, r, x, nil) }
func SpecRXX(sig string, r, x Evaler) *exp.Spec { return FullSpec(sig, r, x, x) }

func FullSpec(sig string, r, x, p Evaler) *exp.Spec {
	s := exp.MustSig(sig)
	return &exp.Spec{s, exp.ReslFunc(
		func(c *exp.Ctx, env exp.Env, e *exp.Call, hint typ.Type) (exp.El, error) {
			if e.Type == typ.Void {
				return nil, cor.Errorf("type not instantiated for %s %s", s, e.Type)
			}
			lo, err := exp.LayoutArgs(e.Type, e.Args)
			if err != nil {
				return nil, err
			}
			req := CallCtx{c, env, e, lo, hint}
			if r != nil {
				_, err = r(req)
				if err != nil {
					if p != nil && c.Part && err == exp.ErrUnres {
						return p(req)
					}
					return e, err
				}
			}
			if r == nil || c.Exec {
				return x(req)
			}
			return e, nil
		},
	)}
}
