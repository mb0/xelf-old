package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ExprBody is the body for normal functions consisting of a list of expression elements.
type ExprBody struct {
	Els []El
}

func (f *ExprBody) WriteBfr(b bfr.Ctx) error {
	for _, el := range f.Els {
		b.WriteByte(' ')
		err := el.WriteBfr(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *ExprBody) ResolveCall(c *Ctx, env Env, fc *Call) (El, error) {
	// build a parameter object from all arguments
	param, err := lit.MakeObj(typ.Obj(fc.Type.FuncParams()))
	if err != nil {
		return fc.Expr, err
	}
	for i, a := range fc.Args {
		l, err := c.Resolve(env, a.Args[0])
		if err != nil {
			return fc.Expr, err
		}
		param.SetIdx(i, l.(Lit))
	}
	// create a function scope and set the parameter object
	s := funcScope{NewScope(env), param}
	// and execute all body elements using the new scope
	var res El
	for _, e := range f.Els {
		res, err = c.WithPart(false).Resolve(s, e)
		if err != nil {
			return fc.Expr, err
		}
	}
	return res, nil
}

type funcScope struct {
	*Scope
	Param lit.Obj
}

func (f funcScope) Get(s string) Resolver {
	if s == "" {
		return nil
	}
	if s[0] == '$' {
		l, err := lit.Select(f.Param, s[1:])
		if err != nil {
			return nil
		}
		return LitResolver{l}
	}
	return f.Scope.Get(s)
}
