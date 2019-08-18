package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

func (c *Ctx) EvalAll(env Env, els []El, hint typ.Type) (res []El, err error) {
	return doAll(c, env, els, hint, (*Ctx).Eval)
}

func (c *Ctx) Eval(env Env, el El, hint typ.Type) (_ El, err error) {
	if el == nil {
		return &Atom{Lit: typ.Void}, nil
	}
	switch v := el.(type) {
	case *Atom:
		// we can only resolve types
		return c.reslAtom(env, v, hint)
	case *Sym:
		return c.evalSym(env, v, hint)
	case *Named:
		if v.El != nil {
			if d, ok := v.El.(*Dyn); ok {
				els, err := c.EvalAll(env, d.Els, typ.Void)
				if err != nil {
					return nil, err
				}
				d.Els = els
			} else {
				el, err := c.Eval(env, v.El, typ.Void)
				if err != nil {
					return nil, err
				}
				v.El = el
			}
		}
		return v, nil
	case *Dyn:
		return c.EvalDyn(env, v, hint)
	case *Call:
		return v.Spec.Execute(c, env, v, hint)
	}
	return el, cor.Errorf("unexpected expression %T %v", el, el)
}

func (c *Ctx) EvalDyn(env Env, d *Dyn, h typ.Type) (El, error) {
	res, err := c.dynCall(env, d)
	if err != nil {
		return res, err
	}
	if call, ok := res.(*Call); ok {
		return call.Spec.Execute(c, env, call, h)
	}
	return res, nil
}

func (c *Ctx) evalSym(env Env, s *Sym, hint typ.Type) (El, error) {
	e, err := c.reslSym(env, s, hint)
	if err != nil {
		return e, err
	}
	s = e.(*Sym)
	if s.Lit == nil {
		return s, ErrUnres
	}
	a := &Atom{Lit: s.Lit, Src: s.Src}
	return c.checkHint(hint, a)
}
