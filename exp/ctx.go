package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

// Ctx is the resolution context that defines the resolution level and collects information.
type Ctx struct {
	// Part indicates that the resolution should replace partially resolved results.
	Part bool

	// Unres is a list of all unresolved expressions and type and symbol references.
	Unres []El

	*typ.Ctx
}

func NewCtx() *Ctx {
	return &Ctx{Ctx: &typ.Ctx{}}
}

// WithPart returns a copy of c with part set to val.
func (c Ctx) WithPart(val bool) *Ctx {
	c.Part = val
	return &c
}

func (c *Ctx) NewCall(s *Spec, args []El, src lex.Src) (*Call, error) {
	inst := c.Inst(s.Type)
	lo, err := SigLayout(inst, args)
	if err != nil {
		return nil, err
	}
	return &Call{Layout: *lo, Spec: s, Src: src}, nil
}

func (c *Ctx) BuiltinCall(env Env, name string, args []El, src lex.Src) (*Call, error) {
	def := LookupSupports(env, name, '~')
	if def == nil {
		return nil, cor.Errorf("new call name %q not defined", name, def.Lit)
	}
	s, ok := def.Lit.(*Spec)
	if !ok {
		return nil, cor.Errorf("new call name %q is a %T", name, def.Lit)
	}
	return c.NewCall(s, args, src)
}

func (c *Ctx) checkHint(hint typ.Type, el El) (El, error) {
	if hint == typ.Void {
		return el, nil
	}
	r := ResType(el)
	if r == typ.Void {
		return nil, cor.Errorf("check hint: unexpected element %s", el)
	}
	_, err := typ.Unify(c.Ctx, r, hint)
	if err != nil {
		return nil, cor.Errorf("check hint: %v", err)
	}
	return el, nil
}
