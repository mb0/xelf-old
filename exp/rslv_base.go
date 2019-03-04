package exp

import "github.com/mb0/xelf/typ"

// TypedUnresolver will only set enclosed type to a ref to be resolved.
// This is useful for partial resolution phases where the type is known but the literal isn't.
type TypedUnresolver struct{ Type }

func (r TypedUnresolver) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	if e.Typ().Kind == typ.ExpSym {
		e.(*Sym).Type = r.Type
	}
	return e, ErrUnres
}

// LitResilver will resolve to the enclosed literal.
type LitResolver struct{ Lit }

func (r LitResolver) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	return r.Lit, nil
}
