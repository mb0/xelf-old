package typ

import "github.com/mb0/xelf/cor"

func Common(a, b Type) (s, t Type, err error) {
	if isAlt(a) || isAlt(b) {
		a = NewAlt(a, b)
		return a, a, nil
	}
	x, y := a.Kind, b.Kind
	if x&SlotMask == KindVar {
		return b, b, nil
	}
	if y&SlotMask == KindVar {
		return a, a, nil
	}
	if x == KindVoid || y == KindVoid {
		return Void, Void, nil
	}
	if x == KindRef || y == KindRef {
		return Void, Void, cor.Errorf("cannot compare type reference")
	}
	if x == KindAny {
		return Any, b, nil
	}
	if y == KindAny {
		return Any, a, nil
	}
	if x == y {
		if x&KindCont != 0 {
			a, err = commonCont(a, b)
		}
		return a, a, err
	}
	if y == KindCont && x&KindCont != 0 || y&x == x || specialCommon(x, y) {
		if x&KindCont != 0 {
			a, err = commonCont(a, b)
		}
		if x&MaskBase == x {
			return a, b, nil
		}
		return a, a, err
	}
	if x == KindCont && y&KindCont != 0 || x&y == y || specialCommon(y, x) {
		if y&KindCont != 0 {
			b, err = commonCont(b, a)
		}
		if y&MaskBase == y {
			return b, a, nil
		}
		return b, b, err
	}
	if base := x & MaskBase; y&base == base {
		a := Type{Kind: base}
		return a, a, nil
	}
	if x&KindAny != 0 && y&KindAny != 0 {
		return Any, Any, nil
	}
	return Void, Void, cor.Errorf("no common type for %s and %s", a, b)
}

func commonCont(a, b Type) (_ Type, err error) {
	if !b.HasParams() || a.ParamLen() != b.ParamLen() {
		if a.HasParams() {
			n := *a.Info
			n.Params = nil
			a.Info = &n
		}
	} else {
		var ps []Param
		for i, p := range a.Params {
			p.Type, _, err = Common(p.Type, b.Params[i].Type)
			if err != nil {
				return Void, err
			}
			if cont := a.Kind & KindCont; cont != 0 && cont != KindCont &&
				p.Kind == KindAny && len(a.Params) == 1 {
				break
			}
			if ps == nil {
				ps = make([]Param, 0, len(a.Params))
			}
			ps = append(ps, p)
		}
		n := *a.Info
		n.Params = ps
		a.Info = &n
	}
	return a, nil
}

func specialCommon(d, s Kind) bool {
	d, s = d&MaskRef, s&MaskRef
	return s == KindTime && (d == KindInt || d == KindNum) ||
		s == KindSpan && (d == KindInt || d == KindChar)
}
