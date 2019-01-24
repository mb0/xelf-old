package lit

import (
	"errors"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

var ErrNilKeyer = errors.New("nil keyer")

// Dict is a generic container implementing Keyer.
type Dict struct {
	List []Keyed
}

// Keyed is a key associated with a literal.
type Keyed struct {
	Key string
	Lit
}

func (*Dict) Typ() typ.Type  { return typ.Dict }
func (d *Dict) IsZero() bool { return d == nil || len(d.List) == 0 }

func (d *Dict) Len() int {
	if d == nil {
		return 0
	}
	return len(d.List)
}
func (d *Dict) Keys() []string {
	if d == nil {
		return nil
	}
	res := make([]string, 0, len(d.List))
	for _, v := range d.List {
		res = append(res, v.Key)
	}
	return res
}
func (d *Dict) Key(k string) (Lit, error) {
	if d == nil {
		return Nil, nil
	}
	for _, v := range d.List {
		if v.Key == k {
			return v.Lit, nil
		}
	}
	return Nil, nil
}
func (d *Dict) SetKey(k string, el Lit) error {
	if d == nil {
		return ErrNilKeyer
	}
	for i, v := range d.List {
		if v.Key == k {
			if el != nil {
				d.List[i].Lit = el
			} else {
				d.List = append(d.List[:i], d.List[i+1:]...)
			}
			return nil
		}
	}
	d.List = append(d.List, Keyed{k, el})
	return nil
}

func (d *Dict) IterKey(it func(string, Lit) error) error {
	if d == nil {
		return nil
	}
	for _, el := range d.List {
		if err := it(el.Key, el.Lit); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}

func (v *Dict) String() string               { return bfr.String(v) }
func (v *Dict) MarshalJSON() ([]byte, error) { return bfr.JSON(v) }
func (v *Dict) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('{')
	for i, e := range v.List {
		if i > 0 {
			writeSep(b)
		}
		writeKey(b, e.Key)
		writeLit(b, e.Lit)
	}
	return b.WriteByte('}')
}

func (v *Dict) Ptr() interface{} { return v }
func (v *Dict) Assign(l Lit) error {
	switch lv := l.(type) {
	case *Dict:
		*v = *lv
	case Keyer:
		res := v.List[:0]
		err := v.IterKey(func(k string, e Lit) error {
			res = append(res, Keyed{k, e})
			return nil
		})
		if err != nil {
			return err
		}
		v.List = res
	default:
		return ErrNotAssignable
	}
	return nil
}

func writeKey(b bfr.Ctx, key string) (err error) {
	if !b.JSON && lex.IsName(key) {
		b.WriteByte('+')
		b.WriteString(key)
		return b.WriteByte(' ')
	}
	if b.JSON {
		key, err = lex.Quote(key, '"')
	} else {
		key, err = lex.Quote(key, '\'')
	}
	if err != nil {
		return err
	}
	b.WriteString(key)
	return b.WriteByte(':')
}
