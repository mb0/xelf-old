package exp

import (
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

// Func is the common type for all function literals and implements both literal and resolver.
// It consists of a signature and body. A func is consider to a zero value if the body is nil,
// any other body value must be a valid function body. If the body implements bfr writer
// it is called for printing the body expressions.
// Resolution handles reference and delegates expression resolution to the body.
type Func struct {
	Sig
	Body FuncResolver
}

func (f *Func) IsZero() bool   { return f.Body == nil }
func (f *Func) String() string { return bfr.String(f) }
func (f *Func) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('(')
	err := f.Sig.WriteBfr(b)
	if err != nil {
		return err
	}
	if f.Body == nil {
		b.WriteString(" null")
	} else if v, ok := f.Body.(bfr.Writer); ok {
		err = v.WriteBfr(b)
		if err != nil {
			return err
		}
	} else {
		b.WriteString(" (() builtin)")
	}
	return b.WriteByte(')')
}

func (f *Func) MarshalJSON() ([]byte, error) {
	v, err := lex.Quote(f.String(), '"')
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (f *Func) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	if e.Typ() == typ.Sym {
		return f, nil
	}
	if x, ok := e.(*Expr); ok {
		if f.Body == nil {
			return e, ErrUnres
		}
		return f.Body.ResolveFunc(c, env, x, hint)
	}
	return nil, cor.Errorf("unexpected element %T", e)
}

// FuncResolver must be implemented by all function resolvers.
type FuncResolver interface {
	ResolveFunc(*Ctx, Env, *Expr, Type) (El, error)
}

var layoutArgs = []typ.Param{{Name: "args"}}

// FuncArgs matches arguments of x to the parameters of f and returns them or an error.
func FuncArgs(x *Expr) (*Layout, error) {
	lo, err := LayoutArgs(layoutArgs, x.Args)
	if err != nil {
		return nil, err
	}
	tags, err := lo.Tags(0)
	if err != nil {
		return nil, err
	}
	params := x.Rslv.Arg()
	args := make([][]El, len(params))
	for i, tag := range tags {
		if tag.Name == "" {
			if i < len(args) {
				args[i] = tag.Args
			} else {
				i = len(args) - 1
				args[i] = append(args[i], tag.Args...)
			}
		} else {
			key := strings.ToLower(tag.Name[1:])
			_, idx, err := x.Rslv.Typ().ParamByKey(key)
			if err != nil {
				return nil, err
			}
			args[idx] = tag.Args
		}
	}
	return &Layout{x.Rslv.Arg(), args}, nil
}
