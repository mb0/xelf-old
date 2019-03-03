package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var errAsType = cor.StrError("the 'as' expression must start with a type")

var formAs *Form

func init() {
	core.add("dyn", typ.Infer, nil, rslvDyn)
	formAs = core.add("as", typ.Bool, nil, rslvAs)
}

// rslvDyn resolves a dynamic expressions. If the first element resolves to a type it is
// resolves as the 'as' expression. If it is a literal it selects an appropriate combine
// expression for that literal. The time and uuid literals have no such combine expression.
// (form +args? list +decls? dict - @)
func rslvDyn(c *Ctx, env Env, e *Expr, hint Type) (_ El, err error) {
	if len(e.Args) == 0 {
		return typ.Void, nil
	}
	return defaultDyn(c, env, e.Args, hint)
}
func defaultDyn(c *Ctx, env Env, d Dyn, hint Type) (_ El, err error) {
	if len(d) == 0 {
		return typ.Void, nil
	}
	fst := d[0]
	switch t := fst.Typ(); t.Kind {
	case typ.ExpSym, typ.ExpForm, typ.ExpFunc:
		fst, err = c.Resolve(env, fst, typ.Void)
	case typ.ExpDyn:
		v := fst.(Dyn)
		if len(v) == 0 {
			return typ.Void, nil
		}
		fst, err = defaultDyn(c, env, v, typ.Void)
	}
	if err != nil {
		return d, err
	}
	var xr ExprResolver
	var sym string
	args := d
	switch t := fst.Typ(); t.Kind & typ.MaskElem {
	case typ.KindTyp:
		sym = "as"
		if fst == typ.Bool {
			sym, args = "bool", args[1:]
		}
	case typ.KindExp:
		r, ok := fst.(ExprResolver)
		if ok {
			xr, args = r, args[1:]
		}
	default:
		if len(d) == 1 {
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
		r, ok := Lookup(env, sym).(ExprResolver)
		if ok {
			xr = r
		}
	}
	if xr != nil {
		return xr.Resolve(c, env, &Expr{xr, args}, hint)
	}
	return nil, cor.Errorf("unexpected first argument %T in dynamic expression", d[0])
}

// rslvAs is a type conversion or constructor and must start with a type. It has four forms:
//    Without further arguments it returns the zero literal for that type.
//    With one literal compatible to that type it returns the converted literal.
//    For keyer types one or more declarations are set.
//    For idxer types one ore more literals are appended.
// (form +t typ +rest? list - @t)
func rslvAs(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	if len(e.Args) == 0 {
		return nil, errAsType
	}
	targ, err := c.Resolve(env, e.Args[0], typ.Typ)
	if err != nil {
		return e, err
	}
	t, ok := targ.(Type)
	if !ok {
		return nil, errAsType
	}
	if t == typ.Void { // just in case we have a dynamic comment
		return nil, nil
	}
	args := e.Args[1:]
	// first rule: return zero literal
	if len(args) == 0 {
		return lit.Zero(t), nil
	}
	// resolve all arguments
	args, err = c.ResolveAll(env, args, typ.Any)
	if err != nil {
		return e, err
	}
	fst, ok := args[0].(Lit)
	// second rule: convert compatible literals
	if ok && len(args) == 1 {
		res, err := lit.Convert(fst, t, 0)
		if err == nil {
			return res, nil
		}
	}
	// third rule: set declarations
	if !ok && t.Kind&typ.BaseDict != 0 {
		decls, err := UniDeclForm(args)
		if err != nil {
			return nil, err
		}
		res := deopt(lit.Zero(t)).(lit.Keyer)
		for _, d := range decls {
			el, ok := d.Args[0].(Lit)
			if !ok {
				return nil, cor.Error("want literal in declaration argument")
			}
			err = res.SetKey(d.Name[1:], el)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	}
	// fourth rule: append list
	if ok && t.Kind&typ.BaseList != 0 {
		err = ArgsForm(args)
		if err != nil {
			return nil, err
		}
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
				err = res.SetIdx(i, el)
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
