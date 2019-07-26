package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// Realize finalizes all types in el or returns an error.
// If successful, el is independent of its type context.
func Realize(c *Ctx, el El) error {
	v := &realizer{Ctx: c}
	err := el.Traverse(v)
	if err != nil {
		return cor.Errorf("traversal of %s: %v", el, err)
	}
	if len(v.free) > 0 {
		return cor.Errorf("free variables: %s", v.free)
	}
	return nil
}

type realizer struct {
	Ghost
	*Ctx
	free typ.Vars
}

func (v *realizer) visit(o typ.Type) typ.Type {
	t, err := v.Realize(o)
	if err != nil {
		v.free = v.Free(t, v.free)
	}
	return t
}
func (v *realizer) VisitLit(a *Atom) error {
	if s, ok := a.Lit.(*Spec); ok {
		if _, ok := s.Resl.(*ExprBody); ok {
			s.Type = v.visit(s.Type)
		}
	}
	return nil
}
func (v *realizer) VisitType(a *Atom) error {
	a.Lit = v.visit(a.Lit.(typ.Type))
	return nil
}
func (v *realizer) VisitSym(a *Sym) error {
	a.Type = v.visit(a.Type)
	return nil
}
func (v *realizer) EnterCall(a *Call) error {
	a.Type = v.visit(a.Type)
	return nil
}
