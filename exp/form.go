package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type Sig Type

func AnonSig(args []typ.Param) Sig { return FuncSig("", args) }
func FuncSig(name string, args []typ.Param) Sig {
	return Sig{typ.ExpFunc, &typ.Info{Ref: name, Params: args}}
}
func FormSig(name string, args []typ.Param) Sig {
	return Sig{typ.ExpForm, &typ.Info{Ref: name, Params: args}}
}

func (s Sig) Typ() typ.Type { return typ.Type(s) }
func (s Sig) IsZero() bool  { return s.Info == nil || len(s.Params) == 0 }

func (s Sig) Key() string {
	if s.Info != nil {
		return s.Ref
	}
	return ""
}

// Arg returns the argument parameters or nil.
func (s Sig) Arg() []typ.Param {
	if s.IsZero() {
		return nil
	}
	return s.Params[:len(s.Params)-1]
}

// Res returns the result type or void.
func (s Sig) Res() Type {
	if s.IsZero() {
		return typ.Void
	}
	return s.Params[len(s.Params)-1].Type
}

func (s Sig) String() string { return bfr.String(s) }
func (s Sig) WriteBfr(b *bfr.Ctx) error {
	key := s.Key()
	if key != "" {
		return b.Fmt(key)
	}
	return typ.Type(s).WriteBfr(b)
}

// Form represents a from resolver and implements both literal and resolver interface.
type Form struct {
	Sig
	Body FormResolver
}

type FormResolver interface {
	ResolveForm(c *Ctx, env Env, e *Expr, hint Type) (El, error)
}

func (f *Form) MarshalJSON() ([]byte, error) {
	v, err := cor.Quote(f.String(), '"')
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (f *Form) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	switch t := e.Typ(); t.Kind {
	case typ.ExpSym:
		return f, nil
	case typ.ExpForm:
		return f.Body.ResolveForm(c, env, e.(*Expr), hint)
	}
	return nil, cor.Errorf("unexpected element %s for %s", e.Typ(), f.Sig.Ref)
}
