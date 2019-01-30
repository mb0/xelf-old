package lit

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrEmptySeg = errors.New("empty segment")
	ErrIdxSeg   = errors.New("idx segment expects idxer")
	ErrKeySeg   = errors.New("key segment expects keyer")
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

// SelectPath returns the literal selected by path p from within the cotainer l or an error.
func SelectPath(l Lit, p Path) (_ Lit, err error) {
	for _, s := range p {
		if s.Key != "" {
			v, ok := l.(Keyer)
			if !ok {
				return nil, ErrKeySeg
			}
			l, err = v.Key(s.Key)
		} else {
			v, ok := l.(Idxer)
			if !ok {
				return nil, ErrIdxSeg
			}
			l, err = v.Idx(s.Idx)
		}
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}
