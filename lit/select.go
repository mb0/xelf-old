package lit

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// PathSeg is one segment of a path it can either be a non-empty key or an index.
type PathSeg struct {
	Key string
	Idx int
	Sel bool
}

func (s PathSeg) String() string {
	if s.Key != "" {
		return s.Key
	}
	return strconv.Itoa(s.Idx)
}

// Path consists of non-empty segments separated by dots '.'. Segments starting with a digit or
// minus sign are idx segments that try to select into an idxer container literal, otherwise the
// segment represents a key used to select into a keyer container literal.
type Path []PathSeg

func (p Path) String() string {
	var b strings.Builder
	for i, s := range p {
		if i != 0 {
			if s.Sel {
				b.WriteByte('/')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteString(s.String())
	}
	return b.String()
}

func addSeg(p Path, s string, sel bool) (Path, error) {
	if s == "" {
		return nil, cor.Error("empty segment")
	}
	if c := s[0]; c == '-' || c >= '0' && c <= '9' {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		p = append(p, PathSeg{Idx: i, Sel: sel})
	} else {
		p = append(p, PathSeg{Key: s, Sel: sel})
	}
	return p, nil
}

// ReadPath reads and returns the dot separated segments for the path or an error.
func ReadPath(path string) (res Path, err error) {
	if path == "" {
		return nil, nil
	}
	res = make(Path, 0, len(path)>>2)
	var last int
	var sel bool
	for i, r := range path {
		switch r {
		case '.', '/':
			res, err = addSeg(res, path[last:i], sel)
			if err != nil {
				return nil, err
			}
			sel = r == '/'
			last = i + 1
		}
	}
	if len(path) > last {
		res, err = addSeg(res, path[last:], sel)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

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
			l, err = getKey(l, s.Key)
		} else {
			l, err = getIdx(l, s.Idx)
		}
		if err != nil {
			return nil, cor.Errorf("select %s: %w", p, err)
		}
	}
	return l, nil
}

func getKey(l Lit, key string) (Lit, error) {
	switch v := Deopt(l).(type) {
	case typ.Type:
		if key == "_" {
			return v.Elem(), nil
		}
		if v.Kind == typ.KindAny {
			return typ.Any, nil
		}
		switch v.Kind & typ.MaskElem {
		case typ.KindKeyr:
			return typ.Any, nil
		case typ.KindDict:
			return v.Elem(), nil
		case typ.KindRec:
			f, _, err := v.ParamByKey(key)
			if err != nil {
				return nil, err
			}
			return f.Type, nil
		}
	case Keyer:
		return v.Key(key)
	}
	return nil, cor.Errorf("key segment expects keyer got %s", l.Typ())
}

func getIdx(l Lit, idx int) (Lit, error) {
	switch v := l.(type) {
	case typ.Type:
		if v.Kind == typ.KindAny {
			return typ.Any, nil
		}
		switch v.Kind & typ.MaskElem {
		case typ.KindIdxr:
			return typ.Any, nil
		case typ.KindList:
			return v.Elem(), nil
		case typ.KindRec:
			if idx < 0 {
				idx = v.ParamLen() + idx
			}
			f, err := v.ParamByIdx(idx)
			if err != nil {
				return nil, err
			}
			return f.Type, nil
		}
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
		sl, err := getKey(v, key)
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
		sl, err := getIdx(l, idx)
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
