package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

type Spec struct {
	typ.Type
	Resl
}

type Resl interface {
	// Resolve resolves a call and returns the result or an error.
	//
	// A successful resolution returns a literal and no error.
	// If the type hint is not void, it is used to check or infer the element type.
	// When parts of the element could not be resolved it returns the special error ErrUnres,
	// and either the original element or if the context allows it a partially resolved element.
	// If the resolution cannot proceed with execution it returns the special error ErrExec.
	// Any other error ends the whole resolution process.
	Resolve(c *Ctx, env Env, e *Call, hint typ.Type) (El, error)
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
	return sp == nil || sp.Resl == nil || !sp.HasParams()
}

func (sp *Spec) String() string { return bfr.String(sp) }
func (sp *Spec) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('(')
	switch r := sp.Resl.(type) {
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
