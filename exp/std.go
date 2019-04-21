package exp

import "github.com/mb0/xelf/typ"

var StdEnv = Builtin{Core, Std}

var core = make(formMap, 32)
var std = make(formMap, 8)

func Core(sym string) *Spec {
	if f, ok := core[sym]; ok {
		return f
	}
	return nil
}

func Std(sym string) *Spec {
	if f, ok := std[sym]; ok {
		return f
	}
	return nil
}

type FormResolverFunc func(*Ctx, Env, *Call, Type) (El, error)

func (rf FormResolverFunc) ResolveCall(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	return rf(c, env, e, hint)
}

type formMap map[string]*Spec

func (m formMap) add(ref string, params []typ.Param, r FormResolverFunc) *Spec {
	f := &Spec{typ.Type{Kind: typ.ExpForm, Info: &typ.Info{
		Ref:    ref,
		Params: params,
	}}, r}
	m[ref] = f
	return f
}
