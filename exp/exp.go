package exp

import (
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// El is the common interface for all language elements.
type El interface {
	// WriteBfr writes the element to a bfr.Ctx.
	WriteBfr(bfr.Ctx) error
	// String returns the xelf representation as string.
	String() string
}

// Sym holds symbol data and is used by expressions and symbol references.
type Sym struct {
	Name string
	Rslv Resolver
	Type Type
	key  string
}

func (s *Sym) Key() string {
	if s.key == "" {
		s.key = strings.ToLower(s.Name)
	}
	return s.key
}

// Lookup returns the cached resolver, if nil it queries and caches if from env.
func (s *Sym) Lookup(env Env) Resolver {
	if s.Rslv == nil {
		s.Rslv = env.Get(s.Key())
	}
	return s.Rslv
}

// All language elements
type (
	// Type is a type as defined in package typ. It also implements Lit.
	Type = typ.Type

	// Lit is a literal as defined in package lit.
	Lit = lit.Lit

	// Ref is a unresolved symbol that refers to an element.
	Ref struct {
		Sym
	}
	// Dyn represents an unresolved dynamic expression where the spec is not yet resolved.
	Dyn []El

	// Tag is a tag element; its meaning is determined by the parent expression spec.
	Tag struct {
		Name string
		Args []El
	}

	// Decl is a declaration element; its meaning is determined by the parent expression spec.
	Decl struct {
		Name string
		Args []El
	}

	// Expr is unresolved expression with a resolved spec.
	Expr struct {
		Sym
		Args []El
	}
)

// Env is a scoped symbol environment used to define and lookup resolvers by symbol.
type Env interface {
	// Parent returns the parent environment or nil for the root environment.
	Parent() Env

	// Def defines a symbol resolver binding in this environment.
	Def(string, Resolver) error

	// Get looks for resolver with symbol sym in this or the parent environments.
	Get(sym string) Resolver
}

// ErrUnres is a special error value that is returned by a resolver when
// the result is unresolved but otherwise valid.
var ErrUnres = cor.StrError("unresolved")

// Resolver is the common interface of all element resolvers.
type Resolver interface {
	// Resolve resolves el with a context and environment and returns the result or an error.
	//
	// The passed in unresolved element is either a expression, a type ref or symbol ref.
	//
	// A successful resolution returns a literal and no error.
	// When parts of the element could not be resolved it returns the special error ErrUnres,
	// and either the original element or if the context allows a partially resolved element.
	// Any other error ends the whole resolution process.
	Resolve(c *Ctx, enc Env, el El) (El, error)
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

func (x *Ref) String() string  { return bfr.String(x) }
func (x Dyn) String() string   { return bfr.String(x) }
func (x Tag) String() string   { return bfr.String(x) }
func (x Decl) String() string  { return bfr.String(x) }
func (x *Expr) String() string { return bfr.String(x) }

func (x *Ref) WriteBfr(b bfr.Ctx) error  { return b.Fmt(x.Name) }
func (x Dyn) WriteBfr(b bfr.Ctx) error   { return writeExpr(b, "", x) }
func (x Tag) WriteBfr(b bfr.Ctx) error   { return writeExpr(b, x.Name, x.Args) }
func (x Decl) WriteBfr(b bfr.Ctx) error  { return writeExpr(b, x.Name, x.Args) }
func (x *Expr) WriteBfr(b bfr.Ctx) error { return writeExpr(b, x.Name, x.Args) }

func writeExpr(b bfr.Ctx, name string, args []El) error {
	b.WriteByte('(')
	if name != "" {
		b.WriteString(name)
		b.WriteByte(' ')
	}
	for i, x := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		err := x.WriteBfr(b)
		if err != nil {
			return err
		}
	}
	return b.WriteByte(')')
}
