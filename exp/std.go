package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type ReslFunc func(*Ctx, Env, *Call, Type) (El, error)

func (rf ReslFunc) Resolve(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	return rf(c, env, e, hint)
}

func Sig(sig string) (Type, error) {
	s, err := typ.ParseString(sig)
	if err != nil {
		return typ.Void, cor.Errorf("cannot parse signature %s: %v", sig, err)
	}
	switch s.Kind {
	case typ.KindForm, typ.KindFunc:
	default:
		return typ.Void, cor.Errorf("not a form or func signature %s", sig)
	}
	return s, nil
}

func MustSig(sig string) Type {
	s, err := Sig(sig)
	if err != nil {
		panic(cor.Errorf("implement spec error: %v", err))
	}
	return s
}

type LayoutResolverFunc = func(*Ctx, Env, *Call, *Layout, Type) (El, error)

type ReslReq struct {
	*Ctx
	Env  Env
	Call *Call
	*Layout
	Hint Type
}
type ReslReqFunc = func(ReslReq) (El, error)

func Implement(sig string, resolve bool, r LayoutResolverFunc) *Spec {
	s := MustSig(sig)
	return &Spec{s, ReslFunc(func(c *Ctx, env Env, e *Call, hint Type) (El, error) {
		if e.Type == typ.Void {
			e.Type = c.Inst(s)
		}
		lo, err := LayoutArgs(e.Type, e.Args)
		if err != nil {
			return nil, err
		}
		if resolve {
			err = lo.Resolve(c, env, hint)
			if err != nil {
				e.Type = lo.Sig
				return e, err
			}
		}
		return r(c, env, e, lo, hint)
	})}
}

func ImplementReq(sig string, resolve bool, r ReslReqFunc) *Spec {
	s := MustSig(sig)
	return &Spec{s, ReslFunc(func(c *Ctx, env Env, e *Call, hint Type) (El, error) {
		if e.Type == typ.Void {
			e.Type = c.Inst(s)
		}
		lo, err := LayoutArgs(e.Type, e.Args)
		if err != nil {
			return nil, err
		}
		if resolve {
			err = lo.Resolve(c, env, hint)
			if err != nil {
				e.Type = lo.Sig
				return e, err
			}
		}
		return r(ReslReq{c, env, e, lo, hint})
	})}
}
