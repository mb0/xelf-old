package typ

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/cor"
)

// PathSeg is one segment of a path. It consists of a dot or slash, followed by a key or index.
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

// Path consists of non-empty segments separated by dots '.' or slashes '/'. Segments starting with
// a digit or minus sign are idx segments that try to select into an idxer literal,
// otherwise the segment represents a key used to select into a keyer literal.
// Segments starting with a slash signify a selection from a idxer literal.
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

// Select reads path and returns the selected type from t or an error.
func Select(t Type, path string) (Type, error) {
	p, err := ReadPath(path)
	if err != nil {
		return Void, err
	}
	return SelectPath(t, p)
}

// SelectPath returns the selected type from t or an error.
func SelectPath(t Type, p Path) (r Type, err error) {
	for _, s := range p {
		if s.Key != "" {
			r, err = SelectKey(t, s.Key)
		} else {
			r, err = SelectIdx(t, s.Idx)
		}
		if err != nil {
			return Void, err
		}
		t = r
	}
	return t, nil
}

func SelectKey(t Type, key string) (Type, error) {
	switch t.Kind & MaskElem {
	case KindAny, KindKeyr:
		return Any, nil
	case KindDict:
		return t.Elem(), nil
	case KindRec:
		f, _, err := t.ParamByKey(key)
		if err != nil {
			return Void, err
		}
		return f.Type, nil
	case KindIdxr, KindList:
		if key == "_" {
			return t.Elem(), nil
		}
	}
	return Void, cor.Errorf("key segment expects keyer type got %s", t)
}

func SelectIdx(t Type, idx int) (Type, error) {
	switch t.Kind & MaskElem {
	case KindAny, KindIdxr:
		return Any, nil
	case KindList:
		return t.Elem(), nil
	case KindRec:
		if idx < 0 {
			idx = t.ParamLen() + idx
		}
		f, err := t.ParamByIdx(idx)
		if err != nil {
			return Void, err
		}
		return f.Type, nil
	}
	return Void, cor.Errorf("idx segment expects idxer type got %s", t)
}
