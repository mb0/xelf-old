package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// eqSpec returns a bool whether the arguments are equivalent literals.
// The result is negated, if the expression symbol is 'ne'.
var eqSpec = core.add(SpecXX("(form 'eq' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return resolveBinaryComp(x, true, lit.Equiv)
	}))

var neSpec = core.add(SpecXX("(form 'ne' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		res, err := resolveBinaryComp(x, true, lit.Equiv)
		if err != nil || !x.Exec {
			return res, err
		}
		a := res.(*exp.Atom)
		a.Lit = !a.Lit.(lit.Bool)
		return a, nil
	}))

// equalSpec returns a bool whether the arguments are same types or same literals.
var equalSpec = core.add(SpecXX("(form 'equal' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return resolveBinaryComp(x, true, lit.Equal)
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
	a := x.Arg(0).(*exp.Atom)
	b := x.Arg(1).(*exp.Atom)
	list, ok := b.Lit.(lit.Indexer)
	if !ok {
		return nil, cor.Errorf("expect idxer got %s", b.Typ())
	}
	var found bool
	err := list.IterIdx(func(idx int, el lit.Lit) error {
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
var ltSpec = core.add(SpecXX("(form 'lt' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && res
		})
	}))

var geSpec = core.add(SpecXX("(form 'ge' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(a, b)
			return ok && !res
		})
	}))

// gtSpec returns a bool whether the arguments are monotonic decreasing literals.
// Or the inverse, if the expression symbol is 'le'.
var gtSpec = core.add(SpecXX("(form 'gt' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && res
		})
	}))
var leSpec = core.add(SpecXX("(form 'le' @1 :plain list|@1 bool)",
	func(x CallCtx) (exp.El, error) {
		return resolveBinaryComp(x, false, func(a, b lit.Lit) bool {
			res, ok := lit.Less(b, a)
			return ok && !res
		})
	}))

type cmpf = func(a, b lit.Lit) bool

func resolveBinaryComp(x CallCtx, sym bool, cmp cmpf) (exp.El, error) {
	err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
	if x.Part && err == exp.ErrUnres {
		err = nil
	}
	if err != nil || (!x.Part && !x.Exec) {
		return x.Call, err
	}
	var res, init bool
	var unres []exp.El
	var last *exp.Atom
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
		x.Call.Args = unres
		return x.Call, exp.ErrUnres
	}
	return &exp.Atom{Lit: lit.True}, nil
}
