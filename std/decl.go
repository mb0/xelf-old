package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var withSpec = core.impl("(form 'with' any :rest list|expr @)",
	func(x exp.ReslReq) (exp.El, error) {
		dot := x.Arg(0)
		el, err := x.Ctx.Resolve(x.Env, dot, typ.Void)
		if err != nil {
			return x.Call, err
		}
		env := &exp.DataScope{x.Env, el.(lit.Lit)}
		rest := x.Args(1)
		if len(rest) == 0 {
			return nil, cor.Errorf("with must have body expressions")
		}
		rest, err = x.ResolveAll(env, rest, typ.Void)
		if err != nil {
			return x.Call, err
		}
		return rest[len(rest)-1], nil
	})

// letSpec declares one or more resolvers in a new scope and resolves the tailing actions.
// It returns the last actions result.
var letSpec = decl.impl("(form 'let' :tags dict|any :plain list|expr @)",
	func(x exp.ReslReq) (exp.El, error) {
		decls, err := x.Unis(0)
		if err != nil {
			return nil, err
		}
		rest := x.Args(1)
		if len(rest) == 0 || len(decls) == 0 {
			return nil, cor.Errorf("let must have declarations and a body")
		}
		s := exp.NewScope(x.Env)
		if len(decls) > 0 {
			res, err := letDecls(x.Ctx, s, decls)
			if err != nil {
				return x.Call, err
			}
			if len(rest) == 0 {
				return res, nil
			}
		}
		rest, err = x.ResolveAll(s, rest, typ.Void)
		if err != nil {
			return x.Call, err
		}
		return rest[len(rest)-1], nil
	})

// fnSpec declares a function literal from its arguments.
var fnSpec = decl.impl("(form 'fn' :tags? dict|typ :plain list|expr @)",
	func(x exp.ReslReq) (exp.El, error) {
		decls, err := x.Unis(0)
		if err != nil {
			return nil, err
		}
		rest := x.Args(1)
		var sig typ.Type
		if len(decls) == 0 {
			// TODO infer signature
			return nil, cor.Errorf("inferred fn expressions not implemented")
		} else {
			// construct sig from decls
			fs := make([]typ.Param, 0, len(decls))
			for _, d := range decls {
				l, err := x.Ctx.Resolve(x.Env, d.El, typ.Void)
				if err != nil {
					return x.Call, err
				}
				dt, ok := l.(typ.Type)
				if !ok {
					return nil, cor.Errorf("want type in func parameters got %T", l)
				}
				fs = append(fs, typ.Param{Name: d.Name[1:], Type: dt})
			}
			sig = typ.Type{Kind: typ.KindFunc, Info: &typ.Info{Params: fs}}
		}
		return &exp.Spec{sig, &exp.ExprBody{rest, x.Env}}, nil
	})

func letDecls(c *exp.Ctx, env *exp.Scope, decls []*exp.Named) (res exp.El, err error) {
	for _, d := range decls {
		if len(d.Name) < 2 {
			return nil, cor.Error("unnamed declaration")
		}
		if d.El == nil {
			return nil, cor.Error("naked declaration")
		}
		res, err = c.Resolve(env, d.El, typ.Void)
		if err != nil {
			return nil, err
		}
		switch l := res.(type) {
		case lit.Lit:
			if r, ok := l.(*exp.Spec); ok {
				err = env.Def(d.Key(), exp.NewDef(r))
			} else {
				err = env.Def(d.Key(), exp.NewDef(l))
			}
		default:
			return nil, cor.Errorf("unexpected element as declaration value %v", res)
		}
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
