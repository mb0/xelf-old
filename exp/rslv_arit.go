package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var ErrExpectNumer = cor.StrError("expected numer argument")

func opAdd(r, n float64) (float64, error) { return r + n, nil }
func opMul(r, n float64) (float64, error) { return r * n, nil }

// addSpec adds up all arguments and converts the sum to the first argument's type.
var addSpec = core.impl("(form 'add' :plain list|num : num)",
	// @1:num :plain? list|@2:num : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return reduceNums(c, env, e, 0, true, opAdd)
	})

// mulSpec multiplies all arguments and converts the product to the first argument's type.
var mulSpec = core.impl("(form 'mul' :plain list|num : num)",
	// @1:num :plain? list|@2:num : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		return reduceNums(c, env, e, 1, true, opMul)
	})

// subSpec subtracts the sum of the rest from the first argument and
// converts to the first argument's type.
var subSpec = core.impl("(form 'sub' :a num :b num :plain list|num : num)",
	// @1:num :plain list|@2 : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		fst, err := c.Resolve(env, e.Args[0], hint)
		if err == ErrUnres {
			if c.Part { // resolve the rest and return partial result
				rest := &Call{Spec: addSpec, Args: e.Args[1:]}
				sub, err := reduceNums(c, env, rest, 0, false, opAdd)
				if err == nil {
					e.Args = append(e.Args[:1], sub)
				} else if err == ErrUnres {
					e.Args = append(e.Args[:1], sub.(*Call).Args...)
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
	})

// divSpec divides the product of the rest from the first argument.
// If the first argument is an int div, integer division is used, otherwise it uses float division.
// The result is converted to the first argument's type.
var divSpec = core.impl("(form 'div' :a num :b num :plain list|num : num)",
	// @1:num :plain list|@2 : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		fst, err := c.Resolve(env, e.Args[0], hint)
		if err == ErrUnres {
			if c.Part { // resolve the rest and return partial result
				rest := &Call{Spec: mulSpec, Args: e.Args[1:]}
				sub, err := reduceNums(c, env, rest, 1, false, opMul)
				if err == nil {
					e.Args = append(e.Args[:1], sub)
				} else if err == ErrUnres {
					e.Args = append(e.Args[:1], sub.(*Call).Args...)
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
	})

// remSpec calculates the remainder of the first two arguments and always returns an int.
var remSpec = core.implResl("(form 'rem' :a int :b int : int)",
	// @1:int @1 : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		res := lo.Arg(0).(lit.Numeric).Num()
		mod := lo.Arg(1).(lit.Numeric).Num()
		return lit.Int(res) % lit.Int(mod), nil
	})

// absSpec returns the argument with the absolute numeric value.
var absSpec = core.implResl("(form 'abs' @1:num : @1)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (fst El, err error) {
		fst = lo.Arg(0)
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
		case lit.Numeric:
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
	})

// minSpec returns the argument with the smalles numeric value or an error.
var minSpec = core.impl("(form 'min' :a num :plain list|num : num)",
	// @1:num plain? list|@1 : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		var i int
		return reduceNums(c, env, e, 0, true, func(r, n float64) (float64, error) {
			if i++; i > 1 && r < n {
				return r, nil
			}
			return n, nil
		})
	})

// maxSpec returns the argument with the greatest numeric value or an error.
var maxSpec = core.impl("(form 'max' :a num :plain list|num : num)",
	// @1:num plain? list|@1 : @1
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		var i int
		return reduceNums(c, env, e, 0, true, func(r, n float64) (float64, error) {
			if i++; i > 1 && r > n {
				return r, nil
			}
			return n, nil
		})
	})

func convNumerType(e El, r lit.Numeric) (El, error) {
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

func getNumer(e El) lit.Numeric {
	v, _ := deopt(e).(lit.Numeric)
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

func reduceNums(c *Ctx, env Env, e *Call, res float64, conv bool, f numerReducer) (_ El, err error) {
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
		e.Args = unres
		return e, ErrUnres
	}
	return l, nil
}
