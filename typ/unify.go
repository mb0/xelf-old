package typ

import (
	"github.com/mb0/xelf/cor"
)

// Unify returns a unified type for a and b or an error.
func Unify(c *Ctx, a, b Type) (Type, error) {
	x, av := c.apply(a, nil)
	y, bv := c.apply(b, nil)
	if isVar(x) {
		return unifyVar(c, x, y)
	}
	if isVar(y) {
		return unifyVar(c, y, x)
	}
	if isAlt(x) && isAlt(y) {
		return Choose(Alt(x, y))
	}
	_, t, err := unify(c, x, y)
	if err != nil {
		return Void, err
	}
	var res error
	if av {
		res = bindAlt(c, a, Void, t, Void)
	}
	if bv {
		err := bindAlt(c, b, Void, t, Void)
		if res == nil {
			res = err
		}
	}
	return c.Apply(t), res
}

func isVar(t Type) bool { return t.Kind&MaskRef == KindVar }
func isAlt(t Type) bool { return t.Kind == KindAlt && t.HasParams() }

func unify(c *Ctx, a, b Type) (Type, Type, error) {
	if a == Any || b == Any {
		return Any, Any, nil
	}
	s, t, err := Common(a, b)
	if err != nil {
		return s, t, err
	}
	x, err := Choose(s)
	if err != nil {
		return s, t, err
	}
	switch x.Kind {
	case KindAny:
		return Void, Void, cor.Errorf("cannot unify %s with %s", a, b)
	case KindList, KindDict:
		_, err := Unify(c, a.Elem(), b.Elem())
		if err != nil {
			return Void, Void, err
		}
	default:
		if a.HasParams() && b.HasParams() {
			if len(a.Params) != len(b.Params) {
				return s, t, nil
			}
			for i, ap := range a.Params {
				_, err := Unify(c, ap.Type, b.Params[i].Type)
				if err != nil {
					return Void, Void, err
				}
			}
		}
	}
	return s, t, nil
}

func unifyVar(c *Ctx, v, t Type) (Type, error) {
	if isVar(t) {
		if t.HasParams() && !v.HasParams() {
			m := mergeHint(v, t)
			m.Kind = t.Kind
			t = m
		} else {
			t = mergeHint(t, v)
		}
	} else if v.HasParams() {
		err := checkHint(c, t, v)
		if err != nil {
			return Void, err
		}
	}
	if v.Kind != t.Kind {
		return t, c.Bind(v.Kind, t)
	}
	return t, nil
}

func mergeHint(v, o Type) Type {
	vp, op := v.HasParams(), o.HasParams()
	if !vp && !op {
		return v
	}
	if !vp {
		if v.Info == nil {
			v.Info = &Info{}
		}
		v.Params = append(v.Params, o.Params...)
	}
	if !op && o.Info != nil {
		o.Params = append(o.Params, v.Params...)
	}
	// TODO merge
	return v
}

func checkHint(c *Ctx, t, v Type) (res error) {
	if !v.HasParams() {
		return nil
	}
	for _, p := range v.Params {
		_, err := Unify(c, t, p.Type)
		if err == nil {
			return nil
		}
	}
	return cor.Errorf("cannot unify %s with constraint var %s", t, v)
}

func bindAlt(c *Ctx, a, x, w, s Type) error {
	if isVar(a) {
		t := w
		if x.Kind&MaskRef != KindAlt && s != Void {
			t = Alt(s, x, w)
		} else if x != Void {
			t = Alt(x, w)
		}
		n := a
		for isVar(n) {
			b, ok := c.binds.Get(n.Kind)
			if !ok || isVar(b) || !isVar(t) {
				err := c.Bind(n.Kind, t)
				if err != nil {
					return err
				}
			}
			n = b
		}
		return nil
	}
	if !a.HasParams() || !w.HasParams() || len(a.Params) != len(w.Params) {
		return nil // TODO cor.Errorf("params dont match for %s %s", a, w)
	}
	for i, p := range a.Params {
		o := w.Params[i].Type
		if isVar(o) {
			if b, ok := c.binds.Get(o.Kind); ok {
				o = b
			}
		}
		err := bindAlt(c, p.Type, Void, o, Void)
		if err != nil {
			continue // TODO return err
		}
	}
	return nil
}
