package typ

import (
	"github.com/mb0/xelf/cor"
)

// Alt returns a new type alternative for a list of types. Other alternatives are flattened.
// If the first type is already an alterantive, the following types are added.
func Alt(alts ...Type) (res Type) {
	if len(alts) == 0 {
		return Void
	}
	fst := alts[0]
	if isAlt(fst) {
		return addAlts(fst, alts[1:])
	}
	res = Type{KindAlt, &Info{Params: make([]Param, 0, len(alts)*2)}}
	return addAlts(res, alts)
}

func AddAlts(t Type, alts []Type) (_ Type, err error) {
	if !isAlt(t) {
		return t, cor.Errorf("invalid type alternaive")
	}
	return addAlts(t, alts), nil
}

// Choose returns type t with all type alternatives reduced to its most specific representation.
func Choose(c *Ctx, t Type) Type { t, _ = choose(c, t); return t }
func choose(c *Ctx, t Type) (min Type, ok bool) {
	if t.Kind&MaskRef != KindAlt {
		if !t.HasParams() {
			return t, false
		}
		var ps []Param
		for i, p := range t.Params {
			p.Type, ok = choose(c, p.Type)
			if ok && ps == nil {
				ps = make([]Param, 0, len(t.Params))
				ps = append(ps, t.Params[:i]...)
			}
			ps = append(ps, p)
		}
		if ps == nil {
			return t, false
		}
		nfo := *t.Info
		nfo.Params = ps
		t.Info = &nfo
		return t, true
	}
	if !t.HasParams() {
		return Void, true
	}
	min = t.Params[0].Type
	max := Void
	for i, p := range t.Params[1:] {
		v, _ := choose(c, min)
		w, _ := choose(c, p.Type)
		tmp, err := Unify(c, v, w)
		if err != nil {
			t.Params = t.Params[i-1:]
			t.Params[0].Type = min
			return t, true
		}
		min = tmp
		if max == Void || Compare(max, p.Type) > LvlComp {
			max = p.Type
		} else {
			max = min
		}
	}
	if max == Void {
		return min, true
	}
	return max, true
}

func hasAlt(t, alt Type) bool {
	for _, p := range t.Params {
		if alt.Equal(p.Type) {
			return true
		}
	}
	return false
}

func addAlt(t, a Type) Type {
	if a.Kind != KindAlt {
		if !hasAlt(t, a) {
			t.Params = append(t.Params, Param{Type: a})
		}
	} else if a.ParamLen() > 0 {
		for _, p := range a.Params {
			t = addAlt(t, p.Type)
		}
	}
	return t
}

func addAlts(t Type, alts []Type) Type {
	for _, a := range alts {
		t = addAlt(t, a)
	}
	return t
}
