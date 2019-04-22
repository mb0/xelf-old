package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// Keyr is a generic container implementing the keyer type.
type Keyr struct {
	List []Keyed
}

// Keyed is a key associated with a literal.
type Keyed struct {
	Key string
	Lit
}

func (*Keyr) Typ() typ.Type  { return typ.Keyer }
func (d *Keyr) IsZero() bool { return d == nil || len(d.List) == 0 }

func (d *Keyr) Len() int {
	if d == nil {
		return 0
	}
	return len(d.List)
}
func (d *Keyr) Keys() []string {
	if d == nil {
		return nil
	}
	res := make([]string, 0, len(d.List))
	for _, v := range d.List {
		res = append(res, v.Key)
	}
	return res
}
func (d *Keyr) Key(k string) (Lit, error) {
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
func (d *Keyr) SetKey(k string, el Lit) (Keyer, error) {
	if d == nil {
		d = &Keyr{}
	}
	for i, v := range d.List {
		if v.Key == k {
			if el != nil {
				d.List[i].Lit = el
			} else {
				d.List = append(d.List[:i], d.List[i+1:]...)
			}
			return d, nil
		}
	}
	d.List = append(d.List, Keyed{k, el})
	return d, nil
}

func (d *Keyr) IterKey(it func(string, Lit) error) error {
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

func (v *Keyr) String() string               { return bfr.String(v) }
func (v *Keyr) MarshalJSON() ([]byte, error) { return bfr.JSON(v) }
func (v *Keyr) WriteBfr(b *bfr.Ctx) error {
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

func (v *Keyr) Ptr() interface{} { return v }
func (v *Keyr) Assign(l Lit) error {
	if v == nil {
		return cor.Errorf("nil keyer")
	}
	switch lv := Deopt(l).(type) {
	case *Keyr:
		*v = *lv
	case Keyer:
		res := v.List[:0]
		err := lv.IterKey(func(k string, e Lit) error {
			res = append(res, Keyed{k, e})
			return nil
		})
		if err != nil {
			return err
		}
		v.List = res
	default:
		return cor.Errorf("%q %T not assignable to %q", l.Typ(), l, v.Typ())
	}
	return nil
}

func writeKey(b *bfr.Ctx, key string) (err error) {
	if !b.JSON && cor.IsName(key) {
		b.WriteString(key)
		return b.WriteByte(':')
	}
	if b.JSON {
		key, err = cor.Quote(key, '"')
	} else {
		key, err = cor.Quote(key, '\'')
	}
	if err != nil {
		return err
	}
	b.WriteString(key)
	return b.WriteByte(':')
}
