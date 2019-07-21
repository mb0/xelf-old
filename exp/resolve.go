package exp

import (
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// Resolve creates a new non-executing resolution context and resolves x with with given env.
func Resolve(env Env, x El) (El, error) {
	return NewCtx(true, false).Resolve(env, x, typ.Void)
}

// Execute creates a new executing resolution context and evaluates x with with given env.
func Execute(env Env, x El) (El, error) {
	return NewCtx(false, true).Resolve(env, x, typ.Void)
}

// ResolveAll tries to resolve each element in xs in place and returns the first error if any.
func (c *Ctx) ResolveAll(env Env, els []El, hint typ.Type) ([]El, error) {
	var res error
	xs := els
	if !c.Part {
		xs = make([]El, len(els))
	}
	for i, x := range els {
		r, err := c.Resolve(env, x, hint)
		if err != nil {
			if err != ErrUnres {
				return nil, err
			}
			res = err
		}
		xs[i] = r
	}
	return xs, res
}

// Resolve resolves x within env and returns the result or an error.
//
// This method will not resolve any element itself but instead tries to look up an applicable
// resolver in the environment. If it cannot find a resolver it will add the element to the
// context's unresolved slice.
// The resolver implementations usually use this method either directly or indirectly to resolve
// arguments, which are then again added to the unresolved elements when appropriate.
func (c *Ctx) Resolve(env Env, x El, hint typ.Type) (res El, err error) {
	if x == nil {
		return &Atom{Lit: typ.Void}, nil
	}
	switch v := x.(type) {
	case *Atom:
		if v.Typ() == typ.Typ { // resolve type references
			t := v.Lit.(typ.Type)
			if !t.Resolved() {
				t, err = c.resolveType(env, t)
				if err != nil {
					if err == ErrUnres {
						c.Unres = append(c.Unres, v)
						return x, err
					}
					return nil, err
				}
				v.Lit = t
			}
		}
		return c.checkHint(hint, v)
	case *Sym:
		return c.resolveSym(env, v, hint)
	case *Named:
		if v.El != nil {
			if d, ok := v.El.(*Dyn); ok {
				els, err := c.ResolveAll(env, d.Els, typ.Void)
				if err != nil {
					return nil, err
				}
				d.Els = els
			} else {
				el, err := c.Resolve(env, v.El, typ.Void)
				if err != nil {
					return nil, err
				}
				v.El = el
			}
		}
		return v, nil
	case *Dyn:
		return c.resolveDyn(env, v, hint)
	case *Call:
		if v.Type == typ.Void {
			v.Type = c.Inst(v.Spec.Type)
		} //else {
		//	log.Printf("resolve call %s type already instantiated %s", v, v.Type)
		//}
		return v.Spec.Resolve(c, env, v, hint)
	}
	return x, cor.Errorf("unexpected expression %T %v", x, x)
}

func (c *Ctx) checkHint(hint typ.Type, l El) (El, error) {
	if hint != typ.Void {
		if lt := l.Typ(); lt != typ.Void {
			_, err := typ.Unify(c.Ctx, lt, hint)
			if err != nil {
				return nil, err
			}
		}
	}
	return l, nil
}

func (c *Ctx) resolveDyn(env Env, d *Dyn, h typ.Type) (El, error) {
	if c.Dyn == nil {
		def := Lookup(env, "dyn")
		if def != nil {
			c.Dyn, _ = def.Lit.(*Spec)
		}
	}
	if c.Dyn == nil {
		return d, ErrUnres
	}
	return c.Dyn.Resolve(c, env, &Call{Spec: c.Dyn, Type: c.Inst(c.Dyn.Type), Args: d.Els}, h)
}

func (c *Ctx) resolveSym(env Env, ref *Sym, hint typ.Type) (El, error) {
	r, _, path, err := findDef(env, ref.Name)
	if r == nil || r.Lit == nil || err == ErrUnres {
		if r != nil && r.Type != typ.Void {
			res := r.Type
			if path != "" {
				l, err := lit.Select(res, path)
				if err != nil {
					return nil, err
				}
				res = l.(typ.Type)
			}
			if hint != typ.Void {
				res, err = typ.Unify(c.Ctx, res, hint)
				if err != nil {
					return nil, err
				}
			}
		}
		c.Unres = append(c.Unres, ref)
		return ref, ErrUnres
	}
	if err != nil {
		return nil, err
	}
	res := r.Lit
	if path != "" {
		res, err = lit.Select(res, path)
		if err != nil {
			return nil, err
		}
	}
	return c.checkHint(hint, &Atom{res, ref.Source()})
}

func (c *Ctx) resolveType(env Env, t typ.Type) (_ typ.Type, err error) {
	last := t.Last()
	if !last.HasRef() {
		return t, ErrUnres
	}
	key := last.Key()
	switch last.Kind {
	case typ.KindSch, typ.KindBits, typ.KindEnum, typ.KindObj:
		// return already resolved schema types, otherwise add schema prefix '~'
		if len(last.Params) > 0 || len(last.Consts) > 0 {
			return t, nil
		}
		key = "~" + key
	}
	def, _, path, err := findDef(env, key)
	if def == nil || err != nil {
		return t, ErrUnres
	}
	s := def.Type
	if def.Lit != nil && s == typ.Typ {
		s = def.Lit.(typ.Type)
	}
	if path != "" {
		res, err := lit.Select(s, path)
		if err != nil {
			return typ.Void, err
		}
		s = res.(typ.Type)
	}
	if last.Kind&typ.KindOpt != 0 {
		s = typ.Opt(s)
	}
	t, _ = replaceRef(t, s)
	return t, nil
}

func findDef(env Env, sym string) (r *Def, name, path string, err error) {
	if sym == "" {
		return nil, "", "", cor.Error("empty symbol")
	}
	// check prefixes
	var lookup bool
	switch x := sym[0]; x {
	case '~':
		return LookupSupports(env, sym, x), sym, "", nil
	case '$', '/':
		tmp := sym[1:]
		if len(tmp) > 0 && tmp[0] == x {
			env = lastSupports(env, x)
			return env.Get(tmp), sym, "", nil
		}
		env = Supports(env, x)
		if env == nil {
			return nil, "", "", cor.Errorf("no env found for prefix %q", x)
		}
		for len(tmp) > 0 && tmp[0] == '.' {
			tmp = tmp[1:]
			env = Supports(env.Parent(), x)
			if env == nil {
				return nil, "", "", cor.Errorf("no env found for prefix %q", x)
			}
		}
		sym = string(x) + tmp
		return env.Get(sym), sym, "", nil
	case '.':
		if len(sym) > 1 && sym[1] == '?' {
			tmp := "." + sym[2:]
			return LookupSupports(env, tmp, x), tmp, "", nil

		}
		env = Supports(env, x)
		if env == nil {
			return nil, "", "", cor.Errorf("no env found for prefix %q", x)
		}
		// always leave one dot as symbol prefix to indicate a relative path symbol
		for len(sym) > 1 && sym[1] == '.' {
			sym = sym[1:]
			env = Supports(env.Parent(), x)
			if env == nil {
				return nil, "", "", cor.Errorf("no env found for prefix %q", x)
			}
		}
	default:
		lookup = true
		// check for path
		idx := strings.IndexByte(sym, '.')
		if idx > 0 {
			sym, path = sym[:idx], sym[idx+1:]
		}
	}
	if lookup {
		r = Lookup(env, sym)
	} else {
		r = env.Get(sym)
	}
	return r, sym, path, nil
}

func lastSupports(env Env, x byte) (last Env) {
	e := Supports(env, x)
	for e != nil {
		last = e
		e = Supports(e.Parent(), x)
	}
	return last
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
