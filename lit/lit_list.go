package lit

import (
	"errors"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

var ErrIdxBounds = errors.New("idx out of bounds")

// List is a generic container implementing Idxer.
type List []Lit

func (List) Typ() typ.Type  { return typ.List }
func (l List) IsZero() bool { return len(l) == 0 }

func (l List) Len() int { return len(l) }
func (l List) Idx(i int) (Lit, error) {
	if i < 0 || i >= len(l) {
		return nil, ErrIdxBounds
	}
	return l[i], nil
}
func (l List) SetIdx(i int, el Lit) error {
	if i < 0 || i >= len(l) {
		return ErrIdxBounds
	}
	l[i] = el
	return nil
}
func (l List) IterIdx(it func(int, Lit) error) error {
	for i, el := range l {
		if err := it(i, el); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}

func (l List) String() string               { return bfr.String(l) }
func (l List) MarshalJSON() ([]byte, error) { return bfr.JSON(l) }
func (l List) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('[')
	for i, e := range l {
		if i > 0 {
			writeSep(b)
		}
		writeLit(b, e)
	}
	return b.WriteByte(']')
}

func writeSep(b bfr.Ctx) error {
	if b.JSON {
		return b.WriteByte(',')
	}
	return b.WriteByte(' ')
}

func writeLit(b bfr.Ctx, e Lit) error {
	if e == nil {
		_, err := b.WriteString("null")
		return err
	}
	return e.WriteBfr(b)
}
