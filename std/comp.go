package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// eqSpec returns a bool whether the arguments are equivalent literals.
// The result is negated, if the expression symbol is 'ne'.
var eqSpec = core.add(SpecDXX("(form 'eq' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		return evalBinaryComp(x, true, lit.Equiv)
	}))

var neSpec = core.add(SpecDXX("(form 'ne' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		res, err := evalBinaryComp(x, true, lit.Equiv)
		if err != nil {
			return nil, err
		}
		a := res.(*exp.Atom)
		a.Lit = !a.Lit.(lit.Bool)
		return a, nil
	}))

// equalSpec returns a bool whether the arguments are same types or same literals.
var equalSpec = core.add(SpecDXX("(form 'equal' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		return evalBinaryComp(x, true, lit.Equal)
	}))

var inSpec = core.add(SpecDXX("(form 'in' @1 list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return inOrNi(x, false)
	}))

var niSpec = core.add(SpecDXX("(form 'ni' @1 list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return inOrNi(x, true)
	}))

func inOrNi(x CallCtx, neg bool) (exp.El, error) {
	err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
	if err != nil {
		return nil, err
	}
	a := x.Arg(0).(*exp.Atom)
	b := x.Arg(1).(*exp.Atom)
	list, ok := b.Lit.(lit.Indexer)
	if !ok {
		return nil, cor.Errorf("expect idxer got %s", b.Typ())
	}
	var found bool
	err = list.IterIdx(func(idx int, el lit.Lit) error {
		if found = lit.Equal(el, a.Lit); found {
			return lit.BreakIter
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if neg {
		found = !found
	}
	return &exp.Atom{Lit: lit.Bool(found)}, nil
}

// ltSpec returns a bool whether the arguments are monotonic increasing literals.
// Or the inverse, if the expression symbol is 'ge'.
var ltSpec = core.add(SpecDXX("(form 'lt' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		return evalBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && res
		})
	}))

var geSpec = core.add(SpecDXX("(form 'ge' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		return evalBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && !res
		})
	}))

// gtSpec returns a bool whether the arguments are monotonic decreasing literals.
// Or the inverse, if the expression symbol is 'le'.
var gtSpec = core.add(SpecDXX("(form 'gt' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		return evalBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && res
		})
	}))
var leSpec = core.add(SpecDXX("(form 'le' @ :plain list bool)",
	func(x CallCtx) (exp.El, error) {
		return evalBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && !res
		})
	}))

type cmpf = func(a, b lit.Lit) bool

func evalBinaryComp(x CallCtx, sym bool, cmp cmpf) (exp.El, error) {
	err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
	if x.Part && err == exp.ErrUnres {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	var res, init bool
	var unres []exp.El
	var last *exp.Atom
	for _, args := range x.Groups {
		for _, arg := range args {
			if arg.Typ().Kind&typ.KindAny == 0 {
				if len(unres) == 0 {
					unres = make([]exp.El, 0, 1+len(x.Groups[1]))
					if res {
						init = true
						unres = append(unres, last)
					}
				}
				res = false
				unres = append(unres, arg)
				continue
			}
			el := arg.(*exp.Atom)
			if last != nil {
				if !cmp(last.Lit, el.Lit) {
					return &exp.Atom{Lit: lit.False}, nil
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
		call := *x.Call
		call.Groups = [][]exp.El{unres[:1], unres[1:]}
		return &call, exp.ErrUnres
	}
	return &exp.Atom{Lit: lit.True}, nil
}
