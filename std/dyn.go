package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var errConType = cor.StrError("the 'con' expression must start with a type")

// dynSpec resolves a dynamic expressions. If the first element resolves to a type it is
// resolves as the 'con' expression. If it is a literal it selects an appropriate combine
// expression for that literal. The time and uuid literals have no such combine expression.
var dynSpec = core.add(SpecX("(form 'dyn' @1 :rest? list @2)",
	func(x CallCtx) (_ exp.El, err error) {
		if len(x.Call.Args) == 0 {
			return &exp.Atom{Lit: typ.Void}, nil
		}
		return defaultDyn(x.Ctx, x.Env, &exp.Dyn{Els: x.Call.Args}, x.Hint)
	}))

func defaultDyn(c *exp.Ctx, env exp.Env, d *exp.Dyn, hint typ.Type) (_ exp.El, err error) {
	if len(d.Els) == 0 {
		return &exp.Atom{Lit: typ.Void}, nil
	}
	fst := d.Els[0]
	switch t := fst.Typ(); t.Kind {
	case typ.KindSym, typ.KindCall:
		fst, err = c.Resolve(env, fst, typ.Void)
	case typ.KindDyn:
		v, _ := fst.(*exp.Dyn)
		if len(v.Els) == 0 {
			return &exp.Atom{Lit: typ.Void}, nil
		}
		fst, err = defaultDyn(c, env, v, typ.Void)
	}
	if err != nil {
		return d, err
	}
	var spec *exp.Spec
	var sym string
	args := d.Els
	switch t := fst.Typ(); t.Kind & typ.MaskRef {
	case typ.KindTyp:
		a := fst.(*exp.Atom)
		tt := a.Lit.(typ.Type)
		if tt == typ.Void {
			return fst, nil
		}
		if tt == typ.Bool {
			spec, args = boolSpec, args[1:]
		} else {
			sym = "con"
		}
	case typ.KindFunc, typ.KindForm:
		a := fst.(*exp.Atom)
		r, ok := a.Lit.(*exp.Spec)
		if ok {
			spec, args = r, args[1:]
		}
	default:
		if len(d.Els) == 1 && t.Kind&typ.KindAny != 0 {
			return fst, nil
		}
		switch t.Kind & typ.MaskElem {
		case typ.KindBool:
			sym = "and"
		case typ.KindNum, typ.KindInt, typ.KindReal, typ.KindSpan:
			sym = "add"
		case typ.KindChar, typ.KindStr, typ.KindRaw:
			sym = "cat"
		case typ.KindIdxr, typ.KindList:
			sym = "apd" // TODO maybe cat
		case typ.KindKeyr, typ.KindDict, typ.KindRec:
			sym = "set" // TODO maybe merge
		}
	}
	if sym != "" {
		def := exp.LookupSupports(env, sym, '~')
		if def != nil {
			spec, _ = def.Lit.(*exp.Spec)
		}
	}
	if spec != nil {
		t := c.Inst(spec.Type)
		return spec.Resolve(c, env, &exp.Call{Spec: spec, Type: t, Args: args}, hint)
	}
	return nil, cor.Errorf("unexpected first argument %[1]T %[1]s in dynamic expression\n%s %s",
		fst, sym, fst.Typ())
}

// conSpec is a type conversion or constructor and must start with a type. It has four forms:
//    Without further arguments it returns the zero literal for that type.
//    With one literal compatible to that type it returns the converted literal.
//    For keyer types one or more declarations are set.
//    For idxer types one ore more literals are appended.
var conSpec = core.add(SpecRX("(form 'con' typ :plain? list :tags? dict @)",
	func(x CallCtx) (exp.El, error) {
		err := x.Layout.Resolve(x.Ctx.WithExec(false), x.Env, x.Hint)
		if a, ok := x.Arg(0).(*exp.Atom); ok {
			if t, ok := a.Lit.(typ.Type); ok {
				if x.Hint != typ.Void {
					_, err := typ.Unify(x.Ctx.Ctx, x.Hint, t)
					if err != nil {
						return nil, err
					}
				}
				x.Call.Type = x.Apply(x.Call.Type)
			}
		}
		return x.Call, err
	},
	func(x CallCtx) (exp.El, error) {
		// resolve all arguments
		err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
		if err != nil {
			return x.Call, err
		}
		t, ok := x.Arg(0).(*exp.Atom).Lit.(typ.Type)
		if !ok {
			return nil, errConType
		}
		if t == typ.Void { // just in case we have a dynamic comment
			return &exp.Atom{Lit: typ.Void}, nil
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
