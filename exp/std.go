package exp

import "github.com/mb0/xelf/typ"

var StdEnv = Builtin{Core, Std}

var core = make(formMap, 32)
var std = make(formMap, 8)

func Core(sym string) Resolver {
	if f, ok := core[sym]; ok {
		return f
	}
	return nil
}

func Std(sym string) Resolver {
	if f, ok := std[sym]; ok {
		return f
	}
	return nil
}

type ExprResolverFunc func(*Ctx, Env, *Expr, Type) (El, error)

func (rf ExprResolverFunc) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	xp, ok := e.(*Expr)
	if !ok {
		return &Form{Type{Kind: typ.KindForm}, rf}, nil
	}
	return rf(c, env, xp, hint)
}

type formMap map[string]*Form

func (m formMap) add(ref string, res Type, params []typ.Field, r ExprResolverFunc) *Form {
	f := &Form{Type{Kind: typ.KindForm, Info: &typ.Info{
		Ref:    ref,
		Fields: append(params, typ.Field{Type: res}),
	}}, r}
	m[ref] = f
	return f
}
