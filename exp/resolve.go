package exp

import (
	"fmt"

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
		if err != nil {
			if !c.Exec && res == ErrUnres {
				res = err
				continue
			}
			return nil, err
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
func (c *Ctx) Resolve(env Env, x El) (_ El, err error) {
	if x == nil {
		return nil, nil
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
		rslv = env.Get(v.Name)
	case Dyn:
		if len(v) == 0 {
			return lit.Nil, nil
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
	if t.Last().Kind&typ.MaskRef != typ.KindRef {
		return t, nil
	}
	if t.Info == nil || t.Info.Ref == "" {
		// TODO infer type
		return t, ErrUnres
	}
	rslv := env.Get(t.Info.Ref)
	if rslv == nil {
		return t, ErrUnres
	}
	el, err := rslv.Resolve(c, env, t)
	if err != nil {
		return t, err
	}
	return elType(el)
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
