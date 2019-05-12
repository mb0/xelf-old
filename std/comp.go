package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// eqSpec returns a bool whether the arguments are equivalent literals.
// The result is negated, if the expression symbol is 'ne'.
var eqSpec = core.impl("(form 'eq' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		return resolveBinaryComp(x, true, lit.Equiv)
	})

var neSpec = core.impl("(form 'ne' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		res, err := resolveBinaryComp(x, true, lit.Equiv)
		if err != nil {
			return res, err
		}
		return !res.(lit.Bool), nil
	})

// equalSpec returns a bool whether the arguments are same types or same literals.
var equalSpec = core.impl("(form 'equal' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		return resolveBinaryComp(x, true, lit.Equal)
	})

var inSpec = core.implResl("(form 'in' @1 list|@1 bool)",
	// @1 list|@1 : bool
	func(x exp.ReslReq) (exp.El, error) {
		return inOrNi(x, false)
	})

var niSpec = core.implResl("(form 'ni' @1 list|@1 bool)",
	// @1 list|@1 : bool
	func(x exp.ReslReq) (exp.El, error) {
		return inOrNi(x, true)
	})

func inOrNi(x exp.ReslReq, neg bool) (exp.El, error) {
	a := x.Arg(0).(lit.Lit)
	list, ok := x.Arg(1).(lit.Indexer)
	if !ok {
		return nil, cor.Errorf("expect idxer got %s", x.Arg(1).Typ())
	}
	var found bool
	err := list.IterIdx(func(idx int, el lit.Lit) error {
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
var ltSpec = core.impl("(form 'lt' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && res
		})
	})

var geSpec = core.impl("(form 'ge' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && !res
		})
	})

// specGt returns a bool whether the arguments are monotonic decreasing literals.
// Or the inverse, if the expression symbol is 'le'.
var gtSpec = core.impl("(form 'gt' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && res
		})
	})
var leSpec = core.impl("(form 'le' @1 :plain list|@1 bool)",
	func(x exp.ReslReq) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && !res
		})
	})

type cmpf = func(a, b lit.Lit) bool

func resolveBinaryComp(x exp.ReslReq, sym bool, cmp cmpf) (exp.El, error) {
	err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
	if err == exp.ErrUnres {
		if !x.Part {
			return x.Call, err
		}
	} else if err != nil {
		return nil, err
	}
	var res, init bool
	var unres []exp.El
	var last lit.Lit
	for _, args := range x.Layout.All() {
		for _, arg := range args {
			if arg.Typ().Kind&typ.KindAny == 0 {
				if len(unres) == 0 {
					unres = make([]exp.El, 0, len(x.Call.Args))
					if res {
						init = true
						unres = append(unres, last)
					}
				}
				res = false
				unres = append(unres, arg)
				continue
			}
			el := arg.(lit.Lit)
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
	}
	if len(unres) != 0 {
		x.Call.Args = unres
		return x.Call, exp.ErrUnres
	}
	return lit.True, nil
}
