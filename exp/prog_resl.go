package exp

import (
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func Resl(env Env, el El) (El, error) { return NewProg().Resl(env, el, typ.Void) }

// ReslAll resolves all element or returns the first error.
func (p *Prog) ReslAll(env Env, els []El, h typ.Type) (res []El, err error) {
	return doAll(p, env, els, h, (*Prog).Resl)
}

// Resolve resolves x within env and returns the result or an error.
//
// This method will not resolve any element itself but instead tries to look up an applicable
// resolver in the environment. If it cannot find a resolver it will add the element to the
// context's unresolved slice.
// The resolver implementations usually use this method either directly or indirectly to resolve
// arguments, which are then again added to the unresolved elements when appropriate.
func (p *Prog) Resl(env Env, el El, h typ.Type) (_ El, err error) {
	switch v := el.(type) {
	case *Atom:
		return p.reslAtom(env, v, h)
	case *Sym:
		return p.reslSym(env, v, h)
	case *Tag:
		x, err := p.Resl(env, v.El, h)
		if err != nil {
			return el, err
		}
		v.El = x
	case *Dyn:
		res, err := p.dynCall(env, v)
		if err != nil {
			return v, err
		}
		if call, ok := res.(*Call); ok {
			return call.Spec.Resl(p, env, call, h)
		}
		return res, nil
	case *Call:
		return v.Spec.Resl(p, env, v, h)
	}
	return el, nil
}

func Ignore(src lex.Src) (El, error) {
	return &Atom{Lit: typ.Void, Src: src}, ErrVoid
}
func (p *Prog) reslAtom(env Env, a *Atom, hint typ.Type) (El, error) {
	switch a.Typ().Kind & typ.MaskRef {
	case typ.KindTyp: // resolve type references
		err := p.reslType(env, a)
		if err != nil {
			return a, err
		}
		//case typ.KindForm, typ.FormFunc:
		//	err := c.reslType(env, v)
		//	if err != nil {
		//		return v, err
		//	}
	}
	return p.checkHint(hint, a)
}

func (p *Prog) reslSym(env Env, s *Sym, hint typ.Type) (El, error) {
	switch s.Name[0] {
	case '.':
		err := p.reslDot(env, s)
		if err != nil {
			return s, err
		}
	case '$', '/':
		err := p.reslAbs(env, s)
		if err != nil {
			return s, err
		}
	default:
		def := Lookup(env, s.Name)
		if def == nil {
			p.Unres = append(p.Unres, s)
			return s, ErrUnres
		}
		s.Type = def.Type
		s.Lit = def.Lit
	}
	return p.checkHint(hint, s)
}

func (p *Prog) reslType(env Env, a *Atom) error {
	// last type is t or the element type for container types
	at := a.Lit.(typ.Type)
	t := at.Last()
	if t.Resolved() {
		return nil
	}
	var d *Def
	key := t.Key()
	switch t.Kind & typ.MaskRef {
	case typ.KindSch, typ.KindBits, typ.KindEnum, typ.KindObj:
		key = "~" + key
		d = LookupSupports(env, key, '~')
	case typ.KindRef:
		idx, path := strings.IndexByte(key, '.'), ""
		if idx > 0 {
			key, path = key[:idx], key[idx+1:]
		}
		sym := &Sym{Name: key}
		_, err := p.reslSym(env, sym, typ.Void)
		if err != nil {
			p.Unres = append(p.Unres, a)
			return ErrUnres
		}
		d = &Def{Type: sym.Type, Lit: sym.Lit}
		if path != "" {
			if d.Lit != nil {
				l, err := lit.Select(d.Lit, path)
				if err != nil {
					p.Unres = append(p.Unres, a)
					return err
				}
				d.Type = l.Typ()
				d.Lit = l
			} else {
				l, err := typ.Select(d.Type, path)
				if err != nil {
					p.Unres = append(p.Unres, a)
					return err
				}
				d.Type = l
			}
		}
	}
	if d == nil {
		p.Unres = append(p.Unres, a)
		return ErrUnres
	}
	s := d.Type
	if d.Lit != nil && s == typ.Typ {
		s = d.Lit.(typ.Type)
	}
	if t.Kind&typ.KindOpt != 0 {
		s = typ.Opt(s)
	}
	a.Lit, _ = replaceRef(at, s)
	return nil
}

func (p *Prog) reslDot(env Env, a *Sym) error {
	n := a.Name
	var d *Def
	if len(n) > 1 && n[1] == '?' {
		n = "." + n[2:]
		d = LookupSupports(env, n, '.')
	} else {
		env = Supports(env, '.')
		for env != nil && len(n) > 1 && n[1] == '.' {
			n = n[1:]
			env = Supports(env.Parent(), '.')
		}
		if env != nil {
			d = env.Get(n)
		}
	}
	if d == nil {
		p.Unres = append(p.Unres, a)
		return ErrUnres
	}
	a.Type = d.Type
	a.Lit = d.Lit
	return nil
}

func (p *Prog) reslAbs(env Env, a *Sym) error {
	x, n := a.Name[0], a.Name
	env = Supports(env, x)
	if env == nil {
		return cor.Errorf("no env found for path symbol %s", a.Name)
	}
	d := env.Get(n)
	if d == nil {
		p.Unres = append(p.Unres, a)
		return ErrUnres
	}
	a.Type = d.Type
	a.Lit = d.Lit
	return nil
}

func (p *Prog) dynCall(env Env, d *Dyn) (El, error) {
	if d == nil || len(d.Els) == 0 {
		return Ignore(d.Src)
	}
	fst, err := p.Resl(env, d.Els[0], typ.Void)
	if err != nil && err != ErrUnres {
		return d, err
	}
	t, l := ResInfo(fst)
	var sym string
	var cons bool
	switch t.Kind & typ.MaskElem {
	case typ.KindVoid:
		if fst.Typ() == typ.Typ {
			return Ignore(d.Src)
		}
		return d, ErrUnres
	case typ.KindTyp:
		lt, ok := l.(typ.Type)
		if !ok {
			return d, ErrUnres
		}
		if lt == typ.Void {
			return Ignore(d.Src)
		}
	case typ.KindFunc, typ.KindForm:
		ls, ok := l.(*Spec)
		if !ok {
			return d, ErrUnres
		}
		return p.NewCall(ls, d.Els[1:], d.Src)
	}
	if len(d.Els) == 1 && t.Kind&typ.KindAny != 0 {
		return fst, nil
	}
	sym, cons = p.Dyn(t)
	if sym == "" {
		return d, cor.Errorf("dyn unexpected first element %s %s", fst, fst.Typ())
	}
	args := d.Els
	if cons {
		args = args[1:]
	}
	call, err := p.BuiltinCall(env, sym, args, d.Src)
	if err != nil {
		return d, err
	}
	return call, nil
}

func replaceRef(t, el typ.Type) (typ.Type, bool) {
	switch k := t.Kind & typ.MaskRef; k {
	case typ.KindList, typ.KindDict:
		n, ok := replaceRef(t.Elem(), el)
		if ok {
			if k == typ.KindList {
				return typ.List(n), true
			}
			return typ.Dict(n), true
		}
	case typ.KindRef, typ.KindVar:
		return el, true
	}
	return t, false
}

type rfunc = func(*Prog, Env, El, typ.Type) (El, error)

func doAll(p *Prog, env Env, els []El, h typ.Type, f rfunc) (res []El, err error) {
	for i, el := range els {
		el, er := f(p, env, el, h)
		if er != nil {
			if err == nil || err == ErrUnres {
				err = er
			}
		}
		els[i] = el
	}
	return els, err
}
