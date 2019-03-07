package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

func init() {
	std.add("let", []typ.Param{
		{Name: "unis"},
		{Type: typ.Infer},
	}, rslvLet)
	unisRest := []typ.Param{
		{Name: "unis"},
		{Name: "plain"},
		{Type: typ.Infer},
	}
	std.add("with", unisRest, rslvWith)
	std.add("fn", unisRest, rslvFn)
}

// rslvLet declares one or more resolvers in the existing scope.
// (form 'let' +unis - @)
func rslvLet(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := LayoutArgs(e.Rslv.Arg(), e.Args)
	if err != nil {
		return nil, err
	}
	decls, err := lo.Unis(0)
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
// (form 'with' +unis +rest - @)
func rslvWith(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := LayoutArgs(e.Rslv.Arg(), e.Args)
	if err != nil {
		return nil, err
	}
	decls, err := lo.Unis(0)
	if err != nil {
		return nil, err
	}
	rest := lo.Args(1)
	if len(rest) == 0 {
		return nil, cor.Errorf("with must have an expression")
	}
	s := NewScope(env)
	_, err = letDecls(c, s, decls)
	if err != nil {
		return e, err
	}
	rest, err = c.ResolveAll(s, rest, typ.Void)
	if err != nil {
		return e, err
	}
	return rest[len(rest)-1], nil
}

// rslvFn declares a function literal from its arguments.
// (form 'fn' +unis +rest - @)
func rslvFn(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := LayoutArgs(e.Rslv.Arg(), e.Args)
	if err != nil {
		return nil, err
	}
	decls, err := lo.Unis(0)
	if err != nil {
		return nil, err
	}
	rest := lo.Args(1)
	var sig Sig
	if len(decls) == 0 {
		// TODO infer signature
		return nil, cor.Errorf("inferred fn expressions not implemented")
	} else {
		// construct sig from decls
		fs := make([]typ.Param, 0, len(decls))
		for _, d := range decls {
			l, err := c.Resolve(env, d.Args[0], typ.Void)
			if err != nil {
				return e, err
			}
			dt, ok := l.(Type)
			if !ok {
				return nil, cor.Errorf("want type in func parameters got %T", l)
			}
			fs = append(fs, typ.Param{Name: d.Name[1:], Type: dt})
		}
		sig = Sig{Kind: typ.ExpFunc, Info: &typ.Info{Params: fs}}
	}
	return &Func{sig, &ExprBody{rest, env}}, nil
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
