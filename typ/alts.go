package typ

import "sort"

func NewAlt(alts ...Type) (res Type) {
	res = Type{KindAlt, &Info{Params: make([]Param, 0, len(alts)*2)}}
	return addAlts(res, alts)
}

// Alt returns a new type alternative for a list of types. Other alternatives are flattened.
// If the first type is already an alterantive, the following types are added.
func Alt(alts ...Type) (res Type) {
	if len(alts) == 0 {
		return Void
	}
	if fst := alts[0]; isAlt(fst) {
		return addAlts(fst, alts[1:])
	}
	return NewAlt(alts...)
}

// Choose returns type t with all type alternatives reduced to its most specific representation.
func Choose(t Type) (_ Type, err error) {
	return choose(t, nil)
}
func choose(t Type, hist []*Info) (_ Type, err error) {
	if t.Kind&MaskRef != KindAlt {
		if !t.HasParams() {
			return t, nil
		}
		for i := 0; i < len(hist); i++ {
			h := hist[len(hist)-1-i]
			if t.Info == h {
				return t, nil
			}
		}
		hist = append(hist, t.Info)
		var ps []Param
		for i, p := range t.Params {
			p.Type, err = choose(p.Type, hist)
			if err != nil {
				return Void, err
			}
			if ps == nil {
				ps = make([]Param, 0, len(t.Params))
				ps = append(ps, t.Params[:i]...)
			}
			ps = append(ps, p)
		}
		if ps != nil {
			nfo := *t.Info
			nfo.Params = ps
			t.Info = &nfo
		}
		return t, nil
	}
	if !t.HasParams() {
		return Void, nil
	}
	var a, b, tmp Type
	for i, p := range t.Params {
		if i == 0 {
			a = p.Type
			continue
		}
		a, tmp, err = Common(a, p.Type)
		if err != nil {
			return Void, err
		}
		if b == Void || b.Kind > tmp.Kind {
			b = tmp
		} else {
			b = a
		}
	}
	if b != Void {
		return b, nil
	}
	return a, nil
}

func hasAlt(t, alt Type) bool {
	ps := t.Params
	i := sort.Search(len(ps), func(i int) bool {
		return ps[i].Kind >= alt.Kind
	})
	return i < len(ps) && ps[i].Type == alt
}

func addAlt(t, a Type) Type {
	if a.Kind != KindAlt {
		ps := t.Params
		i := sort.Search(len(ps), func(i int) bool {
			return ps[i].Kind >= a.Kind
		})
		if i >= len(ps) {
			ps = append(ps, Param{Type: a})
		} else if ps[i].Type != a {
			ps = append(ps[:i+1], ps[i:]...)
			ps[i] = Param{Type: a}
		}
		t.Params = ps
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
