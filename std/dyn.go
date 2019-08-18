package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var errConType = cor.StrError("the 'con' expression must start with a type")

// conSpec is a type conversion or constructor and must start with a type. It has four forms:
//    Without further arguments it returns the zero literal for that type.
//    With one literal compatible to that type it returns the converted literal.
//    For keyer types one or more declarations are set.
//    For idxer types one ore more literals are appended.
var conSpec = core.add(SpecRX("(form 'con' typ :plain? list :tags? dict @)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Resl(x.Prog, x.Env, x.Hint)
		if a, ok := x.Arg(0).(*exp.Atom); ok {
			if t, ok := a.Lit.(typ.Type); ok {
				switch t.Kind & typ.MaskElem {
				case typ.KindVoid:
					return exp.Ignore(x.Call.Src)
				case typ.KindBool:
					return x.BuiltinCall(x.Env, ":bool", x.Args(1), x.Call.Src)
				}
				ct := x.Sig
				r := &ct.Params[len(ct.Params)-1]
				r.Type = t
				if x.Hint != typ.Void {
					_, err := typ.Unify(x.Prog.Ctx, x.Hint, t)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		return x.Call, err
	},
	func(x CallCtx) (exp.El, error) {
		// resolve all arguments
		err := x.Layout.Eval(x.Prog, x.Env, x.Hint)
		if err != nil {
			return x.Call, err
		}
		t, ok := x.Arg(0).(*exp.Atom).Lit.(typ.Type)
		if !ok {
			return nil, errConType
		}
		switch t.Kind & typ.MaskElem {
		case typ.KindVoid:
			return exp.Ignore(x.Call.Src)
		case typ.KindBool:
			c, err := x.BuiltinCall(x.Env, ":bool", x.Args(1), x.Call.Src)
			if err != nil {
				return c, err
			}
			return c.Spec.Eval(x.Prog, x.Env, c, x.Hint)
		}
		args := x.Args(1)
		decls, err := x.Unis(2)
		if err != nil {
			return nil, err
		}
		// first rule: return zero literal
		if len(args) == 0 && len(decls) == 0 {
			return &exp.Atom{Lit: lit.Zero(t)}, nil
		}
		// second rule: convert compatible literals
		if len(args) == 1 && len(decls) == 0 {
			fst := args[0].(*exp.Atom)
			res, err := lit.Convert(fst.Lit, t, 0)
			if err == nil {
				return &exp.Atom{Lit: res}, nil
			}
		}
		// third rule: set declarations
		if t.Kind&typ.KindKeyr != 0 {
			res := deopt(lit.Zero(t)).(lit.Keyer)
			for _, d := range decls {
				el, ok := d.Arg().(*exp.Atom)
				if !ok {
					return nil, cor.Errorf("want literal in declaration got %s", d.El)
				}
				_, err = res.SetKey(d.Key(), el.Lit)
				if err != nil {
					return nil, err
				}
			}
			return &exp.Atom{Lit: res}, nil
		}
		// fourth rule: append list
		if ok && t.Kind&typ.KindIdxr != 0 {
			res := deopt(lit.Zero(t)).(lit.Indexer)
			apd, _ := res.(lit.Appender)
			for i, a := range args {
				el, ok := a.(*exp.Atom)
				if !ok {
					return nil, cor.Error("want literal in argument list")
				}
				if apd != nil { // list uses append
					apd, err = apd.Append(el.Lit)
				} else { // otherwise its a record literal set by index
					_, err = res.SetIdx(i, el.Lit)
				}
				if err != nil {
					return nil, err
				}
			}
			if apd != nil {
				return &exp.Atom{Lit: apd}, nil
			}
			return &exp.Atom{Lit: res}, nil
		}
		return nil, cor.Error("not implemented")
	}))
