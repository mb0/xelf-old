package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// eqSpec returns a bool whether the arguments are equivalent literals.
// The result is negated, if the expression symbol is 'ne'.
var eqSpec = core.impl("(form 'eq' :a any :b any :plain : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return resolveBinaryComp(c, env, e, lo, true, lit.Equiv)
	})

var neSpec = core.impl("(form 'ne' :a any :b any :plain : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		res, err := resolveBinaryComp(c, env, e, lo, true, lit.Equiv)
		if err != nil {
			return res, err
		}
		return !res.(lit.Bool), nil
	})

// equalSpec returns a bool whether the arguments are same types or same literals.
var equalSpec = core.impl("(form 'equal' :a any :b any :plain : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return resolveBinaryComp(c, env, e, lo, true, lit.Equal)
	})

var inSpec = core.implResl("(form 'in' :a any :b list : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return inOrNi(c, env, e, lo, false)
	})

var niSpec = core.implResl("(form 'ni' :a any :b list : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return inOrNi(c, env, e, lo, true)
	})

func inOrNi(c *Ctx, env Env, e *Call, lo *Layout, neg bool) (El, error) {
	a := lo.Arg(0).(Lit)
	list, ok := lo.Arg(1).(lit.Indexer)
	if !ok {
		return nil, cor.Errorf("expect idxer got %s", lo.Arg(1).Typ())
	}
	var found bool
	err := list.IterIdx(func(idx int, el Lit) error {
		if found = lit.Equal(el, a); found {
			return lit.BreakIter
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if neg {
		return lit.Bool(!found), nil
	}
	return lit.Bool(found), nil
}

// ltSpec returns a bool whether the arguments are monotonic increasing literals.
// Or the inverse, if the expression symbol is 'ge'.
var ltSpec = core.impl("(form 'lt' :a any :b any :rest list : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return resolveBinaryComp(c, env, e, lo, false, func(a, b Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && res
		})
	})

var geSpec = core.impl("(form 'ge' :a any :b any :rest list : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return resolveBinaryComp(c, env, e, lo, false, func(a, b Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && !res
		})
	})

// specGt returns a bool whether the arguments are monotonic decreasing literals.
// Or the inverse, if the expression symbol is 'le'.
var gtSpec = core.impl("(form 'gt' :a any :b any :rest list : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return resolveBinaryComp(c, env, e, lo, false, func(a, b Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && res
		})
	})
var leSpec = core.impl("(form 'le' :a any :b any :rest list : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return resolveBinaryComp(c, env, e, lo, false, func(a, b Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && !res
		})
	})

type cmpf = func(a, b Lit) bool

func resolveBinaryComp(c *Ctx, env Env, e *Call, lo *Layout, sym bool, cmp cmpf) (El, error) {
	var res, init bool
	var unres []El
	var last Lit
	for _, arg := range e.Args {
		arg, err := c.Resolve(env, arg, typ.Void)
		if err == ErrUnres {
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
			if !cmp(last, el) {
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
