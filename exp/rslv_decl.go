package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

func init() {
	std.add("let", typ.Infer, nil, rslvLet)
	std.add("with", typ.Infer, nil, rslvWith)
	std.add("fn", typ.Infer, nil, rslvFn)
}

// rslvLet declares one or more resolvers in the existing scope.
// (form +decls dict - @)
func rslvLet(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	decls, err := UniDeclForm(e.Args)
	if err != nil {
		return nil, err
	}
	res, err := letDecls(c, env, decls)
	if err != nil {
		return e, err
	}
	return res, nil
}

// rslvWith declares one or more resolvers in a new scope and resolves the tailing actions.
// It returns the last actions result.
// (form +decls dict +tail list - @)
func rslvWith(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	decls, tail, err := UniDeclRest(e.Args)
	if err != nil {
		return nil, err
	}
	err = ArgsMin(tail, 1)
	if err != nil {
		return nil, cor.Errorf("%v %v %v", err, decls, tail)
	}
	s := NewScope(env)
	_, err = letDecls(c, s, decls)
	if err != nil {
		return e, err
	}
	tail, err = c.ResolveAll(s, tail, typ.Void)
	if err != nil {
		return e, err
	}
	return tail[len(tail)-1], nil
}

// rslvFn declares a function literal from its arguments.
// (form +decls? dict +tail list - @)
func rslvFn(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	decls, tail, err := UniDeclRest(e.Args)
	if err != nil {
		return nil, err
	}
	var sig Type
	if len(decls) == 0 {
		// TODO infer signature
		return nil, cor.Errorf("inferred fn expressions not implemented")
	} else {
		// construct sig from decls
		fs := make([]typ.Field, 0, len(decls))
		for _, d := range decls {
			l, err := c.Resolve(env, d.Args[0], typ.Void)
			if err != nil {
				return e, err
			}
			dt, ok := l.(Type)
			if !ok {
				return nil, cor.Errorf("want type in func parameters got %T", l)
			}
			fs = append(fs, typ.Field{Name: d.Name[1:], Type: dt})
		}
		sig = Type{Kind: typ.KindFunc, Info: &typ.Info{Fields: fs}}
	}
	return &Func{sig, &ExprBody{tail, env}}, nil
}

func letDecls(c *Ctx, env Env, decls []Decl) (El, error) {
	var res El
	for _, d := range decls {
		if len(d.Name) < 2 {
			return nil, cor.Error("unnamed declaration")
		}
		args, err := c.ResolveAll(env, d.Args, typ.Void)
		if err != nil {
			return nil, err
		}
		var rslv Resolver
		switch dv := args[0].(type) {
		case Lit:
			res = dv
			if r, ok := dv.(Resolver); ok {
				rslv = r
			} else {
				rslv = LitResolver{dv}
			}
		default:
			return nil, cor.Errorf("unexpected element as declaration value %v", d.Args[0])
		}
		err = env.Def(d.Name[1:], rslv)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
