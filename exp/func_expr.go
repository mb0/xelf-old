package exp

import (
	"fmt"

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
		// TODO allow implicit argument spread for the last param
		return nil, cor.Error("argument mismatch")
	}
	// use the calling env to resove parameters
	s := &ParamScope{NewScope(env), nil}
	if len(ps) > 0 {
		// initialize an empty dict obj
		o := &lit.DictObj{Type: typ.Obj(ps[:0])}
		o.List = make([]lit.Keyed, 0, len(ps))
		s.Param = o
		for i, a := range fc.Args {
			p := ps[i]
			el, err := c.Resolve(s, a.Args[0], p.Type)
			if err != nil {
				return fc.Expr, err
			}
			// ensure conversion to param type until hints are used everywhere
			l, err := lit.Convert(el.(Lit), p.Type, 0)
			if err != nil {
				return nil, err
			}
			name := a.Name
			if name == "" {
				// fall back to signature for empty parameter names
				name = p.Key()
				if name == "" {
					// otherwise use a synthetic name
					name = fmt.Sprintf("p_%d", i)
				}
				// TODO make sure the fallback name has no conflicts
				// with other parameter declarations.
			}
			// update parameters on each iteral so the next parameter can
			// refer to previous ones.
			o.List = append(o.List, lit.Keyed{name, l})
			// make new field accessible to following parameters
			o.Type.Fields = ps[:i+1]
		}
	}
	// switch the function scope's parent to the declaration environment
	s.parent = f.Env
	// and execute all body elements using the new scope
	var res El
	for _, e := range f.Els {
		var err error
		res, err = c.WithPart(false).Resolve(s, e, typ.Void)
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
