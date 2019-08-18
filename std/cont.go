package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

/*
Container operations

The len form returns the length of a str, raw, container literal, or the field count of a record.

The fst, lst and nth are a short-circuiting loops that optionally accept a predicate and return the
first match from the start for fst, end for lst or the nth match from the start if the given index
is positive or from the end otherwise:

The filter and map loops accept any container and an predicate or mapper function. The each loop
resolves to the given container, while filter returns a new container of the same type and map a new
one with another element type.

A predicate or mapper function's first parameter must accept the element type and can optionally
be followed by a int and str parameter for idx or key parameters. The key parameter can only be used
for keyer literals. The filter predicate must return bool and mapper a literal of any type.

	(form 'pred' :val @1 :idx? int :key? str bool)
	(form 'mapr' :val @1 :idx? int :key? str @2)

The fold and foldr forms accumulate a container into a given literal. They accept any container and
a reducer function with a compatible accumulator parameter followed by iterator parameters. Fold
accumulates from first to last and foldr in reverse. Fold is technically a left fold and foldr a
right fold, but as the difference of cons lists and mostly linear xelf containers might lead to
confusion foldr should be thought of as reverse.

	(form 'accu' :a @1 :val @2 :idx? int :key? str @1)

The list, dict constructor forms accept any container with an appropriate iterator
to construct a new container literal by effectively using each or foldr.

(with [1 2 3 4 5] +even (fn (eq (rem _ 2) 0)) (and
	(eq (len "test") 4)
	(eq (len .) 5)
	(eq (fst .) (nth . 0) 1)
	(eq (lst .) (nth . -1) 5)
	(eq (fst . even) 2)
	(eq (lst . even) 4)
	(eq (nth . 1 even) 4)
	(eq (nth . -2 even) 4)
	(eq (filter . even) [2 4])
	(eq (map . even) [false true false true false])
	(eq (fold . 0 (fn (add _ .val))) 15)
	(eq (fold . [0] (fn (apd _ .val))) [0 1 2 3 4 5])
	(eq (foldr . [0] (fn (apd _ .val))) [0 5 4 3 2 1])
))
*/

type litLener interface {
	Len() int
}

var lenSpec = core.add(SpecDX("(form 'len' (@:alt cont str raw) int)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
		if err != nil {
			return nil, err
		}
		fst := x.Arg(0).(*exp.Atom)
		if v, ok := deopt(fst.Lit).(litLener); ok {
			return &exp.Atom{lit.Int(v.Len()), x.Source()}, nil
		}
		return nil, cor.Errorf("cannot call len on %s", fst.Typ())
	}))

var fstSpec = decl.add(SpecDX("(form 'fst' list|@1 :pred? (func @ bool) @2)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
		if err != nil {
			return nil, err
		}
		return nth(x, x.Arg(0).(*exp.Atom), x.Arg(1), 0)
	}))

var lstSpec = decl.add(SpecDX("(form 'lst' list|@1 :pred? (func @1 bool) @2)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
		if err != nil {
			return nil, err
		}
		return nth(x, x.Arg(0).(*exp.Atom), x.Arg(1), -1)
	}))

var nthSpec = decl.add(SpecDX("(form 'nth' cont|@1 int :pred? (func @1 bool) @2)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
		if err != nil {
			return nil, err
		}
		l, ok := x.Arg(1).(*exp.Atom).Lit.(lit.Numeric)
		if !ok {
			return nil, cor.Errorf("want number got %s", x.Arg(1))
		}
		return nth(x, x.Arg(0).(*exp.Atom), x.Arg(2), int(l.Num()))
	}))

func nth(x CallCtx, cont *exp.Atom, pred exp.El, idx int) (_ exp.El, err error) {
	if pred != nil {
		iter, err := getIter(x, pred, cont.Typ(), false)
		if err != nil {
			return nil, err
		}
		cont.Lit, err = iter.filter(x, cont)
		if err != nil {
			return nil, err
		}
	}
	switch v := deopt(cont.Lit).(type) {
	case lit.Indexer:
		idx, err = checkIdx(idx, v.Len())
		if err != nil {
			return nil, err
		}
		l, err := v.Idx(idx)
		if err != nil {
			return nil, err
		}
		return &exp.Atom{Lit: l}, nil
	case *lit.Dict:
		idx, err = checkIdx(idx, v.Len())
		if err != nil {
			return nil, err
		}
		keyed := v.List[idx]
		return &exp.Atom{Lit: keyed.Lit}, nil
	}
	return nil, cor.Errorf("nth wants idxer or dict got %s", cont.Typ())
}
func checkIdx(idx, l int) (int, error) {
	if idx < 0 {
		idx = l + idx
	}
	if idx < 0 || idx >= l {
		return idx, lit.ErrIdxBounds
	}
	return idx, nil
}

type fIter struct {
	*exp.Spec
	n, a, v, i, k int
	args          []exp.El
	ator          bool
}

func getIter(x CallCtx, e exp.El, ct typ.Type, ator bool) (r *fIter, _ error) {
	e, err := x.Ctx.Resl(x.Env, e, typ.Void)
	if err != nil && err != exp.ErrUnres {
		return nil, err
	}
	if a, ok := e.(*exp.Atom); ok {
		if s, ok := a.Lit.(*exp.Spec); ok {
			r = &fIter{Spec: s}
		}
	}
	if r == nil {
		return nil, cor.Errorf("iter not a func or form %s", e.Typ())
	}
	r.ator = ator
	args := r.Arg()
	if len(args) == 0 {
		return nil, cor.Errorf("iter must have at least one argument %s", e.Typ())
	}
	r.n = 1
	if ator {
		r.v = 1
		ct = args[0].Typ()
		r.n++
		if len(args) == 1 {
			return nil, cor.Errorf("ator must have at least two arguments %s", e.Typ())
		}
	}
	fst := args[r.v]
	switch fst.Name { // unless the parameter name is explicitly idx or key we assume val
	case "idx", "key":
		// TODO handle explicit first param
		return nil, cor.Errorf("key and idx iter without value are not implemented")
	}
	if !ator {
		cmp := typ.Compare(ct.Elem(), fst.Type)
		if cmp < typ.LvlCheck {
			return nil, cor.Errorf("iter value %s cannot be used as %s", ct.Elem(), fst.Type)
		}
	}
	for r.n < len(args) && r.n < r.v+3 {
		switch args[r.n].Type.Kind { // default to idx
		case typ.KindStr:
			if r.k > 0 {
				return nil, cor.Errorf("key parameter already set, got %d %s",
					r.n, args[r.n])
			}
			r.k = r.n
			r.n++
		default:
			if r.i > 0 {
				return nil, cor.Errorf("idx parameter already set, got %d %s",
					r.n, args[r.n])
			}
			r.i = r.n
			r.n++
		}
	}
	r.args = make([]exp.El, r.n)
	return r, nil
}

func (r *fIter) eval(x CallCtx, el lit.Lit, idx int, key string) (lit.Lit, error) {
	r.args[0] = &exp.Atom{Lit: el}
	if r.i > 0 {
		r.args[r.i] = &exp.Atom{Lit: lit.Int(idx)}
	}
	if r.k > 0 {
		r.args[r.k] = &exp.Atom{Lit: lit.Str(key)}
	}
	call, err := x.NewCall(r.Spec, r.args, x.Src)
	if err != nil {
		return nil, err
	}
	res, err := x.Ctx.Eval(x.Env, call, typ.Void)
	if err != nil {
		return nil, err
	}
	return res.(*exp.Atom).Lit, nil
}
func (r *fIter) accumulate(x CallCtx, acc *exp.Atom, el lit.Lit, idx int, key string) (*exp.Atom, error) {
	r.args[0] = acc
	if r.v > 0 {
		r.args[r.v] = &exp.Atom{Lit: el}
	}
	if r.i > 0 {
		r.args[r.i] = &exp.Atom{Lit: lit.Int(idx)}
	}
	if r.k > 0 {
		r.args[r.k] = &exp.Atom{Lit: lit.Str(key)}
	}
	call, err := x.NewCall(r.Spec, r.args, x.Src)
	if err != nil {
		return nil, err
	}
	res, err := x.WithPart(false).Eval(x.Env, call, typ.Void)
	if err != nil {
		return nil, err
	}
	return res.(*exp.Atom), nil
}

func (r *fIter) filter(x CallCtx, cont *exp.Atom) (lit.Lit, error) {
	switch v := deopt(cont.Lit).(type) {
	case lit.Keyer:
		out := lit.Zero(v.Typ()).(lit.Keyer)
		idx := 0
		err := v.IterKey(func(key string, el lit.Lit) error {
			res, err := r.eval(x, el, idx, key)
			if err != nil {
				return err
			}
			if !res.IsZero() {
				out.SetKey(key, el)
			}
			idx++
			return nil
		})
		if err != nil {
			return nil, err
		}
		return out, nil
	case lit.Indexer:
		if r.k > 0 {
			return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
		}
		out := lit.Zero(v.Typ()).(lit.Appender)
		err := v.IterIdx(func(idx int, el lit.Lit) error {
			res, err := r.eval(x, el, idx, "")
			if err != nil {
				return err
			}
			if !res.IsZero() {
				out, err = out.Append(el)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	return nil, cor.Errorf("filter requires idxer or keyer got %s", cont.Typ())
}

var filterSpec = decl.add(SpecDX("(form 'filter' cont|@1 (func @1 bool) @2)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
		if err != nil {
			return nil, err
		}
		cont := x.Arg(0).(*exp.Atom)
		iter, err := getIter(x, x.Arg(1), cont.Typ(), false)
		if err != nil {
			return nil, err
		}
		res, err := iter.filter(x, cont)
		if err != nil {
			return nil, err
		}
		return &exp.Atom{res, x.Src}, nil
	}))

var mapSpec = decl.add(SpecDX("(form 'map' cont|@1 (func @1 @2) @3)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, typ.Void)
		if err != nil {
			return nil, err
		}
		cont := x.Arg(0).(*exp.Atom)
		iter, err := getIter(x, x.Arg(1), cont.Typ(), false)
		if err != nil {
			return nil, err
		}
		var rt typ.Type
		it := iter.Res()
		if it == typ.Void {
			it = typ.Any
		}
		switch t := cont.Typ(); t.Kind & typ.MaskElem {
		case typ.KindIdxr:
			if it == typ.Any {
				rt = typ.Idxr(it)
			} else {
				rt = typ.List(it)
			}
		case typ.KindList:
			rt = typ.List(it)
		case typ.KindKeyr:
			if it == typ.Any {
				rt = typ.Keyr(it)
			} else {
				rt = typ.Dict(it)
			}
		case typ.KindDict:
			rt = typ.Dict(it)
		case typ.KindRec:
			rt = typ.Keyr(it)
		}
		switch v := deopt(cont.Lit).(type) {
		case lit.Keyer:
			out := lit.Zero(rt).(lit.Keyer)
			idx := 0
			err := v.IterKey(func(key string, el lit.Lit) error {
				res, err := iter.eval(x, el, idx, key)
				if err != nil {
					return err
				}
				_, err = out.SetKey(key, res)
				if err != nil {
					return err
				}
				idx++
				return nil
			})
			if err != nil {
				return nil, err
			}
			return &exp.Atom{out, x.Src}, nil
		case lit.Indexer:
			out := lit.Zero(rt).(lit.Appender)
			if iter.k > 0 {
				return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
			}
			err := v.IterIdx(func(idx int, el lit.Lit) error {
				res, err := iter.eval(x, el, idx, "")
				if err != nil {
					return err
				}
				out, err = out.Append(res)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			return &exp.Atom{out, x.Src}, nil
		}
		return nil, cor.Errorf("map requires idxer or keyer got %s", cont.Typ())
	}))

var foldSpec = decl.add(SpecDX("(form 'fold' cont|@1 @2 (func @2 @1 @2) @2)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
		if err != nil {
			return nil, err
		}
		cont := x.Arg(0).(*exp.Atom)
		acc := x.Arg(1).(*exp.Atom)
		iter, err := getIter(x, x.Arg(2), acc.Typ(), true)
		if err != nil {
			return nil, err
		}
		switch v := deopt(cont.Lit).(type) {
		case lit.Keyer:
			idx := 0
			err := v.IterKey(func(key string, el lit.Lit) error {
				acc, err = iter.accumulate(x, acc, el, idx, key)
				if err != nil {
					return err
				}
				idx++
				return nil
			})
			if err != nil {
				return nil, err
			}
			return acc, nil
		case lit.Indexer:
			if iter.k > 0 {
				return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
			}
			err := v.IterIdx(func(idx int, el lit.Lit) error {
				acc, err = iter.accumulate(x, acc, el, idx, "")
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			return acc, nil
		}
		return nil, cor.Errorf("fold requires idxer or keyer got %s", cont.Typ())
	}))

var foldrSpec = decl.add(SpecDX("(form 'foldr' cont|@1 @2 (func @2 @1 @2) @2)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Eval(x.Ctx, x.Env, x.Hint)
		if err != nil {
			return nil, err
		}
		cont := x.Arg(0).(*exp.Atom)
		acc := x.Arg(1).(*exp.Atom)
		iter, err := getIter(x, x.Arg(2), acc.Typ(), true)
		if err != nil {
			return nil, err
		}
		switch v := deopt(cont.Lit).(type) {
		case lit.Keyer:
			keys := v.Keys()
			for idx := len(keys) - 1; idx >= 0; idx-- {
				key := keys[idx]
				el, err := v.Key(key)
				if err != nil {
					return nil, err
				}
				acc, err = iter.accumulate(x, acc, el, idx, key)
				if err != nil {
					return nil, err
				}
			}
			return acc, nil
		case lit.Indexer:
			if iter.k > 0 {
				return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
			}
			ln := v.Len()
			for idx := ln - 1; idx >= 0; idx-- {
				el, err := v.Idx(idx)
				if err != nil {
					return nil, err
				}
				acc, err = iter.accumulate(x, acc, el, idx, "")
				if err != nil {
					return nil, err
				}
			}
			return acc, nil
		}
		return nil, cor.Errorf("fold requires idxer or keyer got %s", cont.Typ())
	}))
