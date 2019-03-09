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
	case Tag:
		_, err = c.ResolveAll(env, v.Args, typ.Void)
		return x, err
	case Decl:
		_, err = c.ResolveAll(env, v.Args, typ.Void)
		return x, err
	case Dyn:
		return c.resolveDyn(env, v, hint)
	case *Expr:
		return v.Rslv.Resolve(c, env, v, hint)
	case Lit:
		return x, nil
	}
	return x, cor.Errorf("unexpected expression %T %v", x, x)
}

func (c *Ctx) resolveDyn(env Env, d Dyn, hint Type) (El, error) {
	if c.Dyn != nil {
		return c.Dyn.ResolveDyn(c, env, d, hint)
	}
	return defaultDyn(c, env, d, hint)
}

func (c *Ctx) resolveSym(env Env, ref *Sym, hint Type) (El, error) {
	sym := ref.Key()
	tref := sym != "" && sym[0] == '@'
	if tref {
		sym = sym[1:]
		if sym == "" {
			return typ.Infer, nil
		}
		t := typ.Ref(sym)
		if sym[len(sym)-1] == '?' {
			t = typ.Ref(sym[:len(sym)-1])
			return c.resolveType(env, typ.Opt(t), t)
		} else {
			t = typ.Ref(sym)
			return c.resolveType(env, t, t)
		}
	}
	if strings.HasPrefix(sym, "arr|") || strings.HasPrefix(sym, "map|") {
		t, err := typ.ParseSym(sym, nil)
		if err != nil {
			return nil, err
		}
		return c.resolveType(env, t, t.Last())
	}
	r, name, path, err := findResolver(env, sym)
	if r == nil || err == ErrUnres {
		c.Unres = append(c.Unres, ref)
		return ref, ErrUnres
	}
	if err != nil {
		return nil, err
	}
	tmp := ref
	if sym != name {
		tmp = &Sym{Name: name}
	}
	res, err := r.Resolve(c, env, tmp, typ.Void)
	if err != nil {
		if err == ErrUnres {
			c.Unres = append(c.Unres, ref)
		}
		return ref, err
	}
	if path == "" {
		return res, nil
	}
	return lit.Select(res.(Lit), path)
}

func (c *Ctx) resolveType(env Env, t Type, last Type) (_ Type, err error) {
	if last.Kind&typ.FlagRef == 0 {
		if last.Kind == typ.ExpFunc {
			// TODO resolve func signatures
		}
		return t, nil
	}
	k := last.Kind
	if t.Info == nil || t.Info.Ref == "" {
		if k != typ.FlagRef {
			return t, cor.Errorf("unnamed %s not allowed", k)
		}
		// TODO infer type
		return t, ErrUnres
	}
	key := t.Info.Key()
	switch k {
	case typ.KindFlag, typ.KindEnum, typ.KindRec:
		// return already resolved schema types, otherwise add schema prefix '~'
		if len(t.Params) > 0 || len(t.Consts) > 0 {
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
	return replaceRef(t, et)
}

func findResolver(env Env, sym string) (r Resolver, name, path string, err error) {
	if sym == "" {
		return nil, "", "", cor.Error("empty symbol")
	}
	// check prefixes
	var lookup bool
	switch x := sym[0]; x {
	case '~':
		return LookupSupports(env, sym, x), sym, "", nil
	case '$':
		tmp := sym[1:]
		if len(tmp) > 0 && tmp[0] == '$' {
			// program parameter use the program result prefix to select
			// the program environment.
			return GetSupports(env, tmp, '/'), sym, "", nil
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
		sym = "$" + tmp
		return env.Get(sym), sym, "", nil
	case '/':
		return GetSupports(env, sym, x), sym, "", nil
	case '.':
		sym = sym[1:]
		for len(sym) > 0 && sym[0] == '.' {
			sym = sym[1:]
			env = env.Parent()
			if env == nil {
				return nil, "", "", cor.Errorf("no env found for prefix %q", x)
			}
		}
	default:
		lookup = true
	}
	// check for path
	idx := strings.IndexByte(sym, '.')
	if idx > 0 {
		sym, path = sym[:idx], sym[idx+1:]
	}
	if lookup {
		r = Lookup(env, sym)
	} else {
		r = env.Get(sym)
	}
	return r, sym, path, nil
}

func elType(el El) (Type, error) {
	switch t := el.Typ(); t.Kind {
	case typ.KindTyp:
		return el.(Type), nil
	case typ.ExpSym:
		s := el.(*Sym)
		if t := s.Type; t != typ.Void {
			return t, nil
		}
	case typ.ExpForm, typ.ExpFunc:
		x := el.(*Expr)
		if t := x.Rslv.Res(); t != typ.Void {
			return t, nil
		}
	default:
		return t, nil
	}
	return typ.Void, ErrUnres
}

func replaceRef(t, el Type) (Type, error) {
	var mask, shift typ.Kind
	for shift = 0; ; shift += typ.SlotSize {
		k := t.Kind >> shift
		switch k & typ.MaskElem {
		case typ.KindArr, typ.KindMap:
			mask |= (k & typ.SlotMask) << shift
			continue
		}
		el.Kind |= k & typ.FlagOpt
		el.Kind = (el.Kind << shift) | mask
		return el, nil
	}
}
