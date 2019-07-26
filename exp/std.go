package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type ReslFunc func(*Ctx, Env, *Call, typ.Type) (El, error)

func (rf ReslFunc) Resolve(c *Ctx, env Env, e *Call, hint typ.Type) (El, error) {
	return rf(c, env, e, hint)
}

func Sig(sig string) (typ.Type, error) {
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

func MustSig(sig string) typ.Type {
	s, err := Sig(sig)
	if err != nil {
		panic(cor.Errorf("implement spec error: %v", err))
	}
	return s
}

type LayoutResolverFunc = func(*Ctx, Env, *Call, *Layout, typ.Type) (El, error)

type ReslReq struct {
	*Ctx
	Env  Env
	Call *Call
	*Layout
	Hint typ.Type
}
type ReslReqFunc = func(ReslReq) (El, error)

func ImplementReq(sig string, resolve bool, r ReslReqFunc) *Spec {
	s := MustSig(sig)
	return &Spec{s, ReslFunc(func(c *Ctx, env Env, e *Call, hint typ.Type) (El, error) {
		if e.Type == typ.Void {
			return nil, cor.Errorf("type not instantiated for %s %s", s, e.Type)
		}
		lo, err := LayoutArgs(e.Type, e.Args)
		if err != nil {
			return nil, err
		}
		if resolve {
			err = lo.Resolve(c, env, hint)
			e.Type = lo.Sig
			if err != nil {
				return e, err
			}
			if !c.Exec && !c.Part {
				return e, nil
			}
		}
		return r(ReslReq{c, env, e, lo, hint})
	})}
}
