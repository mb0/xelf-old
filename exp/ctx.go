package exp

import "github.com/mb0/xelf/typ"

// DynResolver is a special resolver for dynamic expressions.
type DynResolver interface {
	ResolveDyn(c *Ctx, env Env, d *Dyn, hint Type) (El, error)
}

// Ctx is the resolution context that defines the resolution level and collects information.
type Ctx struct {
	// Part indicates that the resolution should replace partially resolved results.
	Part bool

	// Exec indicates that the resolution is expected to successfully resolve.
	//
	// A resolver will only be called once with exec set to true for each instance.
	// This means that all sub-expressions must also successfully resolve and any error
	// even the special ErrUnres will end resolution.
	// Expressions with side effects or any interaction outside the resolution environment
	// should only attempt to resolve if exec is true.
	Exec bool

	// Unres is a list of all unresolved expressions and type and symbol references.
	Unres []El

	// Dyn is a configurable resolver for dynamic expressions. A default resolver is
	// used if this field is nil.
	Dyn *Spec

	*typ.Ctx
}

func NewCtx(part, exec bool) *Ctx {
	return &Ctx{Part: part, Exec: exec, Ctx: &typ.Ctx{}}
}

// WithPart returns a copy of c with part set to val.
func (c Ctx) WithPart(val bool) *Ctx {
	c.Part = val
	return &c
}

// WithExec returns a copy of c with exec set to val.
func (c Ctx) WithExec(val bool) *Ctx {
	c.Exec = val
	return &c
}
