package utl

import (
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

type Program struct {
	Builtin  exp.Builtin
	Arg, Res *lit.Dict
}

func Prog(arg *lit.Dict, b ...func(string) exp.Resolver) exp.Env {
	return exp.NewScope(&Program{
		Builtin: b,
		Arg:     arg,
		Res:     &lit.Dict{},
	})
}

func (*Program) Parent() exp.Env { return nil }
func (*Program) Supports(x byte) bool {
	return x == '$' || x == '/'
}
func (p *Program) Def(sym string, r exp.Resolver) error { return exp.ErrNoDefEnv }
func (p *Program) Get(sym string) exp.Resolver {
	var pr PathResolver
	switch sym[0] {
	case '$':
		pr = p.resolveArg
	case '/':
		pr = p.resolveRes
	default:
		return p.Builtin.Get(sym)
	}
	return pr
}
func (p *Program) resolveArg(c *exp.Ctx, env exp.Env, s string) (exp.El, error) {
	return lit.Select(p.Arg, s[1:])
}
func (p *Program) resolveRes(c *exp.Ctx, env exp.Env, s string) (exp.El, error) {
	return lit.Select(p.Res, s[1:])
}

type PathResolver func(*exp.Ctx, exp.Env, string) (exp.El, error)

func (rf PathResolver) Resolve(c *exp.Ctx, env exp.Env, e exp.El) (exp.El, error) {
	switch x := e.(type) {
	case *exp.Ref:
		return rf(c, env, x.Name)
	case *exp.Expr:
		l, err := rf(c, env, x.Name)
		if err != nil {
			return e, err
		}
		return c.Resolve(env, append(exp.Dyn{l}, x.Args...))
	}
	return e, exp.ErrUnres
}
