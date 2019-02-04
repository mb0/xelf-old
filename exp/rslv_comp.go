package exp

import (
	"github.com/mb0/xelf/lit"
)

// rslvEq returns a bool whether the arguments are same types or equivalent literals.
// The result is negated, if the expression symbol is 'ne'.
func rslvEq(c *Ctx, env Env, e *Expr) (El, error) {
	neg := e.Name == "ne"
	res, err := resolveBinaryComp(c, env, e, false, func(a, b El) bool {
		switch v := a.(type) {
		case Lit:
			w, ok := b.(Lit)
			return ok && lit.Equiv(v, w)
		}
		return false
	})
	if !neg || err != nil {
		return res, err
	}
	return !res.(lit.Bool), nil
}

// rslvEqual returns a bool whether the arguments are same types or same literals.
func rslvEqual(c *Ctx, env Env, e *Expr) (El, error) {
	return resolveBinaryComp(c, env, e, false, func(a, b El) bool {
		switch v := a.(type) {
		case Lit:
			w, ok := b.(Lit)
			return ok && lit.Equal(v, w)
		}
		return false
	})
}

// rslvLt returns a bool whether the arguments are monotonic increasing literals.
// Or the inverse, if the expression symbol is 'ge'.
func rslvLt(c *Ctx, env Env, e *Expr) (El, error) {
	return resolveBinaryComp(c, env, e, e.Name == "ge", func(a, b El) bool {
		switch v := a.(type) {
		case Lit:
			if w, ok := b.(Lit); ok {
				res, ok := lit.Less(v, w)
				return ok && res
			}
		}
		return false
	})
}

// rslvGt returns a bool whether the arguments are monotonic decreasing literals.
// Or the inverse, if the expression symbol is 'le'.
func rslvGt(c *Ctx, env Env, e *Expr) (El, error) {
	return resolveBinaryComp(c, env, e, e.Name == "le", func(a, b El) bool {
		switch v := a.(type) {
		case Lit:
			if w, ok := b.(Lit); ok {
				res, ok := lit.Less(w, v)
				return ok && res
			}
		}
		return false
	})
}

func resolveBinaryComp(c *Ctx, env Env, e *Expr, neg bool, cmp func(a, b El) bool) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
	if err != nil {
		return e, err
	}
	last := args[0]
	for i := 1; i < len(args); i++ {
		el := args[i]
		if neg == cmp(last, el) {
			return lit.False, nil
		}
		last = el
	}
	return lit.True, nil
}
