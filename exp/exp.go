package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// El is the common interface of all language elements.
type El interface {
	// WriteBfr writes the element to a bfr.Ctx.
	WriteBfr(*bfr.Ctx) error
	// String returns the xelf representation as string.
	String() string
	// Typ returns the type
	Typ() typ.Type
	// Source returns the source position if available
	Source() lex.Src
}

// All the language elements
type (
	// Atom is a literal or type with source source offsets as returned by the parser.
	Atom struct {
		Lit lit.Lit
		lex.Src
	}

	// Sym is an identifier, that refers to a definition.
	Sym struct {
		Name string
		// Type is the resolved type of lit in this context or void.
		Type typ.Type
		// Lit is the resolved literal or nil. Conversion may be required.
		Lit lit.Lit
		lex.Src
	}

	// Dyn is an expression with an undefined specification, that has to be determined.
	Dyn struct {
		Els []El
		lex.Src
	}

	// Named is a tag or declaration; its meaning is determined by the parent's specification.
	Named struct {
		Name string
		El   El
		lex.Src
	}

	// Call is an expression with a defined specification.
	Call struct {
		// Type is the instantiated and possibly resolved spec type in this context or void
		Type typ.Type
		// Spec is the form or func specification
		Spec *Spec
		Args []El
		lex.Src
	}
)

func (x *Atom) Typ() typ.Type  { return x.Lit.Typ() }
func (x *Sym) Typ() typ.Type   { return typ.Sym }
func (x *Dyn) Typ() typ.Type   { return typ.Dyn }
func (x *Call) Typ() typ.Type  { return typ.Call }
func (x *Named) Typ() typ.Type { return typ.Named }

func (x *Atom) String() string  { return bfr.String(x) }
func (x *Sym) String() string   { return x.Name }
func (x *Dyn) String() string   { return bfr.String(x) }
func (x *Named) String() string { return bfr.String(x) }
func (x *Call) String() string  { return bfr.String(x) }

func (x *Atom) WriteBfr(b *bfr.Ctx) error { return x.Lit.WriteBfr(b) }
func (x *Sym) WriteBfr(b *bfr.Ctx) error  { return b.Fmt(x.Name) }
func (x *Dyn) WriteBfr(b *bfr.Ctx) error  { return writeExpr(b, "", x.Els) }
func (x *Named) WriteBfr(b *bfr.Ctx) error {
	if x.El == nil {
		return b.Fmt(x.Name)
	}
	if x.Name == "" || x.Name[0] != ':' {
		if d := x.Dyn(); d != nil {
			return writeExpr(b, x.Name, d.Els)
		}
	}
	if x.Name != "" {
		b.WriteString(x.Name)
		b.WriteByte(' ')
	}
	return x.El.WriteBfr(b)
}
func (x *Call) WriteBfr(b *bfr.Ctx) error {
	name := x.Spec.Key()
	if name == "" {
		name = x.Spec.String()
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
func (x *Sym) Key() string { return cor.Keyed(x.Name) }

// Res returns the result type or void.
func (x *Call) Res() typ.Type {
	if isSig(x.Type) {
		return x.Type.Params[len(x.Type.Params)-1].Type
	}
	return x.Spec.Res()
}
func NewNamed(name string, els ...El) *Named {
	if len(els) == 0 {
		return &Named{Name: name, El: nil}
	}
	if len(els) > 1 {
		return &Named{Name: name, El: &Dyn{Els: els}}
	}
	return &Named{Name: name, El: els[0]}
}

func (x *Named) Key() string { return cor.Keyed(x.Name) }
func (x *Named) IsTag() bool { return x.Name != "" && x.Name[0] == ':' }

func (x *Named) Args() []El {
	if x.El == nil {
		return nil
	}
	if d, ok := x.El.(*Dyn); ok {
		return d.Els
	}
	return []El{x.El}
}
func (x *Named) Arg() El {
	if d, ok := x.El.(*Dyn); ok && len(d.Els) != 0 {
		return d.Els[0]
	}
	return x.El
}
func (x *Named) Dyn() *Dyn {
	if d, ok := x.El.(*Dyn); ok {
		return d
	}
	return nil
}
