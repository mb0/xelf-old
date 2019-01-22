package exp

import (
	"github.com/mb0/xelf/bfr"
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

// Sym holds symbol data and is used in both Val and Expr elements.
type Sym struct {
	Name string
	Rslv Resolver
	Type Type
}

// All language elements
type (
	// Type is a type as defined in package typ.
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

// Env is the scoped symbol environment used to define and lookup resolvers by symbol.
type Env interface {
	// Parent returns the parent environment or nil for the root environment.
	Parent() Env

	// Def defines a symbol resolver binding in this environment.
	Def(string, Resolver) error

	// Get looks for resolver with symbol sym in this or the parent environments.
	Get(sym string) Resolver
}

// Resolver can resolve an element given a context and environment.
type Resolver interface {
	Resolve(*Ctx, Env, El) (El, error)
}

// Ctx is the resolution context that defines the resolution level and collects information.
type Ctx struct {
	Exec bool

	Unres []El
}

func (x *Ref) String() string  { return bfr.String(x) }
func (x Dyn) String() string   { return bfr.String(x) }
func (x Tag) String() string   { return bfr.String(x) }
func (x Decl) String() string  { return bfr.String(x) }
func (x *Expr) String() string { return bfr.String(x) }

func (x *Ref) WriteBfr(b bfr.Ctx) error  { _, err := b.WriteString(x.Name); return err }
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
