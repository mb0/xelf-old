package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var errAsType = cor.StrError("the 'as' expression must start with a type")

// rslvDyn resolves a dynamic expressions. If the first element resolves to a type it is
// resolves as the 'as' expression. If it is a literal it selects an appropriate combine
// expression for that literal. The time and uuid literals have no such combine expression.
func rslvDyn(c *Ctx, env Env, e *Expr) (_ El, err error) {
	if len(e.Args) == 0 {
		return typ.Void, nil
	}
	var ref Ref
	fst := e.Args[0]
	switch v := fst.(type) {
	case *Ref:
		fst, err = c.Resolve(env, v)
		ref = *v
	case *Expr:
		fst, err = c.Resolve(env, v)
	case Dyn:
		fst, err = rslvDyn(c, env, &Expr{Args: v})
	}
	if err != nil {
		return e, err
	}
	var sym string
	args := e.Args
	switch v := fst.(type) {
	case Callable:
		return v.Resolve(c, env, &Expr{ref, args[1:], v})
	case Type:
		sym = "as"
		if v == typ.Bool {
			sym, args = "bool", args[1:]
		}
	case Lit:
		if len(e.Args) == 1 {
			return v, nil
		}
		switch v.Typ().Kind & typ.MaskElem {
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
		if rslv := Lookup(env, sym); rslv != nil {
			return rslv.Resolve(c, env, &Expr{Ref{Name: sym}, args, rslv})
		}
	}
	return nil, cor.Errorf("unexpected first argument %T in dynamic expression", e.Args[0])
}

// rslvAs is a type conversion or constructor and must start with a type. It has four forms:
//    Without further arguments it returns the zero literal for that type.
//    With one literal compatible to that type it returns the converted literal.
//    For keyer types one or more declarations are set.
//    For idxer types one ore more literals are appended.
func rslvAs(c *Ctx, env Env, e *Expr) (El, error) {
	if len(e.Args) == 0 {
		return nil, errAsType
	}
	targ, err := c.Resolve(env, e.Args[0])
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
	args, err = c.ResolveAll(env, args)
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
