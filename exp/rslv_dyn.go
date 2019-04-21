package exp

import (
	"log"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var errAsType = cor.StrError("the 'as' expression must start with a type")

var formAs *Spec

func init() {
	core.add("dyn", []typ.Param{
		{Name: "args"},
		{Name: "unis"},
		{Type: typ.Infer},
	}, rslvDyn)
	formAs = core.add("as", []typ.Param{
		{Name: "t", Type: typ.Typ},
		{Name: "args", Type: typ.List},
		{Name: "unis", Type: typ.Dict},
		{Type: typ.Infer},
	}, rslvAs)
}

// rslvDyn resolves a dynamic expressions. If the first element resolves to a type it is
// resolves as the 'as' expression. If it is a literal it selects an appropriate combine
// expression for that literal. The time and uuid literals have no such combine expression.
// (form +args +decls - @)
func rslvDyn(c *Ctx, env Env, e *Call, hint Type) (_ El, err error) {
	if len(e.Args) == 0 {
		return typ.Void, nil
	}
	return defaultDyn(c, env, &Dyn{Els: e.Args}, hint)
}
func defaultDyn(c *Ctx, env Env, d *Dyn, hint Type) (_ El, err error) {
	if len(d.Els) == 0 {
		return typ.Void, nil
	}
	fst := d.Els[0]
	switch t := fst.Typ(); t.Kind {
	case typ.ExpSym, typ.ExpForm, typ.ExpFunc:
		fst, err = c.Resolve(env, fst, typ.Void)
	case typ.ExpDyn:
		v, _ := fst.(*Dyn)
		if len(v.Els) == 0 {
			return typ.Void, nil
		}
		fst, err = defaultDyn(c, env, v, typ.Void)
	}
	if err != nil {
		return d, err
	}
	var def *Def
	var sym string
	args := d.Els
	switch t := fst.Typ(); t.Kind & typ.MaskElem {
	case typ.KindTyp:
		if a, ok := fst.(*Atom); ok {
			fst = a.Lit
		}
		tt := fst.(Type)
		if tt == typ.Void {
			return fst, nil
		}
		if tt == typ.Bool {
			def, args = DefSpec(formBool), args[1:]
		} else {
			sym = "as"
		}
	case typ.KindExp:
		r, ok := fst.(*Spec)
		if ok {
			def, args = DefSpec(r), args[1:]
		}
	default:
		if len(d.Els) == 1 && t.Kind&typ.MaskBase != 0 {
			if a, ok := fst.(*Atom); ok {
				fst = a.Lit
			}
			return fst, nil
		}
		switch t.Kind & typ.MaskElem {
		case typ.KindBool:
			sym = "and"
		case typ.BaseNum, typ.KindInt, typ.KindReal, typ.KindSpan:
			sym = "add"
		case typ.BaseChar, typ.KindStr, typ.KindRaw:
			sym = "cat"
		case typ.BaseList, typ.KindArr:
			sym = "apd" // TODO maybe cat
		case typ.BaseDict, typ.KindMap, typ.KindObj:
			sym = "set" // TODO maybe merge
		}
	}
	if sym != "" {
		def = LookupSupports(env, sym, '~')
	}
	if def != nil {
		return def.Resolve(c, env, &Call{Def: def, Args: args}, hint)
	}
	return nil, cor.Errorf("unexpected first argument %[1]T %[1]s in dynamic expression\n%s %s",
		fst, sym, fst.Typ())
}

// rslvAs is a type conversion or constructor and must start with a type. It has four forms:
//    Without further arguments it returns the zero literal for that type.
//    With one literal compatible to that type it returns the converted literal.
//    For keyer types one or more declarations are set.
//    For idxer types one ore more literals are appended.
// (form +t typ +args list +unis dict - @t)
func rslvAs(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	// resolve all arguments
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		t, ok := lo.Arg(0).(Type)
		if ok {
			e.Type = t
		}
		return e, err
	}
	t, ok := lo.Arg(0).(Type)
	if !ok {
		return nil, errAsType
	}
	if t == typ.Void { // just in case we have a dynamic comment
		return typ.Void, nil
	}
	args := lo.Args(1)
	decls, err := lo.Unis(2)
	if err != nil {
		return nil, err
	}
	// first rule: return zero literal
	if len(args) == 0 && len(decls) == 0 {
		return lit.Zero(t), nil
	}
	// second rule: convert compatible literals
	if len(args) == 1 && len(decls) == 0 {
		fst := args[0].(Lit)
		res, err := lit.Convert(fst, t, 0)
		if err == nil {
			return res, nil
		}
	}
	// third rule: set declarations
	if t.Kind&typ.BaseDict != 0 {
		res := deopt(lit.Zero(t)).(lit.Keyer)
		for _, d := range decls {
			el, ok := d.Arg().(Lit)
			if !ok {
				return nil, cor.Errorf("want literal in declaration got %s", d.El)
			}
			log.Printf("got delc %s %s", d.Key(), el)
			_, err = res.SetKey(d.Key(), el)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	}
	// fourth rule: append list
	if ok && t.Kind&typ.BaseList != 0 {
		res := deopt(lit.Zero(t)).(lit.Idxer)
		apd, _ := res.(lit.Appender)
		for i, a := range args {
			el, ok := a.(Lit)
			if !ok {
				return nil, cor.Error("want literal in argument list")
			}
			if apd != nil { // list or arr uses append
				apd, err = apd.Append(el)
			} else { // otherwise its an obj literal set by index
				_, err = res.SetIdx(i, el)
			}
			if err != nil {
				return nil, err
			}
		}
		if apd != nil {
			return apd, nil
		}
		return res, nil
	}
	return nil, cor.Error("not implemented")
}
