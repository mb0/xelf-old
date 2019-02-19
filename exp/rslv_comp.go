package exp

import (
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// rslvEq returns a bool whether the arguments are equivalent literals.
// The result is negated, if the expression symbol is 'ne'.
func rslvEq(c *Ctx, env Env, e *Expr) (El, error) {
	neg := e.Name == "ne"
	res, err := resolveBinaryComp(c, env, e, false, true, lit.Equiv)
	if !neg || err != nil {
		return res, err
	}
	return !res.(lit.Bool), nil
}

// rslvEqual returns a bool whether the arguments are same types or same literals.
func rslvEqual(c *Ctx, env Env, e *Expr) (El, error) {
	return resolveBinaryComp(c, env, e, false, true, lit.Equal)
}

// rslvLt returns a bool whether the arguments are monotonic increasing literals.
// Or the inverse, if the expression symbol is 'ge'.
func rslvLt(c *Ctx, env Env, e *Expr) (El, error) {
	return resolveBinaryComp(c, env, e, e.Name == "ge", false, func(a, b Lit) bool {
		res, ok := lit.Less(a, b)
		return ok && res
	})
}

// rslvGt returns a bool whether the arguments are monotonic decreasing literals.
// Or the inverse, if the expression symbol is 'le'.
func rslvGt(c *Ctx, env Env, e *Expr) (El, error) {
	return resolveBinaryComp(c, env, e, e.Name == "le", false, func(a, b Lit) bool {
		res, ok := lit.Less(b, a)
		return ok && res
	})
}

type cmpf = func(a, b Lit) bool

func resolveBinaryComp(c *Ctx, env Env, e *Expr, neg, sym bool, cmp cmpf) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	var res, init bool
	var unres []El
	var last Lit
	for _, arg := range e.Args {
		arg, err = c.Resolve(env, arg)
		if err == ErrUnres {
			e.Type = typ.Bool
			if !c.Part {
				return e, err
			}
			if len(unres) == 0 {
				unres = make([]El, 0, len(e.Args))
				if res {
					init = true
					unres = append(unres, last)
				}
			}
			res = false
			unres = append(unres, arg)
			continue
		}
		if err != nil {
			return nil, err
		}
		el := arg.(Lit)
		if last != nil {
			if neg == cmp(last, el) {
				return lit.False, nil
			}
		}
		if !res && ((!sym || !init) && len(unres) > 0) || len(unres) == 1 {
			unres = append(unres, el)
		}
		last = el
		res = true
	}
	if len(unres) != 0 {
		e.Args = unres
		return e, ErrUnres
	}
	return lit.True, nil
}
