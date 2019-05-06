package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var ErrIdxBounds = cor.StrError("idx out of bounds")

// List is a generic container implementing the indexer type.
type List struct {
	Elem typ.Type
	Data []Lit
}

// MakeList returns a new abstract list literal with the given type and len or an error.
func MakeList(t typ.Type, len int) (*List, error) {
	return MakeListCap(t, len, len)
}

// MakeListCap returns a new abstract list literal with the given type, len and cap or an error.
func MakeListCap(t typ.Type, len, cap int) (*List, error) {
	if t.Kind&typ.MaskElem != typ.KindList {
		return nil, typ.ErrInvalid
	}
	res := List{t.Elem(), make([]Lit, len, cap)}
	for i := range res.Data {
		res.Data[i] = Null(res.Elem)
	}
	return &res, nil
}

func (l *List) Typ() typ.Type     { return typ.List(l.Elem) }
func (l *List) IsZero() bool      { return l == nil || len(l.Data) == 0 }
func (l *List) Element() typ.Type { return l.Elem }
func (l *List) SetIdx(i int, el Lit) (_ Indexer, err error) {
	if l == nil || i < 0 || i >= len(l.Data) {
		return l, ErrIdxBounds
	}
	if el == nil {
		el = Nil
	}
	if l.Elem != typ.Void && l.Elem != typ.Any {
		el, err = Convert(el, l.Elem, 0)
		if err != nil {
			return l, err
		}
	}
	l.Data[i] = el
	return l, nil
}

func (l *List) Ptr() interface{} { return l }
func (l *List) Append(ls ...Lit) (_ Appender, err error) {
	for _, el := range ls {
		if el == nil {
			el = Nil
		}
		if l.Elem != typ.Void && l.Elem != typ.Any {
			el, err = Convert(el, l.Elem, 0)
			if err != nil {
				return l, err
			}
		}
		l.Data = append(l.Data, el)
	}
	return l, nil
}
func (l *List) Len() int { return len(l.Data) }

func (l *List) Idx(i int) (Lit, error) {
	if i < 0 || i >= len(l.Data) {
		return nil, ErrIdxBounds
	}
	return l.Data[i], nil
}
func (l *List) IterIdx(it func(int, Lit) error) error {
	for i, el := range l.Data {
		if err := it(i, el); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}

func (l *List) String() string               { return bfr.String(l) }
func (l *List) MarshalJSON() ([]byte, error) { return bfr.JSON(l) }
func (l *List) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('[')
	for i, e := range l.Data {
		if i > 0 {
			writeSep(b)
		}
		writeLit(b, e)
	}
	return b.WriteByte(']')
}

func (l *List) Assign(val Lit) error {
	switch v := Deopt(val).(type) {
	case *List:
		*l = *v
	case Indexer:
		res := l.Data[:0]
		err := v.IterIdx(func(i int, e Lit) error {
			res = append(res, e)
			return nil
		})
		if err != nil {
			return err
		}
		l.Data = res
	default:
		return cor.Errorf("%q not assignable to %q", val.Typ(), l.Typ())
	}
	return nil
}

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
