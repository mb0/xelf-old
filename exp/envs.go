package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

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

// Supports returns true for built-in schema type lookups
func (Builtin) Supports(x byte) bool { return x == '~' }

// Def always returns ErrNoDefEnv for the built-in lookups
func (Builtin) Def(string, Resolver) error { return ErrNoDefEnv }

// Get returns a resolver for the given sym
func (b Builtin) Get(sym string) Resolver {
	// check schema prefix
	schema := sym[0] == '~'
	if schema {
		sym = sym[1:]
	}
	t, err := typ.ParseSym(sym, nil)
	if err == nil {
		return LitResolver{t}
	}
	if schema {
		return nil
	}
	// lookup type
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
	return &Scope{parent: parent}
}

func (c *Scope) Parent() Env { return c.parent }

// Supports returns always false for simple scopes.
func (*Scope) Supports(byte) bool { return false }

// Def defines a symbol resolver binding for s and d or returns an error.
func (c *Scope) Def(s string, d Resolver) error {
	if c.decl == nil {
		c.decl = make(map[string]Resolver, 8)
	} else if _, ok := c.decl[s]; ok {
		return ErrRedefine
	}
	c.decl[s] = d
	return nil
}

// Get returns a resolver with symbol s defined in this scope or nil.
func (c *Scope) Get(s string) Resolver {
	d, ok := c.decl[s]
	if ok {
		return d
	}
	return nil
}

// DataScope is a child environment that supports relative paths and is backed by a literal
type DataScope struct {
	Par Env
	Dot Lit
}

// NewDataScope returns a data scope with the given parent environment.
func NewDataScope(parent Env) *DataScope {
	return &DataScope{Par: parent}
}

func (c *DataScope) Parent() Env { return c.Par }

// Supports returns true for '.', false otherwise.
func (*DataScope) Supports(x byte) bool { return x == '.' }

// Def always returns an error for data scopes.
func (c *DataScope) Def(s string, d Resolver) error { return ErrNoDefEnv }

// Get returns a literal resolver for the relative path s or nil.
func (c *DataScope) Get(s string) Resolver {
	if s[0] == '.' {
		l, err := lit.Select(c.Dot, s[1:])
		if err != nil {
			return nil
		}
		return LitResolver{l}
	}
	return nil
}

// ParamScope wraps a scope and provides parameter resolution.
// It is also used as part of the prog scope and for signature definitions.
type ParamScope struct {
	*Scope
	Param *lit.DictObj
}

func (*ParamScope) Supports(x byte) bool { return x == '$' }

func (p *ParamScope) Get(s string) Resolver {
	if s[0] == '$' {
		l, err := lit.Select(p.Param, s[1:])
		if err != nil {
			return nil
		}
		return LitResolver{l}
	}
	return p.Scope.Get(s)
}

// ProgScope wraps a param scope and provides global result resolution.
type ProgScope struct {
	ParamScope
	Result *lit.Dict
}

func (*ProgScope) Supports(x byte) bool { return x == '$' || x == '/' }

func (p *ProgScope) Get(s string) Resolver {
	if s[0] == '/' {
		l, err := lit.Select(p.Result, s[1:])
		if err != nil {
			return nil
		}
		return LitResolver{l}
	}
	return p.ParamScope.Get(s)
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
