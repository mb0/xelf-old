package exp

import (
	"fmt"
	"strings"

	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// Resolve creates a new non-executing resolution context and resolves x with with given env.
func Resolve(env Env, x El) (El, error) { return (&Ctx{Part: true}).Resolve(env, x) }

// Execute creates a new executing resolution context and evaluates x with with given env.
func Execute(env Env, x El) (El, error) { return (&Ctx{Exec: true}).Resolve(env, x) }

// ResolveAll tries to resolve each element in xs in place and returns the first error if any.
func (c *Ctx) ResolveAll(env Env, els []El) ([]El, error) {
	var res error
	xs := els
	if !c.Part {
		xs = make([]El, len(els))
	}
	for i, x := range els {
		r, err := c.Resolve(env, x)
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
func (c *Ctx) Resolve(env Env, x El) (_ El, err error) {
	if x == nil {
		return typ.Void, nil
	}
	var rslv Resolver
	switch v := x.(type) {
	case Type: // resolve type references
		v, err = c.resolveTypRef(env, v)
		if err == ErrUnres {
			break
		}
		return v, err
	case Lit: // already resolved
		return v, nil
	case *Ref:
		if c.Part {
			rslv = v.Lookup(env)
		} else {
			rslv = env.Get(v.Key())
		}
	case Dyn:
		if len(v) == 0 {
			return typ.Void, nil
		}
		rslv = env.Get("dyn")
		if rslv != nil {
			x = &Expr{Sym: Sym{Name: "dyn", Rslv: rslv}, Args: v}
		}
	case Tag:
		v.Args, err = c.ResolveAll(env, v.Args)
		return v, err
	case Decl:
		v.Args, err = c.ResolveAll(env, v.Args)
		return v, err
	case *Expr:
		rslv = v.Lookup(env)
	default:
		return x, fmt.Errorf("unexpected expression %T %v", x, x)
	}
	if rslv == nil {
		c.Unres = append(c.Unres, x)
		return x, ErrUnres
	}
	// resolvers add to unres list themselves
	return rslv.Resolve(c, env, x)
}

func (c *Ctx) resolveTypRef(env Env, t Type) (_ Type, err error) {
	k := t.Last().Kind
	if k&typ.FlagRef == 0 {
		return t, nil
	}
	if t.Info == nil || t.Info.Ref == "" {
		if k != typ.FlagRef {
			return t, fmt.Errorf("unnamed %s not allowed", k)
		}
		// TODO infer type
		return t, ErrUnres
	}
	var key, rest string
	key = t.Info.Key()
	// return already resolved quasi ref types, otherwise add schema prefix '~'
	switch k {
	case typ.KindFlag, typ.KindEnum:
		if len(t.Consts) > 0 {
			return t, nil
		}
		key = "~" + key
	case typ.KindRec:
		if len(t.Fields) > 0 {
			return t, nil
		}
		key = "~" + key
	default:
		if lex.IsLetter(rune(key[0])) {
			split := strings.SplitN(key, ".", 2)
			if len(split) > 1 {
				key, rest = split[0], split[1]
			}
		}
	}
	rslv := env.Get(key)
	if rslv == nil {
		return t, ErrUnres
	}
	el, err := rslv.Resolve(c, env, t)
	if err != nil {
		return t, err
	}
	l, err := elTypeOrLit(el)
	if err != nil {
		return t, err
	}
	if rest != "" {
		l, err = lit.Select(l, rest)
		if err != nil {
			return t, err
		}
	}
	et, err := elType(l)
	if err != nil {
		return t, err
	}
	return replaceRef(t, et)
}

func elTypeOrLit(el El) (Lit, error) {
	switch v := el.(type) {
	case Type:
		return v, nil
	case Lit:
		return v, nil
	case *Ref:
		if v.Type != typ.Void {
			return v.Type, nil
		}
	case *Expr:
		if v.Type != typ.Void {
			return v.Type, nil
		}
	}
	return nil, ErrUnres
}

func elType(el El) (Type, error) {
	switch et := el.(type) {
	case Type:
		return et, nil
	case Lit:
		return et.Typ(), nil
	case *Ref:
		if et.Type != typ.Void {
			return et.Type, nil
		}
	case *Expr:
		if et.Type != typ.Void {
			return et.Type, nil
		}
	}
	return typ.Void, ErrUnres
}
func replaceRef(t, el Type) (Type, error) {
	var mask, shift typ.Kind
	for shift = 0; ; shift += typ.SlotSize {
		k := t.Kind >> shift
		switch k & typ.MaskElem {
		case typ.KindArr, typ.KindMap:
			mask |= k << shift
			continue
		}
		el.Kind |= k & typ.FlagOpt
		el.Kind = (el.Kind << shift) | mask
		return el, nil
	}
}
