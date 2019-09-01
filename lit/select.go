package lit

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

type PathSeg = typ.PathSeg
type Path = typ.Path

var ReadPath = typ.ReadPath

// Select reads path and returns the selected literal from within the container l or an error.
func Select(l Lit, path string) (Lit, error) {
	p, err := ReadPath(path)
	if err != nil {
		return nil, err
	}
	return SelectPath(l, p)
}

// SelectPath returns the literal selected by path p from within l or an error.
func SelectPath(l Lit, p Path) (Lit, error) {
	return selectPath(l, p, false)
}
func selectPath(l Lit, p Path, subs bool) (_ Lit, err error) {
	for i, s := range p {
		if s.Sel && (i > 0 || !subs) {
			sub := p[i:]
			var res []Lit
			switch v := Deopt(l).(type) {
			case Indexer:
				res = make([]Lit, 0, v.Len())
				err = v.IterIdx(func(_ int, el Lit) error {
					el, err = selectPath(el, sub, true)
					if err != nil {
						return err
					}
					res = append(res, el)
					return nil
				})
			case Keyer:
				res = make([]Lit, 0, v.Len())
				err = v.IterKey(func(_ string, el Lit) error {
					el, err = selectPath(el, sub, true)
					if err != nil {
						return err
					}
					res = append(res, el)
					return nil
				})
			case typ.Type:
				return typ.Idxr(typ.Any), nil
			default:
				err = cor.Errorf("want idxer or keyer got %s", l.Typ())
			}
			if err != nil {
				return nil, err
			}
			return &List{Data: res}, nil
		} else if s.Key != "" {
			l, err = SelectKey(l, s.Key)
		} else {
			l, err = SelectIdx(l, s.Idx)
		}
		if err != nil {
			return nil, cor.Errorf("select %s: %w", p, err)
		}
	}
	return l, nil
}

func SelectKey(l Lit, key string) (Lit, error) {
	switch v := Deopt(l).(type) {
	case typ.Type:
		return typ.SelectKey(v, key)
	case Keyer:
		return v.Key(key)
	}
	return nil, cor.Errorf("key segment expects keyer got %s", l)
}

func SelectIdx(l Lit, idx int) (Lit, error) {
	switch v := l.(type) {
	case typ.Type:
		return typ.SelectIdx(v, idx)
	case Indexer:
		if idx < 0 {
			idx = v.Len() + idx
		}
		return v.Idx(idx)
	}
	return nil, cor.Errorf("idx segment expects idxer got %s", l.Typ())
}

// SetPath sets literal l at path p to el or returns an error. If create is true it tries to
// create or resize intermediate containers so they can hold el.
func SetPath(l Lit, p Path, el Lit, create bool) (Lit, error) {
	if l == nil {
		return l, cor.Errorf("set path got nil literal")
	}
	if len(p) == 0 {
		if a, ok := l.(Proxy); ok {
			err := a.Assign(el)
			return a, err
		}
		return nil, cor.Errorf("set path got empty path")
	}
	return setPath(l, el, p, create)
}

func setPath(l Lit, el Lit, p Path, create bool) (Lit, error) {
	s := p[0]
	if s.Key != "" {
		return setKey(l, el, s.Key, p[1:], create)
	}
	return setIdx(l, el, s.Idx, p[1:], create)
}

func setKey(l Lit, el Lit, key string, rest Path, create bool) (Lit, error) {
	t := l.Typ()
	v, ok := l.(Keyer)
	if !ok && create && (t.Kind == typ.KindAny || t.Kind&typ.KindKeyr != 0) {
		v, ok = &Dict{}, true
	}
	if !ok {
		return l, cor.Errorf("key segment %q expects keyer got %s", key, l.Typ())
	}
	if len(rest) > 0 {
		sl, err := SelectKey(v, key)
		if err != nil {
			sl = Nil
		}
		sl, err = setPath(sl, el, rest, create)
		if err != nil {
			return l, err
		}
		el = sl
	}
	v, err := v.SetKey(key, el)
	if err != nil {
		return l, cor.Errorf("set key %q: %w", key, err)
	}
	return v, nil
}

func setIdx(l Lit, el Lit, idx int, rest Path, create bool) (Lit, error) {
	t := l.Typ()
	v, ok := l.(Indexer)
	if !ok && create && (t.Kind == typ.KindAny ||
		t.Kind&typ.KindIdxr != 0) {
		res := make([]Lit, idx+1)
		for i := range res {
			res[i] = Nil
		}
		v, ok = &List{Data: res}, true
	}
	if !ok {
		return nil, cor.Errorf("idx segment expects idxer got %s", l.Typ())
	}
	if len(rest) > 0 {
		sl, err := SelectIdx(l, idx)
		if err != nil {
			sl = Nil
		}
		sl, err = setPath(sl, el, rest, create)
		if err != nil {
			return l, err
		}
		el = sl
	}
	_, err := v.SetIdx(idx, el)
	if err != nil {
		return l, err
	}
	return v, nil
}
