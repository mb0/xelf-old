package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ErrUnres is returned by a resolver if the result is unresolved, but otherwise valid.
var ErrUnres = cor.StrError("unresolved")

// ErrExec is returned by a resolver if it cannot proceed because context exec is false.
var ErrExec = cor.StrError("not executed")

// Def represents a definition in an environment.
type Def struct {
	// Type is the resolved definition type.
	typ.Type
	// Lit is the evaluated literal. The literal must be convertible to the definiton type.
	Lit lit.Lit
}

func NewDef(l lit.Lit) *Def { return &Def{Type: l.Typ(), Lit: l} }

// Env is a scoped symbol environment used to define and lookup resolvers by symbol.
type Env interface {
	// Parent returns the parent environment or nil for the root environment.
	Parent() Env

	// Get looks for a definition with symbol sym defined in this environments.
	// Implementation assume sym is not empty. Callers must ensure that sym is not empty.
	Get(sym string) *Def

	// Supports returns whether the environment supports a special behaviour represented by x.
	Supports(x byte) bool
}

// Lookup returns a first resolver with symbol sym found in env or one of its ancestors.
// If sym starts with a known special prefix only the appropriate environments are called.
func Lookup(env Env, sym string) *Def {
	return LookupSupports(env, sym, 0)
}

// LookupSupports looks up and returns a resolver that supports behaviour indicated by x or nil.
func LookupSupports(env Env, sym string, x byte) *Def {
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
