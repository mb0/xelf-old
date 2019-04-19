package exp

import "github.com/mb0/xelf/cor"

// ErrUnres is returned by a resolver if the result is unresolved, but otherwise valid.
var ErrUnres = cor.StrError("unresolved")

// ErrExec is returned by a resolver if it cannot proceed because context exec is false.
var ErrExec = cor.StrError("not executed")

// Resolver is the common interface of all element resolvers.
type Resolver interface {
	// Resolve resolves el with a context, env, type hint and returns the result or an error.
	//
	// The passed in unresolved element is either a expression or symbol ref.
	//
	// A successful resolution returns a literal and no error.
	// If the type hint is not void, it is used to check or infer the element type.
	// When parts of the element could not be resolved it returns the special error ErrUnres,
	// and either the original element or if the context allows a partially resolved element.
	// If the resolution cannot proceed with execution it returns the special error ErrExec.
	// Any other error ends the whole resolution process.
	Resolve(c *Ctx, env Env, el El, hint Type) (El, error)
}

// Env is a scoped symbol environment used to define and lookup resolvers by symbol.
type Env interface {
	// Parent returns the parent environment or nil for the root environment.
	Parent() Env

	// Get looks for a resolver with symbol sym defined in this environments.
	// Implementation assume sym is not empty. Callers must ensure that sym is not empty.
	Get(sym string) Resolver

	// Supports returns whether the environment supports a special behaviour represented by x.
	Supports(x byte) bool
}

// Lookup returns a first resolver with symbol sym found in env or one of its ancestors.
// If sym starts with a known special prefix only the appropriate environments are called.
func Lookup(env Env, sym string) Resolver {
	return LookupSupports(env, sym, 0)
}

// LookupSupports looks up and returns a resolver that supports behaviour indicated by x or nil.
func LookupSupports(env Env, sym string, x byte) Resolver {
	for env != nil {
		if x == 0 || env.Supports(x) {
			r := env.Get(sym)
			if r != nil {
				return r
			}
		}
		env = env.Parent()
	}
	return nil
}

// GetSupports returns a resolver that supports behaviour indicated by x or nil.
func GetSupports(env Env, sym string, x byte) Resolver {
	env = Supports(env, x)
	if env != nil {
		return env.Get(sym)
	}
	return nil
}

// Supports returns an environment that supports behaviour indicated by x or nil.
func Supports(env Env, x byte) Env {
	for env != nil {
		if env.Supports(x) {
			return env
		}
		env = env.Parent()
	}
	return nil
}
