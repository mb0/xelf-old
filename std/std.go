// Package std provides built-in expression resolvers.
package std

import (
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
	*exp.Prog
	Env exp.Env
	*exp.Call
	Hint typ.Type
}
type Evaler func(CallCtx) (exp.El, error)

func DefaultResl(x CallCtx) (exp.El, error) {
	err := x.Layout.Resl(x.Prog, x.Env, x.Hint)
	return x.Call, err
}

type ReslRXP struct {
	R, X, P Evaler
}

func (r ReslRXP) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.El, error) {
	req := CallCtx{p, env, c, h}
	if r.R == nil {
		return r.X(req)
	}
	res, err := r.R(req)
	if err != nil {
		if r.P != nil && err == exp.ErrUnres {
			return r.P(req)
		}
		return res, err
	}
	return res, nil
}

func (r ReslRXP) Eval(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.El, error) {
	req := CallCtx{p, env, c, h}
	if r.R != nil {
		v, err := r.R(req)
		if err != nil {
			if r.P != nil && err == exp.ErrUnres {
				return r.P(req)
			}
			return v, err
		}
	}
	return r.X(req)
}

func SpecXX(sig string, x Evaler) *exp.Spec    { return Impl(sig, ReslRXP{x, x, nil}) }
func SpecRX(sig string, r, x Evaler) *exp.Spec { return Impl(sig, ReslRXP{r, x, nil}) }
func SpecDX(sig string, x Evaler) *exp.Spec    { return Impl(sig, ReslRXP{DefaultResl, x, nil}) }
func SpecDXX(sig string, x Evaler) *exp.Spec   { return Impl(sig, ReslRXP{DefaultResl, x, x}) }

func Impl(sig string, rxp ReslRXP) *exp.Spec {
	s := exp.MustSig(sig)
	return &exp.Spec{s, rxp}
}
