package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

type Sig Type

func AnonSig(args []typ.Field) Sig { return FuncSig("", args) }
func FuncSig(name string, args []typ.Field) Sig {
	return Sig{typ.ExpFunc, &typ.Info{Ref: name, Fields: args}}
}
func FormSig(name string, args []typ.Field) Sig {
	return Sig{typ.ExpFunc, &typ.Info{Ref: name, Fields: args}}
}

func (s Sig) Typ() typ.Type { return typ.Type(s) }
func (s Sig) IsZero() bool  { return s.Info == nil || len(s.Fields) == 0 }

func (s Sig) Key() string {
	if s.Info != nil {
		return s.Ref
	}
	return ""
}

// Params returns the parameter fields or nil.
func (s Sig) Params() []typ.Field {
	if s.IsZero() {
		return nil
	}
	return s.Fields[:len(s.Fields)-1]
}

// Res returns the result type or void.
func (s Sig) Res() Type {
	if s.IsZero() {
		return typ.Void
	}
	return s.Fields[len(s.Fields)-1].Type
}

func (s Sig) String() string { return bfr.String(s) }
func (s Sig) WriteBfr(b bfr.Ctx) error {
	key := s.Key()
	if key != "" {
		return b.Fmt(key)
	}
	return typ.Type(s).WriteBfr(b)
}

// Form represents a from resolver and implements both literal and resolver interface.
type Form struct {
	Sig
	Rslv Resolver
}

func (f *Form) MarshalJSON() ([]byte, error) {
	v, err := lex.Quote(f.String(), '"')
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (f *Form) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	switch e.(type) {
	case *Ref:
		return f, nil
	case *Expr:
		return f.Rslv.Resolve(c, env, e, hint)
	}
	return nil, cor.Errorf("unexpected element %T", e)
}
