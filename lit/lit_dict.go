package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// Keyed is a key associated with a literal.
type Keyed struct {
	Key string
	Lit
}

// Dict is a generic container implementing the dict type.
type Dict struct {
	Elem typ.Type
	List []Keyed
}

// MakeDict returns a new abstract dict literal with the given type or an error.
func MakeDict(t typ.Type) (*Dict, error) {
	return MakeDictCap(t, 0)
}

// MakeDictCap returns a new abstract dict literal with the given type and cap or an error.
func MakeDictCap(t typ.Type, cap int) (*Dict, error) {
	if t.Kind&typ.MaskElem != typ.KindDict {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, cap)
	return &Dict{t.Elem(), list}, nil
}

func (d *Dict) Typ() typ.Type     { return typ.Dict(d.Elem) }
func (d *Dict) Element() typ.Type { return d.Elem }
func (d *Dict) IsZero() bool      { return d == nil || len(d.List) == 0 }

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
	if d.Elem != typ.Void {
		return Null(d.Elem), nil
	}
	return Nil, nil
}
func (d *Dict) SetKey(k string, el Lit) (_ Keyer, err error) {
	if d == nil {
		return &Dict{List: []Keyed{{k, el}}}, nil
	}
	if el == nil {
		el = Nil
	}
	if d.Elem != typ.Void && d.Elem != typ.Any {
		el, err = Convert(el, d.Elem, 0)
		if err != nil {
			return d, err
		}
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

func (d *Dict) String() string               { return bfr.String(d) }
func (d *Dict) MarshalJSON() ([]byte, error) { return bfr.JSON(d) }
func (d *Dict) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('{')
	for i, e := range d.List {
		if i > 0 {
			b.Sep()
		}
		b.RecordKey(e.Key)
		writeLit(b, e.Lit)
	}
	return b.WriteByte('}')
}

func (d *Dict) New() Proxy       { return &Dict{Elem: d.Elem} }
func (d *Dict) Ptr() interface{} { return d }
func (d *Dict) Assign(l Lit) error {
	if d == nil {
		return cor.Errorf("nil keyer")
	}
	switch ld := Deopt(l).(type) {
	case *Dict:
		*d = *ld
	case Keyer:
		res := d.List[:0]
		err := ld.IterKey(func(k string, e Lit) error {
			res = append(res, Keyed{k, e})
			return nil
		})
		if err != nil {
			return err
		}
		d.List = res
	default:
		return cor.Errorf("%q %T not assignable to %q", l.Typ(), l, d.Typ())
	}
	return nil
}
