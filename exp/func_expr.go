package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ExprBody is the body for normal functions consisting of a list of expression elements
// and its declaration envirnoment that is used for execution.
type ExprBody struct {
	Els []El
	Env Env
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

func (f *ExprBody) ResolveCall(c *Ctx, env Env, fc *Call, hint Type) (El, error) {
	// build a parameter object from all arguments
	ps := fc.Sig.FuncParams()
	if len(ps) != len(fc.Args) {
		return nil, cor.Error("argument mismatch")
	}
	// use the calling env in the function scope to resove parameters
	// we might want to reference other parameter types.
	fenv := funcScope{NewScope(env), nil}
	if len(ps) > 0 {
		var err error
		fenv.Param, err = lit.MakeObj(typ.Obj(ps))
		if err != nil {
			return fc.Expr, cor.Errorf("make param obj for %s: %w", fc.Type, err)
		}
		for i, a := range fc.Args {
			l, err := c.Resolve(fenv, a.Args[0], typ.Void)
			if err != nil {
				return fc.Expr, err
			}
			fenv.Param.SetIdx(i, l.(Lit))
		}
		// create a function scope and set the parameter object
	}
	// switch the function scope's parent to the declaration environment
	fenv.parent = f.Env
	// and execute all body elements using the new scope
	var res El
	for _, e := range f.Els {
		var err error
		res, err = c.WithPart(false).Resolve(fenv, e, typ.Void)
		if err != nil {
			return fc.Expr, err
		}
	}
	rt := fc.Sig.FuncResult()
	if rt == typ.Void {
		return rt, nil
	}
	return lit.Convert(res.(Lit), rt, 0)
}

type funcScope struct {
	*Scope
	Param lit.Obj
}

func (f funcScope) Supports(x byte) bool {
	return x == '$'
}

func (f funcScope) Get(s string) Resolver {
	if s[0] == '$' {
		l, err := lit.Select(f.Param, s[1:])
		if err != nil {
			return nil
		}
		return LitResolver{l}
	}
	return f.Scope.Get(s)
}