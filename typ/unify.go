package typ

import (
	"github.com/mb0/xelf/cor"
)

// Unify unifies types a and b or returns an error.
func Unify(c *Ctx, a, b Type) (Type, error) {
	x, av := c.apply(a, nil)
	y, bv := c.apply(b, nil)
	if isVar(x) {
		return unifyVar(c, x, y)
	}
	if isVar(y) {
		return unifyVar(c, y, x)
	}
	v, _ := choose(c, x)
	w, _ := choose(c, y)
	s, err := unify(c, v, w)
	if err != nil {
		return Void, err
	}
	var res error
	if av {
		res = bindAlt(c, a, x, w, s)
	}
	if bv {
		err := bindAlt(c, b, y, v, s)
		if res == nil {
			res = err
		}
	}
	return c.Apply(s), res
}

func isVar(t Type) bool { return t.Kind&MaskRef == KindVar }
func isAlt(t Type) bool { return t.Kind == KindAlt && t.HasParams() }

func unify(c *Ctx, a, b Type) (Type, error) {
	cmp := Compare(a, b)
	if cmp > LvlEqual {
		return a, nil
	}
	switch cmp {
	case CmpCompAny, CmpCompBase, CmpCompList, CmpCompDict:
		return b, nil
	case CmpCheckList, CmpCheckDict, CmpCompSpec, CmpCheckAny:
		return a, nil
	case CmpConvArr, CmpCheckArr:
		el, err := Unify(c, a.Elem(), b.Elem())
		if err != nil {
			return Void, err
		}
		return Type{KindList, &Info{Params: []Param{{Type: el}}}}, nil
	case CmpConvMap, CmpCheckMap:
		el, err := Unify(c, a.Elem(), b.Elem())
		if err != nil {
			return Void, err
		}
		return Type{KindDict, &Info{Params: []Param{{Type: el}}}}, nil
	default:
		if m := a.Kind & KindAny; m != 0 && b.Kind&m != 0 {
			res := Type{Kind: b.Kind & m}
			if elem := a.Kind & MaskElem; elem == b.Kind&MaskElem {
				res.Kind = elem
			}
			if a.HasParams() && b.HasParams() {
				if len(a.Params) != len(b.Params) {
					return res, nil
				}
				ps := make([]Param, 0, len(a.Params))
				for i, ap := range a.Params {
					c, err := Unify(c, ap.Type, b.Params[i].Type)
					if err != nil {
						return Void, err
					}
					ps = append(ps, Param{Type: c})
				}
				res.Info = &Info{Params: ps}
			}
			return res, nil
		}
	}
	return Void, cor.Errorf("cannot unify %s with %s", a, b)
}

func unifyVar(c *Ctx, v, t Type) (Type, error) {
	if v.HasParams() {
		if isVar(t) {
			t = mergeHint(t, v)
		} else {
			err := checkHint(c, t, v)
			if err != nil {
				return Void, err
			}
		}
	} else if isVar(t) && t.HasParams() {
		m := mergeHint(v, t)
		m.Kind = t.Kind
		return m, c.Bind(v.Kind, m)
	}
	if v.Kind == t.Kind {
		return v, nil
	}
	return t, c.Bind(v.Kind, t)
}

func mergeHint(v, o Type) Type {
	if !v.HasParams() {
		if v.Info == nil {
			v.Info = &Info{}
		}
		v.Params = append(v.Params, o.Params...)
		return v
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
