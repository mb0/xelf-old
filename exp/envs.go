package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var (
	// ErrRedefine is returned when the symbol is redefined in the same scope.
	ErrRedefine = cor.StrError("redefined symbol")
)

// LookupFunc is a simple spec lookup function used by builtins and libraries.
type LookupFunc = func(sym string) *Spec

// Builtin is an environment based on a slice of simple resolver lookup functions.
//
// The lookup functions are check from start to finish returning the first result.
// A builtin environment has no parent and cannot define resolvers.
type Builtin []LookupFunc

// Parent always returns nil for the built-in lookups
func (b Builtin) Parent() Env { return nil }

// Supports returns true for built-in schema type lookups
func (b Builtin) Supports(x byte) bool { return x == '~' }

// Get returns a resolver for the given sym
func (b Builtin) Get(sym string) *Def {
	// check schema prefix
	schema := sym[0] == '~'
	if schema {
		sym = sym[1:]
		t, err := typ.ParseSym(sym, nil)
		if err == nil {
			return NewDef(t)
		}
	}
	// lookup type
	for _, f := range b {
		r := f(sym)
		if r != nil {
			return NewDef(r)
		}
	}
	return nil
}

// Scope is a child environment based on a map of resolvers.
type Scope struct {
	parent Env
	decl   map[string]*Def
}

// NewScope returns a child scope with the given parent environment.
func NewScope(parent Env) *Scope {
	return &Scope{parent: parent}
}

func (s *Scope) Parent() Env { return s.parent }

// Supports returns always false for simple scopes.
func (s *Scope) Supports(byte) bool { return false }

// Def defines a symbol resolver binding for s and d or returns an error.
func (s *Scope) Def(sym string, d *Def) error {
	if s.decl == nil {
		s.decl = make(map[string]*Def, 8)
	} else if _, ok := s.decl[sym]; ok {
		return ErrRedefine
	}
	s.decl[sym] = d
	return nil
}

// Get returns a resolver with symbol s defined in this scope or nil.
func (s *Scope) Get(sym string) *Def {
	d, ok := s.decl[sym]
	if ok {
		return d
	}
	return nil
}

// DataScope is a child environment that supports relative paths and is backed by a literal
type DataScope struct {
	Par Env
	Def
}

// NewDataScope returns a data scope with the given parent environment.
func NewDataScope(parent Env) *DataScope {
	return &DataScope{Par: parent}
}

func (ds *DataScope) Parent() Env { return ds.Par }

// Supports returns true for '.', false otherwise.
func (ds *DataScope) Supports(x byte) bool { return x == '.' }

// Get returns a literal resolver for the relative path s or nil.
func (ds *DataScope) Get(s string) *Def {
	if s[0] == '.' {
		if ds.Lit != nil {
			l, err := lit.Select(ds.Lit, s[1:])
			if err != nil {
				return nil
			}
			return NewDef(l)
		}
		if ds.Type != typ.Void {
			t, err := typ.Select(ds.Type, s[1:])
			if err != nil {
				return nil
			}
			return &Def{Type: t}
		}
	}
	return nil
}

// ParamEnv provides parameter resolution.
type ParamEnv struct {
	Par   Env
	Param lit.Lit
}

func (ps *ParamEnv) Parent() Env          { return ps.Par }
func (ps *ParamEnv) Supports(x byte) bool { return x == '$' }
func (ps *ParamEnv) Get(s string) *Def {
	if s[0] != '$' {
		return nil
	}
	l, err := lit.Select(ps.Param, s[1:])
	if err != nil {
		return nil
	}
	return NewDef(l)
}

type ParamReslEnv struct {
	Par Env
	Ctx *typ.Ctx
	Map map[string]typ.Type
}

func NewParamReslEnv(p Env, c *typ.Ctx) *ParamReslEnv {
	return &ParamReslEnv{Par: p, Ctx: c, Map: make(map[string]typ.Type)}
}

func (ps *ParamReslEnv) Parent() Env          { return ps.Par }
func (ps *ParamReslEnv) Supports(x byte) bool { return x == '$' }
func (ps *ParamReslEnv) Get(s string) *Def {
	if s[0] != '$' {
		return nil
	}
	key := s[1:]
	t, ok := ps.Map[key]
	if !ok {
		t := ps.Ctx.New()
		ps.Map[key] = t
	}
	return &Def{Type: t}
}

// ProgEnv provides global result resolution.
type ProgEnv struct {
	Par    Env
	Result *lit.Dict
}

func (ps *ProgEnv) Parent() Env          { return ps.Par }
func (ps *ProgEnv) Supports(x byte) bool { return x == '/' }
func (ps *ProgEnv) Get(s string) *Def {
	if s[0] == '/' {
		l, err := lit.Select(ps.Result, s[1:])
		if err != nil {
			return nil
		}
		return NewDef(l)
	}
	return nil
}
