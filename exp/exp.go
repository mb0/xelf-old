package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// El is the common interface for all language elements.
type El interface {
	// WriteBfr writes the element to a bfr.Ctx.
	WriteBfr(*bfr.Ctx) error
	// String returns the xelf representation as string.
	String() string
	// Typ returns the type
	Typ() Type
}

// All language elements
type (
	// Type is a type as defined in package typ. It also implements Lit.
	Type = typ.Type

	// Lit is a literal as defined in package lit.
	Lit = lit.Lit

	// Sym is a unresolved symbol that refers to an element.
	Sym struct {
		Name string
		lex.Pos
		// Type is partially resolve result type
		Type Type
	}

	// Raw is an scanned but unparsed token
	Raw lex.Tree

	// Dyn is a unresolved expression where a resolver has to be determined.
	Dyn []El

	// Named is a tag or declaration; its meaning is determined by the parent's resolver.
	Named struct {
		Name string
		lex.Pos
		El
	}

	// Expr is unresolved expression with  resolver is known.
	Expr struct {
		Rslv ExprResolver
		Args []El
		// Type is partially resolve result type
		Type Type
	}
)

func (*Sym) Typ() Type    { return typ.Sym }
func (*Raw) Typ() Type    { return typ.Dyn }
func (Dyn) Typ() Type     { return typ.Dyn }
func (x *Expr) Typ() Type { return x.Rslv.Typ() }
func (x *Named) Typ() Type {
	if x == nil {
		return typ.Void
	}
	if x.Name == "" || x.Name[0] == ':' {
		return typ.Tag
	}
	return typ.Decl
}

// Env is a scoped symbol environment used to define and lookup resolvers by symbol.
type Env interface {
	// Parent returns the parent environment or nil for the root environment.
	Parent() Env

	// Def defines a symbol resolver binding in this environment.
	Def(string, Resolver) error

	// Get looks for a resolver with symbol sym defined in this environments.
	// Implementation assume sym is not empty. Callers must ensure that sym is not empty.
	Get(sym string) Resolver

	// Supports returns whether the environment supports a special behaviour represented by x.
	Supports(x byte) bool
}

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

// DynResolver is a special resolver for dynamic expressions.
type DynResolver interface {
	ResolveDyn(c *Ctx, env Env, d Dyn, hint Type) (El, error)
}

// ExprResolver is the common interface for both function and form resolvers.
type ExprResolver interface {
	Lit
	Resolver
	Key() string
	Arg() []typ.Param
	Res() Type
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

	// Dyn is a configurable resolver for dynamic expressions. A default resolver is
	// used if this field is nil.
	Dyn DynResolver
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

func (x *Sym) String() string   { return x.Name }
func (x *Raw) String() string   { return bfr.String((*lex.Tree)(x)) }
func (x Dyn) String() string    { return bfr.String(x) }
func (x *Named) String() string { return bfr.String(x) }
func (x *Expr) String() string  { return bfr.String(x) }

func (x *Sym) WriteBfr(b *bfr.Ctx) error { return b.Fmt(x.Name) }
func (x *Raw) WriteBfr(b *bfr.Ctx) error { return (*lex.Tree)(x).WriteBfr(b) }
func (x Dyn) WriteBfr(b *bfr.Ctx) error  { return writeExpr(b, "", x) }
func (x *Named) WriteBfr(b *bfr.Ctx) error {
	if x.El == nil {
		return b.Fmt(x.Name)
	}
	if d := x.Dyn(); d != nil {
		return writeExpr(b, x.Name, d)
	}
	if x.Name != "" {
		b.WriteString(x.Name)
		b.WriteByte(' ')
	}
	return x.El.WriteBfr(b)
}
func (x *Expr) WriteBfr(b *bfr.Ctx) error {
	t := x.Rslv.Typ()
	name := t.Key()
	if name == "" {
		name = t.String()
	}
	return writeExpr(b, name, x.Args)
}

func writeExpr(b *bfr.Ctx, name string, args []El) error {
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
func (x *Sym) Key() string   { return cor.Keyed(x.Name) }
func (x *Named) Key() string { return cor.Keyed(x.Name) }

func (x *Raw) Dyn() Dyn {
	if x.Tok != '(' {
		return nil
	}
	d := make(Dyn, 0, len(x.Seq))
	for _, c := range x.Seq {
		d = append(d, (*Raw)(c))
	}
	return d
}

func (x *Named) Args() []El {
	if x.El == nil {
		return nil
	}
	if d, ok := x.El.(Dyn); ok {
		return d
	}
	return []El{x.El}
}
func (x *Named) Arg() El {
	if d, ok := x.El.(Dyn); ok && len(d) != 0 {
		return d[0]
	}
	return x.El
}
func (x *Named) Dyn() Dyn {
	if d, ok := x.El.(Dyn); ok {
		return d
	}
	return nil
}
