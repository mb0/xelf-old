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
var addSpec = core.impl("(form 'add' @1:num :plain list|num : @1)",
	func(x exp.ReslReq) (exp.El, error) {
		return resNums(x, 0, opAdd)
	})

// mulSpec multiplies all arguments and converts the product to the first argument's type.
var mulSpec = core.impl("(form 'mul' @1:num :plain list|num : @1)",
	func(x exp.ReslReq) (exp.El, error) {
		return resNums(x, 1, opMul)
	})

// subSpec subtracts the sum of the rest from the first argument and
// converts to the first argument's type.
var subSpec = core.impl("(form 'sub' @1:num :plain list|@:num : @1)",
	func(x exp.ReslReq) (exp.El, error) {
		err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
		if err != nil {
			if err != exp.ErrUnres || !x.Part {
				return x.Call, err
			}
		}
		x.Call.Type = x.Layout.Sig
		fst := x.Layout.Arg(0)
		n := getNumer(fst)
		ctx := numCtx{}
		if n == nil {
			ctx.idx = -1
		}
		err = redNums(x.Args(1), &ctx, opAdd)
		if err != nil {
			return nil, err
		}
		if n == nil {
			if ctx.idx >= 0 {
				ctx.unres[ctx.idx] = &exp.Atom{Lit: lit.Num(ctx.res)}
			}
			x.Call.Args = append(x.Call.Args[:1], ctx.unres...)
			return x.Call, exp.ErrUnres
		}
		var l lit.Lit = lit.Num(n.Num() - ctx.res)
		if fst.Typ() != typ.Num {
			l, err = lit.Convert(l, fst.Typ(), 0)
			if err != nil {
				return nil, err
			}
		}
		la := &exp.Atom{Lit: l}
		if len(ctx.unres) != 0 {
			x.Call.Args = append(x.Call.Args[:0], la)
			x.Call.Args = append(x.Call.Args, ctx.unres...)
			return x.Call, exp.ErrUnres
		}
		return la, nil
	})

// divSpec divides the product of the rest from the first argument.
// If the first argument is an int div, integer division is used, otherwise it uses float division.
// The result is converted to the first argument's type.
var divSpec = core.impl("(form 'div' @1:num :plain list|@:num : @1)",
	func(x exp.ReslReq) (exp.El, error) {
		err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
		if err != nil {
			if err != exp.ErrUnres || !x.Part {
				return x.Call, err
			}
		}
		x.Call.Type = x.Sig
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
			x.Call.Args = append(x.Call.Args[:1], ctx.unres...)
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
		la := &exp.Atom{Lit: l}
		if len(ctx.unres) != 0 {
			x.Call.Args = append(x.Call.Args[:0], la)
			x.Call.Args = append(x.Call.Args, ctx.unres...)
			return x.Call, exp.ErrUnres
		}
		return la, nil
	})

// remSpec calculates the remainder of the first two arguments and always returns an int.
var remSpec = core.implResl("(form 'rem' @1:int @:int @1)",
	func(x exp.ReslReq) (exp.El, error) {
		res := x.Arg(0).(*exp.Atom).Lit.(lit.Numeric).Num()
		mod := x.Arg(1).(*exp.Atom).Lit.(lit.Numeric).Num()
		return &exp.Atom{Lit: lit.Int(res) % lit.Int(mod)}, nil
	})

// absSpec returns the argument with the absolute numeric value.
var absSpec = core.implResl("(form 'abs' @1:num @1)",
	func(x exp.ReslReq) (fst exp.El, err error) {
		return sign(x, false)
	})

// negSpec returns the argument with the negated numeric value.
var negSpec = core.implResl("(form 'neg' @1:num @1)",
	func(x exp.ReslReq) (fst exp.El, err error) {
		return sign(x, true)
	})

func sign(x exp.ReslReq, neg bool) (_ exp.El, err error) {
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

// minSpec returns the argument with the smalles numeric value or an error.
var minSpec = core.impl("(form 'min' @1:num :plain? list|@1 @1)",
	func(x exp.ReslReq) (exp.El, error) {
		var i int
		return resNums(x, 0, func(r, n float64) (float64, error) {
			if i++; i > 0 && r < n {
				return r, nil
			}
			return n, nil
		})
	})

// maxSpec returns the argument with the greatest numeric value or an error.
var maxSpec = core.impl("(form 'max' @1:num :plain? list|@1 @1)",
	// @1:num plain? list|@1 : @1
	func(x exp.ReslReq) (exp.El, error) {
		var i int
		return resNums(x, 0, func(r, n float64) (float64, error) {
			if i++; i > 0 && r > n {
				return r, nil
			}
			return n, nil
		})
	})

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

func resNums(x exp.ReslReq, res float64, f numOp) (exp.El, error) {
	err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
	if err != nil {
		if err != exp.ErrUnres || !x.Part {
			return x.Call, err
		}
	}
	x.Call.Type = x.Sig
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
	if ctx.idx >= 0 {
		ctx.unres[ctx.idx] = &exp.Atom{Lit: lit.Num(ctx.res)}
	}
	x.Call.Args = ctx.unres
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
