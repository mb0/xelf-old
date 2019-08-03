package exp

import (
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type ReslFunc func(*Ctx, Env, *Call, typ.Type) (El, error)

func (rf ReslFunc) Resolve(c *Ctx, env Env, e *Call, hint typ.Type) (El, error) {
	return rf(c, env, e, hint)
}

func Sig(sig string) (typ.Type, error) {
	s, err := typ.Read(strings.NewReader(sig))
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
