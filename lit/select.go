package lit

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var (
	ErrEmptySeg = cor.StrError("empty segment")
	ErrIdxSeg   = cor.StrError("idx segment expects idxer")
	ErrKeySeg   = cor.StrError("key segment expects keyer")
)

// Seg is one segment of a path it can either be a non-empty key or an index.
type Seg struct {
	Key string
	Idx int
}

func (s Seg) String() string {
	if s.Key != "" {
		return s.Key
	}
	return strconv.Itoa(s.Idx)
}

// Path consists of non-empty segments separated by dots '.'. Segements starting with a digit
// are idx segments that try to select into an idxer container literal, otherwise the segment
// represents a key used to select into a keyer container literal.
type Path []Seg

func (p Path) String() string {
	var b strings.Builder
	for i, s := range p {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(s.String())
	}
	return b.String()
}

// ReadPath reads and returns the dot seperated segments for the path or an error.
func ReadPath(path string) (Path, error) {
	if path == "" {
		return nil, nil
	}
	segs := strings.Split(path, ".")
	res := make(Path, 0, len(segs))
	for _, seg := range segs {
		if seg == "" {
			return nil, ErrEmptySeg
		}
		if c := seg[0]; c >= '0' && c <= '9' {
			i, err := strconv.Atoi(seg)
			if err != nil {
				return nil, err
			}
			res = append(res, Seg{Idx: i})
		} else {
			res = append(res, Seg{Key: seg})
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
func SelectPath(l Lit, p Path) (_ Lit, err error) {
	for _, s := range p {
		if s.Key != "" {
			l, err = getKey(l, s.Key)
		} else {
			l, err = getIdx(l, s.Idx)
		}
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func getKey(l Lit, key string) (Lit, error) {
	switch v := l.(type) {
	case typ.Type:
		if v.Kind == typ.KindAny {
			return typ.Any, nil
		}
		switch v.Kind & typ.MaskElem {
		case typ.BaseDict:
			return typ.Any, nil
		case typ.KindMap:
			return v.Next(), nil
		case typ.KindObj:
			f, _, err := v.ParamByKey(key)
			if err != nil {
				return nil, err
			}
			return f.Type, nil
		}
	case Keyer:
		return v.Key(key)
	}
	return nil, ErrKeySeg
}

func getIdx(l Lit, idx int) (Lit, error) {
	switch v := l.(type) {
	case typ.Type:
		if v.Kind == typ.KindAny {
			return typ.Any, nil
		}
		switch v.Kind & typ.MaskElem {
		case typ.BaseList:
			return typ.Any, nil
		case typ.KindArr:
			return v.Next(), nil
		case typ.KindObj:
			f, err := v.ParamByIdx(idx)
			if err != nil {
				return nil, err
			}
			return f.Type, nil
		}
		return nil, ErrIdxSeg
	case Idxer:
		return v.Idx(idx)
	}
	return nil, ErrIdxSeg
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
	if !ok && create && (t.Kind == typ.KindAny ||
		t.Kind&typ.BaseDict != 0) {
		v, ok = &Dict{}, true
	}
	if !ok {
		return l, ErrKeySeg
	}
	if len(rest) > 0 {
		sl, err := getKey(l, key)
		if err != nil {
			sl = Nil
		}
		sl, err = setPath(sl, el, rest, create)
		if err != nil {
			return l, err
		}
		el = sl
	}
	err := v.SetKey(key, el)
	if err != nil {
		return l, err
	}
	return v, nil
}

func setIdx(l Lit, el Lit, idx int, rest Path, create bool) (Lit, error) {
	t := l.Typ()
	v, ok := l.(Idxer)
	if !ok && create && (t.Kind == typ.KindAny ||
		t.Kind&typ.BaseList != 0) {
		res := make(List, idx+1)
		for i := range res {
			res[i] = Nil
		}
		v, ok = &res, true
	}
	if !ok {
		return l, ErrIdxSeg
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
	err := v.SetIdx(idx, el)
	if err != nil {
		return l, err
	}
	return v, nil
}
