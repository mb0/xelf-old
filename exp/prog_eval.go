package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

func Eval(env Env, el El) (El, error) { return NewProg().Eval(env, el, typ.Void) }

func (p *Prog) EvalAll(env Env, els []El, hint typ.Type) (res []El, err error) {
	return doAll(p, env, els, hint, (*Prog).Eval)
}

func (p *Prog) Eval(env Env, el El, h typ.Type) (_ El, err error) {
	if el == nil {
		return &Atom{Lit: typ.Void}, nil
	}
	switch v := el.(type) {
	case *Atom:
		// we can only resolve types
		return p.reslAtom(env, v, h)
	case *Sym:
		if v.Lit == nil {
			e, err := p.reslSym(env, v, h)
			if err != nil {
				return e, err
			}
			v = e.(*Sym)
		}
		if v.Lit == nil {
			return v, ErrUnres
		}
		return &Atom{Lit: v.Lit, Src: v.Src}, nil
	case *Tag:
		if v.El != nil {
			v.El, err = p.Eval(env, v.El, typ.Void)
		}
		return v, err
	case *Dyn:
		res, err := p.dynCall(env, v)
		if err != nil {
			return res, err
		}
		if c, ok := res.(*Call); ok {
			return c.Spec.Eval(p, env, c, h)
		}
		return res, nil
	case *Call:
		return v.Spec.Eval(p, env, v, h)
	}
	return el, cor.Errorf("unexpected expression %T %v", el, el)
}
