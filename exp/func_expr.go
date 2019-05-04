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

type FuncScope struct {
	DataScope
}

func (f *FuncScope) Get(s string) *Def {
	if s == "_" {
		s = ".0"
	}
	return f.DataScope.Get(s)
}

func (f *ExprBody) Resolve(c *Ctx, env Env, x *Call, hint Type) (El, error) {
	// build a parameter record from all arguments
	lo, err := ResolveFuncArgs(c, env, x)
	if err != nil {
		return x, err
	}
	// use the calling env to resolve parameters
	ps := x.Spec.Arg()
	keyed := make([]lit.Keyed, 0, len(ps))
	for i, p := range ps {
		a := lo.args[i]
		kl := lit.Keyed{p.Key(), nil}
		if len(a) == 0 { // can only be optional parameter; use zero value
			kl.Lit = lit.Zero(p.Type)
		} else {
			kl.Lit = a[0].(Lit)
		}
		if kl.Key == "" {
			// otherwise use a synthetic name
			kl.Key = fmt.Sprintf("arg%d", i)
		}
		keyed = append(keyed, kl)
	}
	s := DataScope{env, lit.Nil}
	if len(keyed) > 0 {
		s.Dot = &lit.Rec{Type: typ.Rec(ps), Keyr: lit.Keyr{List: keyed}}
	}
	// switch the function scope's parent to the declaration environment
	env = NewScope(&FuncScope{s})
	// and execute all body elements using the new scope
	var res El
	for _, e := range f.Els {
		var err error
		res, err = c.WithPart(false).Resolve(env, e, typ.Void)
		if err != nil {
			return x, err
		}
	}
	rt := x.Spec.Res()
	if rt == typ.Void {
		return rt, nil
	}
	return lit.Convert(res.(Lit), rt, 0)
}
