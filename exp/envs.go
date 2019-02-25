package exp

import "github.com/mb0/xelf/cor"

var (
	// ErrNoDefEnv is returned when the environment cannot define any resolvers
	ErrNoDefEnv = cor.StrError("not a definition env")
	// ErrRedefine is returned when the symbol is redefined in the same scope.
	ErrRedefine = cor.StrError("redefined symbol")
)

// LookupFunc is a simple resolver lookup function used by builtins and libraries.
type LookupFunc = func(sym string) Resolver

// Builtin is an environment based on a slice of simple resolver lookup functions.
//
// The lookup functions are check from start to finish returning the first result.
// A builtin environment has no parent and cannot define resolvers.
type Builtin []LookupFunc

// Parent always returns nil for the built-in lookups
func (Builtin) Parent() Env { return nil }

// Supports returns always false for the built-in lookups
func (Builtin) Supports(byte) bool { return false }

// Def always returns ErrNoDefEnv for the built-in lookups
func (Builtin) Def(string, Resolver) error { return ErrNoDefEnv }

// Get returns a resolver for the given sym
func (b Builtin) Get(sym string) Resolver {
	for _, f := range b {
		r := f(sym)
		if r != nil {
			return r
		}
	}
	return nil
}

// Scope is a child environment based on a map of resolvers.
type Scope struct {
	parent Env
	decl   map[string]Resolver
}

// NewScope returns a child scope with the given parent environment.
func NewScope(parent Env) *Scope {
	return &Scope{parent, make(map[string]Resolver)}
}

// Parent returns the parent environment.
func (c *Scope) Parent() Env {
	return c.parent
}

// Supports returns always false for simple scopes
func (*Scope) Supports(byte) bool { return false }

// Def defines a symbol resolver binding for s and d or returns an error.
func (c *Scope) Def(s string, d Resolver) error {
	if _, ok := c.decl[s]; ok {
		return ErrRedefine
	}
	c.decl[s] = d
	return nil
}

// Get returns the resolver with symbol s or nil. If no resolver is found in this scope
// the parent scope is called.
func (c *Scope) Get(s string) Resolver {
	d, ok := c.decl[s]
	if ok {
		return d
	}
	if c.parent == nil {
		return nil
	}
	return c.parent.Get(s)
}

// Supported returns env or a parent environment that supports behaviour indicated by x or nil.
func Supported(env Env, x byte) Env {
	for env != nil {
		if env.Supports(x) {
			return env
		}
		env = env.Parent()
	}
	return nil
}
