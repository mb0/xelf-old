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

// Select returns the literal from within the cotainer literal l at the given path or an error.
//
// A path consists of non-empty segments separated by dots '.'. Segements starting with a digit
// are idx segments that try to select into an idxer container literal, otherwise the segment
// represents a key used to select into a keyer container literal.
func Select(l Lit, path string) (_ Lit, err error) {
	if path == "" {
		return l, nil
	}
	part := strings.Split(path, ".")
	for _, k := range part {
		if k == "" {
			return nil, ErrEmptySeg
		}
		if c := k[0]; c >= '0' && c <= '9' {
			i, err := strconv.Atoi(k)
			if err != nil {
				return nil, err
			}
			v, ok := l.(Idxer)
			if !ok {
				return nil, ErrIdxSeg
			}
			l, err = v.Idx(i)
			if err != nil {
				return nil, err
			}
		} else {
			v, ok := l.(Keyer)
			if !ok {
				return nil, ErrKeySeg
			}
			l, err = v.Key(k)
			if err != nil {
				return nil, err
			}
		}
	}
	return l, nil
}
