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
	Resolve(c *Ctx, env Env, e *Call, hint Type) (El, error)
}

// Arg returns the argument parameters or nil.
func (f *Spec) Arg() []typ.Param {
	if f.IsZero() {
		return nil
	}
	return f.Params[:len(f.Params)-1]
}

// Res returns the result type or void.
func (f *Spec) Res() Type {
	if f.IsZero() {
		return typ.Void
	}
	return f.Params[len(f.Params)-1].Type
}

func (f *Spec) Typ() typ.Type { return f.Type }
func (f *Spec) IsZero() bool {
	return f == nil || f.Resl == nil || !f.HasParams()
}

func (f *Spec) String() string { return bfr.String(f) }
func (f *Spec) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('(')
	err := f.Type.WriteBfr(b)
	if err != nil {
		return err
	}
	if f.Resl == nil {
		b.WriteString("()")
	} else {
		if v, ok := f.Resl.(bfr.Writer); ok {
			b.WriteByte(' ')
			if err = v.WriteBfr(b); err != nil {
				return err
			}
		} else {
			b.WriteString(" _")
		}
	}
	return b.WriteByte(')')
}
