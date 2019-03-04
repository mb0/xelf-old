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

type FormResolverFunc func(*Ctx, Env, *Expr, Type) (El, error)

func (rf FormResolverFunc) ResolveForm(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	return rf(c, env, e, hint)
}

type formMap map[string]*Form

func (m formMap) add(ref string, params []typ.Param, r FormResolverFunc) *Form {
	f := &Form{Sig{Kind: typ.ExpForm, Info: &typ.Info{
		Ref:    ref,
		Params: params,
	}}, r}
	m[ref] = f
	return f
}
