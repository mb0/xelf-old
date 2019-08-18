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

func (f *ExprBody) Resolve(p *Prog, env Env, c *Call, h typ.Type) (El, error) {
	// build a parameter record from all arguments
	_, err := ReslFuncArgs(p, env, c)
	if err != nil {
		return c, err
	}
	env = NewFuncScope(env, c)
	// and execute all body elements using the new scope
	for _, el := range f.Els {
		_, err = p.Resl(env, el, typ.Void)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

func (f *ExprBody) Execute(p *Prog, env Env, c *Call, h typ.Type) (El, error) {
	_, err := EvalFuncArgs(p, env, c)
	if err != nil {
		return c, err
	}
	env = NewFuncScope(env, c)
	// and execute all body elements using the new scope
	var res El
	for _, el := range f.Els {
		res, err = p.Eval(env, el, typ.Void)
		if err != nil {
			return c, err
		}
	}
	rt := c.Spec.Res()
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

func NewFuncScope(par Env, c *Call) *FuncScope {
	ps := c.Spec.Arg()
	keyed := make([]lit.Keyed, 0, len(ps))
	for i, p := range ps {
		a := c.Groups[i]
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
