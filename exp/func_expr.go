package exp

import (
	"fmt"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ExprBody is the body for normal functions consisting of a list of expression elements
// and its declaration environment that is used for execution.
type ExprBody struct {
	Els []El
	Env Env
}

func (f *ExprBody) WriteBfr(b *bfr.Ctx) error {
	for _, el := range f.Els {
		b.WriteByte(' ')
		err := el.WriteBfr(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *ExprBody) Resolve(c *Ctx, env Env, x *Call, hint typ.Type) (El, error) {
	// build a parameter record from all arguments
	_, err := ReslFuncArgs(c, env, x)
	if err != nil {
		return x, err
	}
	env = NewFuncScope(env, x)
	// and execute all body elements using the new scope
	for _, e := range f.Els {
		_, err = c.WithPart(false).Resl(env, e, typ.Void)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

func (f *ExprBody) Execute(c *Ctx, env Env, x *Call, hint typ.Type) (El, error) {
	_, err := EvalFuncArgs(c, env, x)
	if err != nil {
		return x, err
	}
	env = NewFuncScope(env, x)
	// and execute all body elements using the new scope
	var res El
	for _, e := range f.Els {
		res, err = c.WithPart(false).Eval(env, e, typ.Void)
		if err != nil {
			return x, err
		}
	}
	rt := x.Spec.Res()
	if rt == typ.Void {
		return &Atom{Lit: rt}, nil
	}
	a := res.(*Atom)
	a.Lit, err = lit.Convert(a.Lit, rt, 0)
	if err != nil {
		return nil, err
	}
	return a, nil
}

type FuncScope struct {
	DataScope
}

func (f *FuncScope) Get(s string) *Def {
	if s == "_" {
		s = ".0"
	}
	return f.DataScope.Get(s)
}

func NewFuncScope(par Env, x *Call) *FuncScope {
	ps := x.Spec.Arg()
	keyed := make([]lit.Keyed, 0, len(ps))
	for i, p := range ps {
		a := x.Groups[i]
		kl := lit.Keyed{p.Key(), lit.Zero(p.Type)}
		if len(a) > 0 {
			if at, ok := a[0].(*Atom); ok {
				kl.Lit = at.Lit
			}
		}
		if kl.Key == "" {
			kl.Key = fmt.Sprintf("arg%d", i)
		}
		keyed = append(keyed, kl)
	}
	s := DataScope{par, Def{Lit: lit.Nil}}
	if len(keyed) > 0 {
		s.Type = typ.Rec(ps)
		s.Lit = &lit.Rec{Type: s.Type, Dict: lit.Dict{List: keyed}}
	}
	return &FuncScope{s}
}
