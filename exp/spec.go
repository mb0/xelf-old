package exp

import (
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

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

type Impl interface {

	// Resl resolves a call and returns the resulting element or an error.
	//
	// A successful resolution returns a call with all related types resolved and no error.
	// If the type hint is not void, it is used to check or infer the element type.
	// When parts of the element could not be resolved it returns the special error ErrUnres.
	Resl(p *Prog, env Env, c *Call, h typ.Type) (El, error)

	// Eval evaluates a call and returns the resulting element or an error.
	//
	// A successful evaluation returns a literal and no error.
	// If the type hint is not void, it is used to check or infer the element type.
	// When parts of the element could not be evaluation it returns the special error ErrUnres,
	// and - if the context allows it - a partially resolved element.
	Eval(p *Prog, env Env, c *Call, h typ.Type) (El, error)
}

type Spec struct {
	typ.Type
	Impl
}

// Arg returns the argument parameters or nil.
func (sp *Spec) Arg() []typ.Param {
	if sp.IsZero() {
		return nil
	}
	return sp.Params[:len(sp.Params)-1]
}

// Res returns the result type or void.
func (sp *Spec) Res() typ.Type {
	if sp.IsZero() {
		return typ.Void
	}
	return sp.Params[len(sp.Params)-1].Type
}

func (sp *Spec) Typ() typ.Type { return sp.Type }
func (sp *Spec) IsZero() bool {
	return sp == nil || sp.Impl == nil || !sp.HasParams()
}

func (sp *Spec) String() string { return bfr.String(sp) }
func (sp *Spec) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('(')
	switch r := sp.Impl.(type) {
	case *ExprBody:
		b.WriteString("fn")
		if err := r.WriteBfr(b); err != nil {
			return err
		}
	case bfr.Writer:
		err := sp.Type.WriteBfr(b)
		if err != nil {
			return err
		}
		if err = r.WriteBfr(b); err != nil {
			return err
		}
	default:
		err := sp.Type.WriteBfr(b)
		if err != nil {
			return err
		}
		b.WriteString("_")
	}
	return b.WriteByte(')')
}
