package exp

import (
	"fmt"

	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
	"github.com/pkg/errors"
)

var ErrExpectNumer = errors.New("expected numer argument")

// rslvAdd tries to add up all arguments and converts the sum to the first argument's type.
func rslvAdd(c *Ctx, env Env, e *Expr) (El, error) {
	if len(e.Args) == 0 {
		return lit.Num(0), nil
	}
	args, err := resolveArgs(c, env, e.Args)
	if err != nil {
		return e, err
	}
	res, err := reduceNumers(args, 0, func(r, n float64) float64 { return r + n })
	if err != nil {
		return nil, err
	}
	return convNumerType(args[0], res)
}

// rslvMul tries to multiply all arguments and converts the product to the first argument's type.
func rslvMul(c *Ctx, env Env, e *Expr) (El, error) {
	if len(e.Args) == 0 {
		return lit.Num(1), nil
	}
	args, err := resolveArgs(c, env, e.Args)
	if err != nil {
		return e, err
	}
	res, err := reduceNumers(args, 1, func(r, n float64) float64 { return r * n })
	if err != nil {
		return nil, err
	}
	return convNumerType(args[0], res)
}

// rslvSub tries to subtract the sum of the rest from the first argument and
// converts to the first argument's type.
func rslvSub(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
	if err != nil {
		return e, err
	}
	res := getNumer(args[0])
	if res == nil {
		return nil, err
	}
	sub, err := reduceNumers(args[1:], 0, func(r, n float64) float64 { return r + n })
	if err != nil {
		return nil, err
	}
	return convNumerType(args[0], lit.Num(res.Num())-sub)
}

// rslvDiv tries to divide the product of the rest from the first argument.
// If the first argument is an int div, integer division is used, otherwise it uses float division.
// The result is converted to the first argument's type.
func rslvDiv(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
	if err != nil {
		return e, err
	}
	res := getNumer(args[0])
	if res == nil {
		return nil, ErrExpectNumer
	}
	div, err := reduceNumers(args[1:], 1, func(r, n float64) float64 { return r * n })
	if err != nil {
		return nil, err
	}
	if div.IsZero() {
		return nil, errors.Errorf("zero devision")
	}
	if res.Typ().Kind&typ.MaskElem == typ.KindInt {
		res = lit.Int(res.Num()) / lit.Int(div)
	} else {
		res = lit.Num(res.Num()) / div
	}
	return convNumerType(args[0], res)
}

// rslvRem tries to calculate the remainder of the first two arguments and always returns an int.
func rslvRem(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsExact(e.Args, 2)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
	if err != nil {
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

func reduceNumers(es []El, res float64, f numerReducer) (lit.Num, error) {
	for _, arg := range es {
		v := getNumer(arg)
		if v == nil {
			return 0, fmt.Errorf("expected numer argument got %T", arg)
		}
		res = f(res, v.Num())
	}
	return lit.Num(res), nil
}

type numerReducer = func(r, e float64) float64

func resolveArgs(c *Ctx, env Env, es []El) ([]El, error) {
	err := ArgsForm(es)
	if err != nil {
		return nil, err
	}
	return c.ResolveAll(env, es)
}

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
