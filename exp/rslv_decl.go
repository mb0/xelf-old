package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// letSpec declares one or more resolvers in a new scope and resolves the tailing actions.
// It returns the last actions result.
var letSpec = std.impl("(form 'let' :a? any :unis :rest : @)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		dot := lo.Arg(0)
		if dot != nil {
			el, err := c.Resolve(env, dot, typ.Void)
			if err != nil {
				return e, err
			}
			env = &DataScope{env, el.(Lit)}
		}
		decls, err := lo.Unis(1)
		if err != nil {
			return nil, err
		}
		rest := lo.Args(2)
		if len(rest) == 0 && len(decls) == 0 {
			return nil, cor.Errorf("let must have declarations or a body")
		}
		s := NewScope(env)
		if len(decls) > 0 {
			res, err := letDecls(c, s, decls)
			if err != nil {
				return e, err
			}
			if len(rest) == 0 {
				return res, nil
			}
		}
		rest, err = c.ResolveAll(s, rest, typ.Void)
		if err != nil {
			return e, err
		}
		return rest[len(rest)-1], nil
	})

// fnSpec declares a function literal from its arguments.
var fnSpec = std.impl("(form 'fn' :unis :rest : @)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		decls, err := lo.Unis(0)
		if err != nil {
			return nil, err
		}
		rest := lo.Args(1)
		var sig typ.Type
		if len(decls) == 0 {
			// TODO infer signature
			return nil, cor.Errorf("inferred fn expressions not implemented")
		} else {
			// construct sig from decls
			fs := make([]typ.Param, 0, len(decls))
			for _, d := range decls {
				l, err := c.Resolve(env, d.El, typ.Void)
				if err != nil {
					return e, err
				}
				dt, ok := l.(Type)
				if !ok {
					return nil, cor.Errorf("want type in func parameters got %T", l)
				}
				fs = append(fs, typ.Param{Name: d.Name[1:], Type: dt})
			}
			sig = typ.Type{Kind: typ.ExpFunc, Info: &typ.Info{Params: fs}}
		}
		return &Spec{sig, &ExprBody{rest, env}}, nil
	})

func letDecls(c *Ctx, env *Scope, decls []*Named) (res El, err error) {
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
		case Lit:
			if r, ok := l.(*Spec); ok {
				err = env.Def(d.Key(), DefSpec(r))
			} else {
				err = env.Def(d.Key(), DefLit(l))
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
