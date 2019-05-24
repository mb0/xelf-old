package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// MakeRec return a new abstract record literal with the given type or an error.
func MakeRec(t typ.Type) (*Rec, error) {
	if t.Kind&typ.MaskElem != typ.KindRec || !t.HasParams() {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, len(t.Params))
	for _, f := range t.Params {
		list = append(list, Keyed{f.Key(), ZeroProxy(f.Type)})
	}
	return &Rec{t, Dict{List: list}}, nil
}

// RecFromKeyed creates a new abstract record literal from the given list of keyed literals.
func RecFromKeyed(list []Keyed) *Rec {
	fs := make([]typ.Param, 0, len(list))
	for _, d := range list {
		fs = append(fs, typ.Param{d.Key, d.Lit.Typ()})
	}
	return &Rec{typ.Rec(fs), Dict{List: list}}
}

type Rec struct {
	Type typ.Type
	Dict
}

func (a *Rec) Typ() typ.Type { return a.Type }
func (a *Rec) IsZero() bool {
	if a.Dict.IsZero() {
		return true
	}
	for _, k := range a.List {
		if !k.Lit.IsZero() {
			return false
		}
	}
	return true
}
func (a *Rec) Idx(i int) (Lit, error) {
	_, err := a.Type.ParamByIdx(i)
	if err != nil {
		return nil, err
	}
	return a.Dict.List[i].Lit, nil
}
func (a *Rec) Key(key string) (Lit, error) {
	_, _, err := a.Type.ParamByKey(key)
	if err != nil {
		return nil, err
	}
	return a.Dict.Key(key)
}
func (a *Rec) SetIdx(i int, el Lit) (Indexer, error) {
	f, err := a.Type.ParamByIdx(i)
	if err != nil {
		return a, err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return a, err
		}
	}
	a.Dict.List[i].Lit = el
	return a, nil
}
func (a *Rec) SetKey(key string, el Lit) (Keyer, error) {
	f, _, err := a.Type.ParamByKey(key)
	if err != nil {
		return a, err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return a, err
		}
	}
	res, err := a.Dict.SetKey(key, el)
	if err != nil {
		return a, err
	}
	a.Dict = *res.(*Dict)
	return a, nil
}
func (a *Rec) IterIdx(it func(int, Lit) error) error {
	for i, el := range a.Dict.List {
		if err := it(i, el.Lit); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (a *Rec) String() string               { return bfr.String(a) }
func (a *Rec) MarshalJSON() ([]byte, error) { return bfr.JSON(a) }
func (a *Rec) WriteBfr(b *bfr.Ctx) error {
	b.WriteByte('{')
	n := 0
	for i, f := range a.Type.Params {
		el, err := a.Idx(i)
		if err != nil {
			return err
		}
		if f.Opt() && el.IsZero() {
			continue
		}
		if n++; n > 1 {
			b.Sep()
		}
		b.RecordKey(f.Key())
		err = writeLit(b, el)
		if err != nil {
			return err
		}
	}
	return b.WriteByte('}')
}

func (a *Rec) New() Proxy       { return &Rec{Type: a.Elem} }
func (a *Rec) Ptr() interface{} { return a }
func (a *Rec) Assign(l Lit) error {
	c, err := Convert(l, a.Type, 0)
	if err != nil {
		return err
	}
	return a.Dict.Assign(c)
}
