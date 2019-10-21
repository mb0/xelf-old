package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

func DefaultDyn(t typ.Type) (string, bool) {
	switch t.Kind & typ.MaskElem {
	case typ.KindTyp:
		return "con", false
	case typ.KindBool:
		return "and", false
	case typ.KindNum, typ.KindInt, typ.KindReal, typ.KindSpan:
		return "add", false
	case typ.KindChar, typ.KindStr, typ.KindRaw:
		return "cat", false
	case typ.KindIdxr, typ.KindList:
		return "apd", false
	case typ.KindKeyr, typ.KindDict, typ.KindRec:
		return "set", false
	}
	return "", false
}

// Prog is the resolution type context and also collects unresolved elements.
type Prog struct {
	// Ctx is the type context that stores type variable bindings.
	*typ.Ctx
	// Unres is a list of all unresolved expressions and type and symbol references.
	Unres []El

	Dyn func(fst typ.Type) (sym string, consume bool)
}

func NewProg() *Prog {
	return &Prog{Ctx: &typ.Ctx{}, Dyn: DefaultDyn}
}

// NewCall returns a new call or an error if arguments do not match the spec signature.
// The call signature is instantiated in the programs type context.
func (p *Prog) NewCall(s *Spec, args []El, src lex.Src) (*Call, error) {
	inst := p.Inst(s.Type)
	lo, err := SigLayout(inst, args)
	if err != nil {
		return nil, err
	}
	return &Call{Layout: *lo, Spec: s, Src: src}, nil
}

// BuiltinCall looks up the builtin spec by name and returns a new call or returns an error.
func (p *Prog) BuiltinCall(env Env, name string, args []El, src lex.Src) (*Call, error) {
	def := LookupSupports(env, name, '~')
	if def == nil {
		return nil, cor.Errorf("new call name %q not defined", name)
	}
	s, ok := def.Lit.(*Spec)
	if !ok {
		return nil, cor.Errorf("new call name %q is a %T", name, def.Lit)
	}
	return p.NewCall(s, args, src)
}

func (p *Prog) checkHint(hint typ.Type, el El) (El, error) {
	if hint == typ.Void {
		return el, nil
	}
	r := ResType(el)
	if r == typ.Void {
		return nil, cor.Errorf("check hint: unexpected element %s", el)
	}
	_, err := typ.Unify(p.Ctx, r, hint)
	if err != nil {
		return nil, cor.Errorf("check hint: %v", err)
	}
	return el, nil
}

// Realize finalizes all types in el or returns an error.
// If successful, el is independent of its type context.
func (p *Prog) Realize(el El) error {
	v := &realizer{Prog: p}
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
	*Prog
	free typ.Vars
}

func (v *realizer) visit(o typ.Type) typ.Type {
	t, err := v.Ctx.Realize(o)
	if err != nil {
		v.free = v.Free(t, v.free)
	}
	return t
}
func (v *realizer) VisitLit(a *Atom) error {
	if s, ok := a.Lit.(*Spec); ok {
		if _, ok := s.Impl.(*ExprBody); ok {
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
	a.Sig = v.visit(a.Sig)
	return nil
}
