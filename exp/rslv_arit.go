package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var ErrExpectNumer = cor.StrError("expected numer argument")

func init() {
	core.add("add", typ.Num, nil, rslvAdd)
	core.add("mul", typ.Num, nil, rslvMul)
	core.add("sub", typ.Num, nil, rslvSub)
	core.add("div", typ.Num, nil, rslvDiv)
	core.add("rem", typ.Int, nil, rslvRem)
	core.add("abs", typ.Num, nil, rslvAbs)
	core.add("min", typ.Num, nil, rslvMin)
	core.add("max", typ.Num, nil, rslvMax)
}

func opAdd(r, n float64) (float64, error) { return r + n, nil }
func opMul(r, n float64) (float64, error) { return r * n, nil }

// rslvAdd adds up all arguments and converts the sum to the first argument's type.
// (form +args? arr|num - num)
func rslvAdd(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	return reduceNums(c, env, e, 0, true, opAdd)
}

// rslvMul multiplies all arguments and converts the product to the first argument's type.
// (form +args? arr|num - num)
func rslvMul(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	return reduceNums(c, env, e, 1, true, opMul)
}

// rslvSub subtracts the sum of the rest from the first argument and
// converts to the first argument's type.
// (form +a num +rest arr|num - num)
func rslvSub(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	fst, err := c.Resolve(env, e.Args[0], hint)
	if err == ErrUnres {
		if c.Part { // resolve the rest and return partial result
			rest := &Expr{Ref{Name: "add"}, e.Args[1:],
				ExprResolverFunc(rslvAdd)}
			sub, err := reduceNums(c, env, rest, 0, false, opAdd)
			if err == nil {
				e.Args = append(e.Args[:1], sub)
			} else if err == ErrUnres {
				e.Args = append(e.Args[:1], sub.(*Expr).Args...)
			} else {
				return nil, err
			}
		}
		return e, ErrUnres
	}
	if err != nil {
		return nil, err
	}
	num := getNumer(fst)
	if num == nil {
		return nil, ErrExpectNumer
	}
	return reduceNums(c, env, e, 2*num.Num(), true, func(r, n float64) (float64, error) {
		return r - n, nil
	})
}

// rslvDiv divides the product of the rest from the first argument.
// If the first argument is an int div, integer division is used, otherwise it uses float division.
// The result is converted to the first argument's type.
// (form +a num +rest arr|num - num)
func rslvDiv(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	fst, err := c.Resolve(env, e.Args[0], hint)
	if err == ErrUnres {
		if c.Part { // resolve the rest and return partial result
			rest := &Expr{Ref{Name: "mul"}, e.Args[1:],
				ExprResolverFunc(rslvMul)}
			sub, err := reduceNums(c, env, rest, 1, false, opMul)
			if err == nil {
				e.Args = append(e.Args[:1], sub)
			} else if err == ErrUnres {
				e.Args = append(e.Args[:1], sub.(*Expr).Args...)
			} else {
				return nil, err
			}
		}
		return e, ErrUnres
	}
	if err != nil {
		return nil, err
	}
	num := getNumer(fst)
	if num == nil {
		return nil, ErrExpectNumer
	}
	var i int
	isInt := num.Typ().Kind&typ.MaskElem == typ.KindInt
	return reduceNums(c, env, e, 0, true, func(r, n float64) (float64, error) {
		if i++; i == 1 {
			return n, nil
		}
		if n == 0 {
			return 0, cor.Error("zero devision")
		}
		if isInt {
			return float64(int64(r) / int64(n)), nil
		}
		return r / n, nil
	})
}

// rslvRem calculates the remainder of the first two arguments and always returns an int.
// (form +a +b num - int)
func rslvRem(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsExact(e.Args, 2)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args, typ.Int)
	if err != nil {
		e.Type = typ.Int
		return e, err
	}
	res := getNumer(args[0])
	if res == nil {
		return nil, ErrExpectNumer
	}
	mod := getNumer(args[1])
	if mod == nil {
		return nil, ErrExpectNumer
	}
	return lit.Int(res.Num()) % lit.Int(mod.Num()), nil
}

// rslvAbs returns the argument with the absolute numeric value.
// (form +a num - num)
func rslvAbs(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsExact(e.Args, 1)
	if err != nil {
		return nil, err
	}
	fst, err := c.Resolve(env, e.Args[0], hint)
	if err != nil {
		if err != ErrUnres {
			return nil, err
		}
		return e, err
	}
	switch v := fst.(type) {
	case lit.Int:
		if v < 0 {
			fst = -v
		}
	case lit.Num:
		if v < 0 {
			fst = -v
		}
	case lit.Real:
		if v < 0 {
			fst = -v
		}
	case lit.Numer:
		n := v.Num()
		if n >= 0 {
			break
		}
		nl := lit.Num(-n)
		if a, ok := v.(lit.Assignable); ok {
			err = a.Assign(nl)
			if err != nil {
				return nil, err
			}
		} else {
			fst, err = lit.Convert(nl, v.Typ(), 0)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, cor.Errorf("%v got %T", ErrExpectNumer, fst)
	}
	return fst, nil
}

// rslvMin returns the argument with the smalles numeric value or an error.
// (form +a num +rest? arr|num - num)
func rslvMin(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsMin(e.Args, 1)
	if err != nil {
		return nil, err
	}
	var i int
	return reduceNums(c, env, e, 0, true, func(r, n float64) (float64, error) {
		if i++; i > 1 && r < n {
			return r, nil
		}
		return n, nil
	})
}

// rslvMax returns the argument with the greatest numeric value or an error.
// (form +a num +rest? arr|num - num)
func rslvMax(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsMin(e.Args, 1)
	if err != nil {
		return nil, err
	}
	var i int
	return reduceNums(c, env, e, 0, true, func(r, n float64) (float64, error) {
		if i++; i > 1 && r > n {
			return r, nil
		}
		return n, nil
	})
}

func convNumerType(e El, r lit.Numer) (El, error) {
	l, ok := e.(Lit)
	if ok {
		n, err := lit.Convert(r, l.Typ(), 0)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	return nil, cor.Errorf("%v got %T", ErrExpectNumer, e)
}

func getNumer(e El) lit.Numer {
	v, _ := deopt(e).(lit.Numer)
	return v
}

type numerReducer = func(r, e float64) (float64, error)

func deopt(el El) Lit {
	if l, ok := el.(Lit); ok {
		if o, ok := l.(lit.Opter); ok {
			if l = o.Some(); l == nil {
				t, _ := o.Typ().Deopt()
				l = lit.Zero(t)
			}
		}
		return l
	}
	return nil
}

func reduceNums(c *Ctx, env Env, e *Expr, res float64, conv bool, f numerReducer) (_ El, err error) {
	t := typ.Void
	var resed int
	var unres []El
	for idx, el := range e.Args {
		el, err = c.Resolve(env, el, typ.Num)
		if err == ErrUnres {
			if c.Part {
				if len(unres) == 0 {
					unres = make([]El, 0, len(e.Args))
					if idx > 0 {
						unres = append(unres, nil)
					}
				}
				unres = append(unres, el)
				continue
			}
			return e, err
		}
		if err != nil {
			return nil, err
		}
		v := getNumer(el)
		if v == nil {
			return nil, cor.Errorf("%v got %T", ErrExpectNumer, el)
		}
		if idx == 0 && conv {
			t = el.(Lit).Typ()
		}
		res, err = f(res, v.Num())
		if err != nil {
			return nil, err
		}
		resed++
	}
	var l Lit
	l = lit.Num(res)
	if t != typ.Void && t != typ.Num {
		l, err = lit.Convert(l, t, 0)
		if err != nil {
			return nil, err
		}
	}
	if len(unres) > 0 {
		if unres[0] == nil {
			unres[0] = l
		} else if resed > 0 {
			unres = append(unres, l)
		}
		if t != typ.Void {
			e.Type = t
		}
		e.Args = unres
		return e, ErrUnres
	}
	return l, nil
}
