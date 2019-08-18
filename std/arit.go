package std

//*
import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var ErrExpectNumer = cor.StrError("expected numer argument")

func opAdd(r, n float64) (float64, error) { return r + n, nil }
func opMul(r, n float64) (float64, error) { return r * n, nil }

// addSpec adds up all arguments and converts the sum to the first argument's type.
var addSpec = core.add(SpecDXX("(form 'add' @1:num :plain list|num : @1)",
	func(x CallCtx) (exp.El, error) {
		return execNums(x, 0, opAdd)
	}))

// mulSpec multiplies all arguments and converts the product to the first argument's type.
var mulSpec = core.add(SpecDXX("(form 'mul' @1:num :plain list|num : @1)",
	func(x CallCtx) (exp.El, error) {
		return execNums(x, 1, opMul)
	}))

// subSpec subtracts the sum of the rest from the first argument and
// converts to the first argument's type.
var subSpec = core.add(SpecDXX("(form 'sub' @1:num :plain list|@:num : @1)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
		if err != nil {
			if err != exp.ErrUnres || !x.Part {
				return x.Call, err
			}
		}
		fst := x.Arg(0)
		rest := x.Args(1)
		n := getNumer(fst)
		ctx := numCtx{}
		if n == nil {
			ctx.idx = -1
		}
		err = redNums(rest, &ctx, opAdd)
		if err != nil {
			return nil, err
		}
		if n == nil {
			if ctx.idx >= 0 {
				ctx.unres[ctx.idx] = &exp.Atom{Lit: lit.Num(ctx.res)}
			}
			x.Groups[1] = ctx.unres
			return x.Call, exp.ErrUnres
		}
		var l lit.Lit = lit.Num(n.Num() - ctx.res)
		if fst.Typ() != typ.Num {
			l, err = lit.Convert(l, fst.Typ(), 0)
			if err != nil {
				return nil, err
			}
		}
		if len(ctx.unres) != 0 {
			x.Groups[0] = []exp.El{&exp.Atom{Lit: l, Src: fst.Source()}}
			x.Groups[1] = ctx.unres
			return x.Call, exp.ErrUnres
		}
		return &exp.Atom{Lit: l}, nil
	}))

// divSpec divides the product of the rest from the first argument.
// If the first argument is an int div, integer division is used, otherwise it uses float division.
// The result is converted to the first argument's type.
var divSpec = core.add(SpecDXX("(form 'div' @1:num :plain list|@:num : @1)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
		if err != nil {
			if err != exp.ErrUnres || !x.Part {
				return x.Call, err
			}
		}
		fst := x.Arg(0)
		n := getNumer(fst)
		ctx := numCtx{res: 1}
		if n == nil {
			ctx.idx = -1
		}
		err = redNums(x.Args(1), &ctx, opMul)
		if err != nil {
			return nil, err
		}
		if n == nil {
			if ctx.idx >= 0 {
				ctx.unres[ctx.idx] = &exp.Atom{Lit: lit.Num(ctx.res)}
			}
			x.Call.Groups[1] = ctx.unres
			return x.Call, exp.ErrUnres
		}
		if ctx.res == 0 {
			return nil, cor.Error("zero devision")
		}
		isint := fst.Typ().Kind&typ.MaskElem == typ.KindInt
		if isint {
			ctx.res = float64(int64(n.Num()) / int64(ctx.res))
		} else {
			ctx.res = n.Num() / ctx.res
		}
		var l lit.Lit = lit.Num(ctx.res)
		if fst.Typ() != typ.Num {
			l, err = lit.Convert(l, fst.Typ(), 0)
			if err != nil {
				return nil, err
			}
		}
		la := &exp.Atom{Lit: l, Src: fst.Source()}
		if len(ctx.unres) != 0 {
			x.Groups[0] = []exp.El{la}
			x.Groups[1] = ctx.unres
			return x.Call, exp.ErrUnres
		}
		return la, nil
	}))

// remSpec calculates the remainder of the first two arguments and always returns an int.
var remSpec = core.add(SpecDX("(form 'rem' @1:int @:int int)", func(x CallCtx) (exp.El, error) {
	err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
	if err != nil {
		return nil, err
	}

	res, aok := getNum(x.Arg(0))
	mod, bok := getNum(x.Arg(1))
	if !aok || !bok {
		return x.Call, exp.ErrUnres
	}
	return &exp.Atom{Lit: lit.Int(res.Num()) % lit.Int(mod.Num())}, nil
}))

func getNum(el exp.El) (lit.Numeric, bool) {
	if a, ok := el.(*exp.Atom); ok {
		if n, ok := a.Lit.(lit.Numeric); ok {
			return n, ok
		}
	}
	return nil, false
}

// absSpec returns the argument with the absolute numeric value.
var absSpec = core.add(SpecDX("(form 'abs' @1:num @1)", func(x CallCtx) (fst exp.El, err error) {
	return sign(x, false)
}))

// negSpec returns the argument with the negated numeric value.
var negSpec = core.add(SpecDX("(form 'neg' @1:num @1)", func(x CallCtx) (fst exp.El, err error) {
	return sign(x, true)
}))

func sign(x CallCtx, neg bool) (_ exp.El, err error) {
	err = x.Layout.Eval(x.Ctx, x.Env, x.Hint)
	if err != nil {
		return x.Call, err
	}

	fst := x.Arg(0).(*exp.Atom)
	switch v := fst.Lit.(type) {
	case lit.Int:
		if neg || v < 0 {
			fst.Lit = -v
		}
	case lit.Num:
		if neg || v < 0 {
			fst.Lit = -v
		}
	case lit.Real:
		if neg || v < 0 {
			fst.Lit = -v
		}
	case lit.Numeric:
		n := v.Num()
		if !neg && n >= 0 {
			break
		}
		nl := lit.Num(-n)
		if a, ok := v.(lit.Proxy); ok {
			err = a.Assign(nl)
			if err != nil {
				return nil, err
			}
		} else {
			fst.Lit, err = lit.Convert(nl, v.Typ(), 0)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, cor.Errorf("%v got %T", ErrExpectNumer, fst)
	}
	return fst, nil
}

// minSpec returns the argument with the smallest numeric value or an error.
var minSpec = core.add(SpecDX("(form 'min' @1:num :plain? list|@1 @1)",
	func(x CallCtx) (exp.El, error) {
		var i int
		return execNums(x, 0, func(r, n float64) (float64, error) {
			if i++; i > 0 && r < n {
				return r, nil
			}
			return n, nil
		})
	}))

// maxSpec returns the argument with the greatest numeric value or an error.
var maxSpec = core.add(SpecDX("(form 'max' @1:num :plain? list|@1 @1)",
	func(x CallCtx) (exp.El, error) {
		var i int
		return execNums(x, 0, func(r, n float64) (float64, error) {
			if i++; i > 0 && r > n {
				return r, nil
			}
			return n, nil
		})
	}))

func getNumer(e exp.El) lit.Numeric {
	if a, ok := e.(*exp.Atom); ok {
		v, _ := deopt(a.Lit).(lit.Numeric)
		return v
	}
	return nil
}

type numOp = func(r, e float64) (float64, error)

func deopt(l lit.Lit) lit.Lit {
	if o, ok := l.(lit.Opter); ok {
		if l = o.Some(); l == nil {
			t, _ := o.Typ().Deopt()
			l = lit.Zero(t)
		}
	}
	return l
}

func execNums(x CallCtx, res float64, f numOp) (exp.El, error) {
	err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
	if err != nil {
		if err != exp.ErrUnres || !x.Part {
			return x.Call, err
		}
	}
	part := err != nil
	ctx := numCtx{res: res, idx: -1}
	fst := x.Arg(0)
	if part {
		ctx.unres = []exp.El{fst}
	}
	n := getNumer(fst)
	if n != nil {
		ctx.idx = 0
		ctx.res = n.Num()
	}
	err = redNums(x.Args(1), &ctx, f)
	if err != nil {
		return nil, err
	}
	if len(ctx.unres) == 0 {
		var l lit.Lit = lit.Num(ctx.res)
		if fst.Typ() != typ.Num {
			l, err = lit.Convert(l, fst.Typ(), 0)
			if err != nil {
				return nil, err
			}
		}
		return &exp.Atom{Lit: l}, nil
	}
	if x.Part {
		if ctx.idx >= 0 && ctx.idx < len(ctx.unres) {
			ctx.unres[ctx.idx] = &exp.Atom{Lit: lit.Num(ctx.res)}
		}
		x.Groups[0] = ctx.unres[:1]
		x.Groups[1] = ctx.unres[1:]
	}
	return x.Call, exp.ErrUnres
}

type numCtx struct {
	res   float64
	idx   int
	unres []exp.El
}

func redNums(args []exp.El, c *numCtx, f numOp) (err error) {
	for _, arg := range args {
		v := getNumer(arg)
		if v == nil {
			c.unres = append(c.unres, arg)
			continue
		}
		if c.idx < 0 {
			c.idx = len(c.unres)
			c.unres = append(c.unres, arg)
		}
		c.res, err = f(c.res, v.Num())
		if err != nil {
			return err
		}
	}
	return nil
}
