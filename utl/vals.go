package utl

import (
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// TypedUnresolver will only set enclosed type to the ref or expr to be resolved.
// This is useful for partial resolution phases where the type is known but the literal isn't.
type TypedUnresolver struct{ typ.Type }

func (r TypedUnresolver) Resolve(c *exp.Ctx, env exp.Env, e exp.El) (exp.El, error) {
	switch v := e.(type) {
	case *exp.Ref:
		v.Type = r.Type
	case *exp.Expr:
		v.Type = r.Type
	}
	return e, exp.ErrUnres
}

// LitResilver will resolve to the enclosed literal. When not used as ref it resolves as dynamic
// starting with the enclosed literal followed by the expr arguments.
type LitResolver struct{ lit.Lit }

func (r LitResolver) Resolve(c *exp.Ctx, env exp.Env, e exp.El) (exp.El, error) {
	switch x := e.(type) {
	case *exp.Expr:
		return c.Resolve(env, append(exp.Dyn{r.Lit}, x.Args...))
	}
	return r.Lit, nil
}
