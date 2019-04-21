package exp

import (
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// Resolve creates a new non-executing resolution context and resolves x with with given env.
func Resolve(env Env, x El) (El, error) {
	return (&Ctx{Part: true}).Resolve(env, x, typ.Void)
}

// Execute creates a new executing resolution context and evaluates x with with given env.
func Execute(env Env, x El) (El, error) {
	return (&Ctx{Exec: true}).Resolve(env, x, typ.Void)
}

// ResolveAll tries to resolve each element in xs in place and returns the first error if any.
func (c *Ctx) ResolveAll(env Env, els []El, hint Type) ([]El, error) {
	var res error
	xs := els
	if !c.Part {
		xs = make([]El, len(els))
	}
	for i, x := range els {
		r, err := c.Resolve(env, x, hint)
		xs[i] = r
		if err != nil {
			if !c.Exec && err == ErrUnres {
				res = err
				continue
			}
			return nil, err
		}
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
func (c *Ctx) Resolve(env Env, x El, hint Type) (res El, err error) {
	if x == nil {
		return typ.Void, nil
	}
	if a, ok := x.(*Atom); ok {
		x = a.Lit
	}
	switch v := x.(type) {
	case Type: // resolve type references
		last := v.Last()
		v, err = c.resolveType(env, v, last)
		if err == ErrUnres {
			c.Unres = append(c.Unres, x)
			return x, err
		}
		return v, err
	case *Sym:
		return c.resolveSym(env, v, hint)
	case *Named:
		if v.El != nil {
			el, err := c.Resolve(env, v.El, typ.Void)
			if err != nil {
				return nil, err
			}
			v.El = el
		}
		return v, nil
	case *Dyn:
		return c.resolveDyn(env, v, hint)
	case *Call:
		return v.Spec.ResolveCall(c, env, v, hint)
	case Lit:
		return x, nil
	}
	return x, cor.Errorf("unexpected expression %T %v", x, x)
}

func (c *Ctx) resolveDyn(env Env, d *Dyn, hint Type) (El, error) {
	if c.Dyn != nil {
		return c.Dyn.ResolveDyn(c, env, d, hint)
	}
	return defaultDyn(c, env, d, hint)
}

func (c *Ctx) resolveSym(env Env, ref *Sym, hint Type) (El, error) {
	r, name, path, err := findResolver(env, ref.Name)
	if r == nil || err == ErrUnres {
		c.Unres = append(c.Unres, ref)
		return ref, ErrUnres
	}
	if err != nil {
		return nil, err
	}
	tmp := ref
	if ref.Name != name {
		tmp = &Sym{Name: name}
	}
	el, err := r.Resolve(c, env, tmp, typ.Void)
	if err != nil {
		if err == ErrUnres {
			c.Unres = append(c.Unres, ref)
		}
		return ref, err
	}
	res := el.(Lit)
	if path == "" {
		return res, nil
	}
	return lit.Select(res, path)
}

func (c *Ctx) resolveType(env Env, t Type, last Type) (_ Type, err error) {
	if last.Kind&typ.FlagRef == 0 {
		if last.Kind == typ.ExpFunc {
			// TODO resolve func signatures
		}
		return t, nil
	}
	k := last.Kind
	if last.Info == nil || last.Ref == "" {
		// TODO infer type
		return t, ErrUnres
	}
	key := last.Ref
	switch k {
	case typ.KindFlag, typ.KindEnum, typ.KindRec:
		// return already resolved schema types, otherwise add schema prefix '~'
		if len(last.Params) > 0 || len(last.Consts) > 0 {
			return t, nil
		}
		key = "~" + key
	}
	res, err := c.resolveSym(env, &Sym{Name: key}, typ.Void)
	if err != nil {
		return t, err
	}
	et, err := elType(res)
	if err != nil {
		return t, err
	}
	t, _ = replaceRef(t, et)
	return t, nil
}

func findResolver(env Env, sym string) (r *Def, name, path string, err error) {
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

func elType(el El) (Type, error) {
	switch t := el.Typ(); t.Kind {
	case typ.KindTyp:
		return el.(Type), nil
	case typ.ExpSym:
		s := el.(*Sym)
		if s.Def != nil && s.Def.Type != typ.Void {
			return s.Def.Type, nil
		}
	case typ.ExpForm, typ.ExpFunc:
		x := el.(*Call)
		if t := x.Spec.Res(); t != typ.Void {
			return t, nil
		}
	default:
		return t, nil
	}
	return typ.Void, ErrUnres
}

func replaceRef(t, el Type) (Type, bool) {
	switch k := t.Kind & typ.MaskRef; k {
	case typ.KindArr, typ.KindMap:
		n, ok := replaceRef(t.Elem(), el)
		if ok {
			if k == typ.KindArr {
				return typ.Arr(n), true
			}
			return typ.Map(n), true
		}
	case typ.KindRef, typ.KindVar:
		return el, true
	}
	return t, false
}
