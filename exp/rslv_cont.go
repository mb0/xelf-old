package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

/*
Container operations

The len form returns the length of a str, raw, container literal, or the field count in an object
type.

The fst, lst and nth are a short-circuiting loops that optionally accept a iterator predicate and
return the first match from the start for fst, end for lst or the nth match from the end if the
given index is negative or the start otherwise:

The filter and map loops accept any container and an iterator function. The each loop resolves to
the given container, while filter returns a new container of the same type and map a new one with
another element type.

A iterator function's first parameter must accept the element type and can optionally be followed by
a int and str parameter for idx or key parameters. The key parameter can only be used for keyer
literals. Iterators for the each loop can returns anything as their result is ignored. The filter
and map loops iterator expects a literal of any type. Filter usually expects a boolean but falls
back on a zero check.

	(form iterator +val @el +idx? int +key? str - any)

The fold and foldr forms accumulate a container into a given literal. They accept any container and
a reducer function with a compatible accumulator parameter followed by iterator parameters. Fold
accumulates from first to last and foldr in reverse. Fold is technically a left fold and foldr a
right fold, but as the difference of cons lists and mostly linear xelf containers might lead to
confusion foldr should be thought of as reverse.

	(form accumulator +a @ +val @el +idx? int +key? str - @a)

The list, dict, arr and map constructor forms accept any container with an appropriate iterator
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

func init() {
	std.add("len", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly str raw list dict)
		{Type: typ.Int},
	}, rslvLen)
	std.add("fst", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "pred?"},
		{Type: typ.Infer},
	}, rslvFst)
	std.add("lst", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "pred?"},
		{Type: typ.Infer},
	}, rslvLst)
	std.add("nth", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "idx", Type: typ.Int},
		{Name: "pred?"},
		{Type: typ.Infer},
	}, rslvNth)
	std.add("filter", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "pred"},             // @iterator
		{Type: typ.Infer},          // @a
	}, rslvFilter)
	std.add("map", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "mut"},              // @iterator
		{Type: typ.Infer},          // @a
	}, rslvMap)
	std.add("fold", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "acc", Type: typ.Any},
		{Name: "ator"},    // @accumulator
		{Type: typ.Infer}, // @acc
	}, rslvFold)
	std.add("foldr", []typ.Param{
		{Name: "a", Type: typ.Any}, // (poly list dict)
		{Name: "acc", Type: typ.Any},
		{Name: "ator"},    // @accumulator
		{Type: typ.Infer}, // @acc
	}, rslvFoldr)
}

type litLener interface {
	Len() int
}

func rslvLen(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	fst := lo.Arg(0)
	if v, ok := deopt(fst).(litLener); ok {
		return lit.Int(v.Len()), nil
	}
	return nil, cor.Errorf("cannot call len on %s", fst.Typ())
}
func rslvFst(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	return nth(c, env, e, hint, lo.Arg(0), lo.Arg(1), 0)
}
func rslvLst(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	return nth(c, env, e, hint, lo.Arg(0), lo.Arg(1), -1)
}
func rslvNth(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	l, ok := lo.Arg(1).(lit.Numer)
	if !ok {
		return nil, cor.Errorf("want number got %s", lo.Arg(1))
	}
	return nth(c, env, e, hint, lo.Arg(0), lo.Arg(2), int(l.Num()))
}
func nth(c *Ctx, env Env, e *Call, hint Type, cont El, pred El, idx int) (_ El, err error) {
	if pred != nil {
		iter, err := getIter(c, env, pred, cont.Typ(), false)
		if err != nil {
			return nil, err
		}
		cont, err = iter.filter(c, env, cont)
		if err != nil {
			return nil, err
		}
	}
	switch v := deopt(cont).(type) {
	case lit.Idxer:
		idx, err = checkIdx(idx, v.Len())
		if err != nil {
			return nil, err
		}
		return v.Idx(idx)
	case *lit.Dict:
		idx, err = checkIdx(idx, v.Len())
		if err != nil {
			return nil, err
		}
		keyed := v.List[idx]
		return keyed.Lit, nil
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
	*Spec
	n, a, v, i, k int
	args          []El
	ator          bool
}

func getIter(c *Ctx, env Env, e El, ct Type, ator bool) (r *fIter, _ error) {
	e, err := Resolve(env, e)
	if err != nil && err != ErrUnres {
		return nil, err
	}
	switch t := e.Typ(); t.Kind {
	case typ.ExpFunc, typ.ExpForm:
		r = &fIter{Spec: e.(*Spec)}
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
		switch args[r.n].Type.Kind {
		case typ.KindInt:
			if r.i > 0 {
			}
			r.i = r.n
			r.n++
		case typ.KindStr:
			if r.k > 0 {
			}
			r.k = r.n
			r.n++
		default:
			return nil, cor.Errorf("unexpected parameter %s", args[r.n])
		}
	}
	r.args = make([]El, r.n)
	return r, nil
}

func (r *fIter) resolve(c *Ctx, env Env, el El, idx int, key string) (Lit, error) {
	r.args[0] = el
	if r.i > 0 {
		r.args[r.i] = lit.Int(idx)
	}
	if r.k > 0 {
		r.args[r.k] = lit.Str(key)
	}
	call := &Call{Def: DefSpec(r.Spec), Args: r.args}
	res, err := r.ResolveCall(c, env, call, typ.Void)
	if err != nil {
		return nil, err
	}
	return res.(Lit), nil
}
func (r *fIter) accumulate(c *Ctx, env Env, acc, el El, idx int, key string) (Lit, error) {
	r.args[0] = acc
	if r.v > 0 {
		r.args[r.v] = el
	}
	if r.i > 0 {
		r.args[r.i] = lit.Int(idx)
	}
	if r.k > 0 {
		r.args[r.k] = lit.Str(key)
	}
	call := &Call{Def: DefSpec(r.Spec), Args: r.args}
	res, err := r.ResolveCall(c, env, call, typ.Void)
	if err != nil {
		return nil, cor.Errorf("accumulate: %w", err)
	}
	return res.(Lit), nil
}

func (r *fIter) filter(c *Ctx, env Env, cont El) (Lit, error) {
	switch v := deopt(cont).(type) {
	case lit.Keyer:
		out := lit.Zero(v.Typ()).(lit.Keyer)
		idx := 0
		err := v.IterKey(func(key string, el Lit) error {
			res, err := r.resolve(c, env, el, idx, key)
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
	case lit.Idxer:
		if r.k > 0 {
			return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
		}
		out := lit.Zero(v.Typ()).(lit.Appender)
		err := v.IterIdx(func(idx int, el Lit) error {
			res, err := r.resolve(c, env, el, idx, "")
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

func rslvFilter(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	cont := lo.Arg(0)
	iter, err := getIter(c, env, lo.Arg(1), cont.Typ(), false)
	if err != nil {
		return nil, err
	}
	res, err := iter.filter(c, env, cont)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func rslvMap(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	cont := lo.Arg(0)
	iter, err := getIter(c, env, lo.Arg(1), cont.Typ(), false)
	if err != nil {
		return nil, err
	}
	var rt Type
	it := iter.Res()
	if it == typ.Void || it == typ.Infer {
		it = typ.Any
	}
	switch t := cont.Typ(); t.Kind & typ.MaskElem {
	case typ.BaseList:
		if it == typ.Any {
			rt = typ.List
		} else {
			rt = typ.Arr(it)
		}
	case typ.KindArr:
		rt = typ.Arr(it)
	case typ.BaseDict:
		if it == typ.Any {
			rt = typ.Dict
		} else {
			rt = typ.Map(it)
		}
	case typ.KindMap:
		rt = typ.Map(it)
	case typ.KindObj:
		rt = typ.Dict
	}
	switch v := deopt(cont).(type) {
	case lit.Keyer:
		out := lit.Zero(rt).(lit.Keyer)
		idx := 0
		err := v.IterKey(func(key string, el Lit) error {
			res, err := iter.resolve(c, env, el, idx, key)
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
		return out, nil
	case lit.Idxer:
		out := lit.Zero(rt).(lit.Appender)
		if iter.k > 0 {
			return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
		}
		err := v.IterIdx(func(idx int, el Lit) error {
			res, err := iter.resolve(c, env, el, idx, "")
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
		return out, nil
	}
	return nil, cor.Errorf("map requires idxer or keyer got %s", cont.Typ())
}
func rslvFold(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	cont := lo.Arg(0)
	acc := lo.Arg(1).(Lit)
	iter, err := getIter(c, env, lo.Arg(2), acc.Typ(), true)
	if err != nil {
		return nil, err
	}
	switch v := deopt(cont).(type) {
	case lit.Keyer:
		idx := 0
		err := v.IterKey(func(key string, el Lit) error {
			acc, err = iter.accumulate(c, env, acc, el, idx, key)
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
	case lit.Idxer:
		if iter.k > 0 {
			return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
		}
		err := v.IterIdx(func(idx int, el Lit) error {
			acc, err = iter.accumulate(c, env, acc, el, idx, "")
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
}
func rslvFoldr(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		return e, err
	}
	cont := lo.Arg(0)
	acc := lo.Arg(1).(Lit)
	iter, err := getIter(c, env, lo.Arg(2), acc.Typ(), true)
	if err != nil {
		return nil, err
	}
	switch v := deopt(cont).(type) {
	case lit.Keyer:
		keys := v.Keys()
		for idx := len(keys) - 1; idx >= 0; idx-- {
			key := keys[idx]
			el, err := v.Key(key)
			if err != nil {
				return nil, err
			}
			acc, err = iter.accumulate(c, env, acc, el, idx, key)
			if err != nil {
				return nil, err
			}
		}
		return acc, nil
	case lit.Idxer:
		if iter.k > 0 {
			return nil, cor.Errorf("iter key parameter for idxer %s", cont.Typ())
		}
		ln := v.Len()
		for idx := ln - 1; idx >= 0; idx-- {
			el, err := v.Idx(idx)
			if err != nil {
				return nil, err
			}
			acc, err = iter.accumulate(c, env, acc, el, idx, "")
			if err != nil {
				return nil, err
			}
		}
		return acc, nil
	}
	return nil, cor.Errorf("fold requires idxer or keyer got %s", cont.Typ())
}
