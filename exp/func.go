package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
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
	v, err := cor.Quote(f.String(), '"')
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

// FuncArgs matches arguments of x to the parameters of f and returns a layout or an error.
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
	if len(params) == 0 {
		if len(tags) > 0 {
			return nil, cor.Errorf("unexpected arguments %s", x)
		}
		return &Layout{}, nil
	}
	vari := isVariadic(params)
	var tagged bool
	args := make([][]El, len(params))
	for i, tag := range tags {
		idx := i
		if tag.Name == "" {
			if tagged {
				return nil,
					cor.Errorf("positional param after tag parameter in %s", x)
			}
			if idx >= len(args) {
				if vari {
					idx = len(args) - 1
					args[idx] = append(args[idx], tag.Args...)
					continue
				}
				return nil, cor.Errorf("unexpected arguments %s", x)
			}
		} else if tag.Name == "::" {
			if vari {
				idx = len(args) - 1
				args[idx] = append(args[idx], tag.Args...)
				continue
			}
			return nil, cor.Errorf("unexpected arguments %s", x)
		} else {
			tagged = true
			_, idx, err = x.Rslv.Typ().ParamByKey(tag.Key())
			if err != nil {
				return nil, err
			}
		}
		if len(args[idx]) > 0 {
			return nil, cor.Errorf("duplicate parameter %s", params[idx].Name)
		}
		args[idx] = tag.Args
	}
	for i, pa := range params {
		arg := args[i]
		if len(arg) == 0 {
			if pa.Opt() {
				continue
			}
			return nil, cor.Errorf("missing non optional parameter %s", pa.Name)
		}
		if len(arg) > 1 {

		}
	}
	return &Layout{x.Rslv.Arg(), args}, nil
}

func resolveListArr(c *Ctx, env Env, et typ.Type, args []El) (*lit.ListArr, error) {
	els, err := c.ResolveAll(env, args, et)
	if err != nil {
		return nil, err
	}
	res := make(lit.List, 0, len(els))
	for _, el := range els {
		l := el.(Lit)
		if et != typ.Any {
			l, err = lit.Convert(l, et, 0)
			if err != nil {
				return nil, err
			}
		}
		res = append(res, l)
	}
	return &lit.ListArr{et, res}, nil
}

func ResolveFuncArgs(c *Ctx, env Env, x *Expr) (*Layout, error) {
	lo, err := FuncArgs(x)
	if err != nil {
		return nil, err
	}
	params := x.Rslv.Arg()
	vari := isVariadic(params)
	for i, p := range params {
		a := lo.args[i]
		if len(a) == 0 { // skip; nothing to resolve
			continue
		}
		if i == len(params)-1 && vari {
			if len(a) > 1 {
				ll, err := resolveListArr(c, env, p.Type.Elem(), a)
				if err != nil {
					return nil, err
				}
				lo.args[i] = []El{ll}
				break
			}
			el, err := c.Resolve(env, a[0], typ.Void)
			if err != nil {
				return nil, err
			}
			l := el.(Lit)
			ll, err := lit.Convert(l, p.Type, 0)
			if err != nil {
				ll, err = resolveListArr(c, env, p.Type.Elem(), []El{el})
				if err != nil {
					return nil, err
				}
			}
			lo.args[i] = []El{ll}
			break // last iteration
		}
		if len(a) > 1 {
			return nil, cor.Errorf(
				"multiple arguments for non variadic parameter %s", p.Name)
		}
		el, err := c.Resolve(env, a[0], p.Type)
		if err != nil {
			return nil, err
		}
		l := el.(Lit)
		if p.Type != typ.Any {
			l, err = lit.Convert(l, p.Type, 0)
			if err != nil {
				return nil, err
			}
		}
		a[0] = l
	}
	return lo, err
}

func isVariadic(ps []typ.Param) bool {
	if len(ps) != 0 {
		switch ps[len(ps)-1].Type.Kind & typ.SlotMask {
		case typ.BaseList, typ.KindArr:
			return true
		}
	}
	return false
}
