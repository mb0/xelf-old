package typ

import "github.com/mb0/xelf/cor"

// Unify unifies types a and b or returns an error.
func Unify(c *Ctx, a, b Type) error {
	x, av := c.apply(a)
	y, bv := c.apply(b)
	if x.Kind&MaskRef == KindVar {
		if x.Kind != y.Kind {
			return c.Bind(x.Kind, y)
		}
		return nil
	}
	if y.Kind&MaskRef == KindVar {
		return c.Bind(y.Kind, x)
	}
	v, w := Choose(x), Choose(y)
	var nfo *Info
	if v.Kind&MaskRef != KindAlt && v.Info != nil && w.Kind&MaskRef != KindAlt && w.Info != nil {
		if len(v.Params) != len(w.Params) {
			return cor.Errorf("param len mismatch")
		}
		ps := make([]Param, 0, len(w.Params))
		for i, ap := range v.Params {
			r := c.New()
			c.Bind(r.Kind, ap.Type)
			err := Unify(c, r, w.Params[i].Type)
			if err != nil {
				return err
			}
			r, _ = c.apply(r)
			ps = append(ps, Param{Type: r})
		}
		nfo = &Info{Params: ps}
	}
	cmp := Compare(v, w)
	if cmp > LvlEqual {
		return nil
	}
	switch cmp {
	case CmpCompAny, CmpCompBase, CmpCompList, CmpCompDict:
		return bindAlts(c, av, bv, a, b, x, y, v, w, y)
	case CmpCheckList, CmpCheckDict, CmpCompSpec, CmpCheckAny:
		return bindAlts(c, av, bv, a, b, x, y, v, w, x)
	case CmpConvArr, CmpCheckArr:
		return bindAlts(c, av, bv, a, b, x, y, v, w, Type{KindArr, nfo})
	case CmpConvMap, CmpCheckMap:
		return bindAlts(c, av, bv, a, b, x, y, v, w, Type{KindMap, nfo})
	default:
		if prim := v.Kind & MaskBase; prim != 0 && w.Kind&prim != 0 {
			if elem := v.Kind & MaskElem; elem == w.Kind&MaskElem {
				prim = elem
			} else {
				prim = w.Kind & prim

			}
			return bindAlts(c, av, bv, a, b, x, y, v, w, Type{prim, nfo})
		}
	}
	return cor.Errorf("type mismatch for %s %s: %s != %s", a, b, v, w)
}

func bindAlts(c *Ctx, av, bv bool, a, b, x, y, v, w, s Type) (res error) {
	if av {
		res = bindAlt(c, a, x, w, s)
	}
	if bv {
		err := bindAlt(c, b, y, v, s)
		if res == nil {
			res = err
		}
	}
	return
}

func bindAlt(c *Ctx, a, x, w, s Type) (err error) {
	if x.Kind&MaskRef != KindAlt && s != Void {
		return c.Bind(a.Kind, Alt(s, x, w))
	}
	return c.Bind(a.Kind, Alt(x, w))
}
