package typ

// Alt returns a new type alternative for a base type and a list of alternative types.
// If base is already alt type the list is added to its alternatives.
func Alt(base Type, alts ...Type) Type {
	if base.Kind&MaskRef != KindAlt {
		ps := append(make([]Param, 0, 1+len(alts)), Param{Type: base})
		base = Type{KindAlt, &Info{Params: ps}}
	}
	for _, p := range alts {
		if !containsAlt(base, p) {
			base.Params = append(base.Params, Param{Type: p})
		}
	}
	return base
}

// Choose returns type t with all type alternatives reduced to its most specific representation.
func Choose(t Type) Type {
	if t.Kind&MaskRef != KindAlt {
		if t.Info == nil {
			return t
		}
		ps := make([]Param, 0, len(t.Params))
		for _, p := range t.Params {
			p.Type = Choose(p.Type)
			ps = append(ps, p)
		}
		nfo := *t.Info
		nfo.Params = ps
		t.Info = &nfo
		return t
	}
	if t.Info == nil || len(t.Params) == 0 {
		return t
	}
	fst := t.Params[0].Type
	alts := make([]Type, 0, len(t.Params)-1)
	for _, p := range t.Params[1:] {
		pt := Choose(p.Type)
		if Compare(pt, fst) < LvlComp {
			continue
		}
		alts = append(alts, pt)
	}
	switch len(alts) {
	case 1:
		return alts[0]
	}
	return Choose(fst)
}

func containsAlt(t, alt Type) bool {
	if t.Kind&MaskRef == KindAlt {
		for _, p := range t.Params {
			if alt.Equal(p.Type) {
				return true
			}
		}
	}
	return false
}
