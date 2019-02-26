package exp

// TypedUnresolver will only set enclosed type to the ref or expr to be resolved.
// This is useful for partial resolution phases where the type is known but the literal isn't.
type TypedUnresolver struct{ Type }

func (r TypedUnresolver) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	switch v := e.(type) {
	case *Ref:
		v.Type = r.Type
	case *Expr:
		v.Type = r.Type
	}
	return e, ErrUnres
}

// LitResilver will resolve to the enclosed literal. When not used as ref it resolves as dynamic
// starting with the enclosed literal followed by the expr arguments.
type LitResolver struct{ Lit }

func (r LitResolver) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	switch x := e.(type) {
	case *Expr:
		return c.Resolve(env, append(Dyn{r.Lit}, x.Args...), hint)
	}
	return r.Lit, nil
}
