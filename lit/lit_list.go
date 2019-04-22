package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var ErrIdxBounds = cor.StrError("idx out of bounds")

// Idxr is a generic container implementing the indexer type.
type Idxr []Lit

func (Idxr) Typ() typ.Type  { return typ.Idxer }
func (l Idxr) IsZero() bool { return len(l) == 0 }

func (l Idxr) Len() int { return len(l) }
func (l Idxr) Idx(i int) (Lit, error) {
	if i < 0 || i >= len(l) {
		return nil, ErrIdxBounds
	}
	return l[i], nil
}
func (l Idxr) SetIdx(i int, el Lit) (Indexer, error) {
	if i < 0 || i >= len(l) {
		return l, ErrIdxBounds
	}
	l[i] = el
	return l, nil
}
func (l Idxr) IterIdx(it func(int, Lit) error) error {
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

func (l Idxr) String() string               { return bfr.String(l) }
func (l Idxr) MarshalJSON() ([]byte, error) { return bfr.JSON(l) }
func (l Idxr) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('[')
	for i, e := range l {
		if i > 0 {
			writeSep(b)
		}
		writeLit(b, e)
	}
	return b.WriteByte(']')
}

func (v *Idxr) Ptr() interface{} { return v }
func (l *Idxr) Assign(val Lit) error {
	switch v := Deopt(val).(type) {
	case *Idxr:
		*l = *v
	case Idxr:
		*l = v
	case Indexer:
		res := (*l)[:0]
		err := v.IterIdx(func(i int, e Lit) error {
			res = append(res, e)
			return nil
		})
		if err != nil {
			return err
		}
		*l = res
	default:
		return cor.Errorf("%q not assignable to %q", val.Typ(), l.Typ())
	}
	return nil
}

func (l Idxr) Append(vals ...Lit) (Appender, error) {
	return append(l, vals...), nil
}
func (l Idxr) Element() typ.Type { return typ.Any }

func writeSep(b *bfr.Ctx) error {
	if b.JSON {
		return b.WriteByte(',')
	}
	return b.WriteByte(' ')
}

func writeLit(b *bfr.Ctx, e Lit) error {
	if e == nil {
		return b.Fmt("null")
	}
	return e.WriteBfr(b)
}
