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

type SpecImpl struct {
	resl Evaler
	eval Evaler
	part bool
}

func (r *SpecImpl) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.El, error) {
	req := CallCtx{p, env, c, h}
	if r.resl == nil {
		return r.eval(req)
	}
	res, err := r.resl(req)
	if err != nil {
		if r.part && err == exp.ErrUnres {
			return r.eval(req)
		}
		return res, err
	}
	return res, nil
}

func (r *SpecImpl) Eval(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.El, error) {
	req := CallCtx{p, env, c, h}
	if r.resl != nil {
		v, err := r.resl(req)
		if err != nil {
			if r.part && err == exp.ErrUnres {
				return r.eval(req)
			}
			return v, err
		}
	}
	return r.eval(req)
}

func SpecXX(sig string, x Evaler) *exp.Spec    { return newSpec(sig, nil, x, false) }
func SpecRX(sig string, r, x Evaler) *exp.Spec { return newSpec(sig, r, x, false) }
func SpecDX(sig string, x Evaler) *exp.Spec    { return newSpec(sig, DefaultResl, x, false) }
func SpecDXX(sig string, x Evaler) *exp.Spec   { return newSpec(sig, DefaultResl, x, true) }

func newSpec(sig string, r, x Evaler, part bool) *exp.Spec {
	s := exp.MustSig(sig)
	return &exp.Spec{s, &SpecImpl{r, x, part}}
}
