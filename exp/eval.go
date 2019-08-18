package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

func (p *Prog) EvalAll(env Env, els []El, hint typ.Type) (res []El, err error) {
	return doAll(p, env, els, hint, (*Prog).Eval)
}

func (p *Prog) Eval(env Env, el El, hint typ.Type) (_ El, err error) {
	if el == nil {
		return &Atom{Lit: typ.Void}, nil
	}
	switch v := el.(type) {
	case *Atom:
		// we can only resolve types
		return p.reslAtom(env, v, hint)
	case *Sym:
		return p.evalSym(env, v, hint)
	case *Named:
		if v.El != nil {
			if d, ok := v.El.(*Dyn); ok {
				els, err := p.EvalAll(env, d.Els, typ.Void)
				if err != nil {
					return nil, err
				}
				d.Els = els
			} else {
				el, err := p.Eval(env, v.El, typ.Void)
				if err != nil {
					return nil, err
				}
				v.El = el
			}
		}
		return v, nil
	case *Dyn:
		return p.EvalDyn(env, v, hint)
	case *Call:
		return v.Spec.Execute(p, env, v, hint)
	}
	return el, cor.Errorf("unexpected expression %T %v", el, el)
}

func (p *Prog) EvalDyn(env Env, d *Dyn, h typ.Type) (El, error) {
	res, err := p.dynCall(env, d)
	if err != nil {
		return res, err
	}
	if c, ok := res.(*Call); ok {
		return c.Spec.Execute(p, env, c, h)
	}
	return res, nil
}

func (p *Prog) evalSym(env Env, s *Sym, hint typ.Type) (El, error) {
	e, err := p.reslSym(env, s, hint)
	if err != nil {
		return e, err
	}
	s = e.(*Sym)
	if s.Lit == nil {
		return s, ErrUnres
	}
	a := &Atom{Lit: s.Lit, Src: s.Src}
	return p.checkHint(hint, a)
}
