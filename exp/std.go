package exp

import (
	"log"

	"github.com/mb0/xelf/typ"
)

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

type LayoutResolverFunc func(*Ctx, Env, *Call, *Layout, Type) (El, error)

func (m formMap) impl(sig string, r LayoutResolverFunc) *Spec {
	s, err := typ.ParseString(sig)
	if err != nil {
		log.Fatalf("cannot parse spec signature %s: %v", sig, err)
	}
	f := &Spec{s, FormResolverFunc(func(c *Ctx, env Env, e *Call, hint Type) (El, error) {
		lo, err := LayoutArgs(e.Spec.Arg(), e.Args)
		if err != nil {
			return nil, err
		}
		return r(c, env, e, lo, hint)
	})}
	m[s.Ref] = f
	return f
}

func (m formMap) implResl(sig string, r LayoutResolverFunc) *Spec {
	s, err := typ.ParseString(sig)
	if err != nil {
		log.Fatalf("cannot parse spec signature %s: %v", sig, err)
	}
	f := &Spec{s, FormResolverFunc(func(c *Ctx, env Env, e *Call, hint Type) (El, error) {
		lo, err := LayoutArgs(e.Spec.Arg(), e.Args)
		if err != nil {
			return nil, err
		}
		err = lo.Resolve(c, env)
		if err != nil {
			return e, err
		}
		return r(c, env, e, lo, hint)
	})}
	m[s.Ref] = f
	return f
}
