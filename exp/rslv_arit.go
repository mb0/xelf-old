package exp

import (
	"fmt"

	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
	"github.com/pkg/errors"
)

var ErrExpectNumer = errors.New("expected numer argument")

func opAdd(r, n float64) (float64, error) { return r + n, nil }
func opMul(r, n float64) (float64, error) { return r * n, nil }

// rslvAdd tries to add up all arguments and converts the sum to the first argument's type.
func rslvAdd(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	return reduceNums(c, env, e, 0, true, opAdd)
}

// rslvMul tries to multiply all arguments and converts the product to the first argument's type.
func rslvMul(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	return reduceNums(c, env, e, 1, true, opMul)
}

// rslvSub tries to subtract the sum of the rest from the first argument and
// converts to the first argument's type.
func rslvSub(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	fst, err := c.Resolve(env, e.Args[0])
	if err == ErrUnres {
		if c.Part { // resolve the rest and return partial result
			rest := &Expr{Sym: Sym{Name: "add"}, Args: e.Args[1:]}
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

// rslvDiv tries to divide the product of the rest from the first argument.
// If the first argument is an int div, integer division is used, otherwise it uses float division.
// The result is converted to the first argument's type.
func rslvDiv(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	fst, err := c.Resolve(env, e.Args[0])
	if err == ErrUnres {
		if c.Part { // resolve the rest and return partial result
			rest := &Expr{Sym: Sym{Name: "mul"}, Args: e.Args[1:]}
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
			return 0, errors.Errorf("zero devision")
		}
		if isInt {
			return float64(int64(r) / int64(n)), nil
		}
		return r / n, nil
	})
}

// rslvRem tries to calculate the remainder of the first two arguments and always returns an int.
func rslvRem(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsExact(e.Args, 2)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
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

func convNumerType(e El, r lit.Numer) (El, error) {
	l, ok := e.(Lit)
	if ok {
		n, err := lit.Convert(r, l.Typ(), 0)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	return nil, fmt.Errorf("%v got %T", ErrExpectNumer, e)
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
		el, err = c.Resolve(env, el)
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
			return nil, fmt.Errorf("%v got %T", ErrExpectNumer, el)
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
