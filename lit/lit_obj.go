package lit

import (
	"github.com/mb0/xelf/typ"
)

// MakeObj return a new abstract obj literal with the given type or an error.
func MakeObj(t typ.Type) (Obj, error) {
	if t.Kind&typ.MaskElem != typ.KindObj || t.Info == nil || len(t.Fields) == 0 {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, len(t.Fields))
	for _, f := range t.Fields {
		list = append(list, Keyed{f.Key(), Null(f.Type)})
	}
	return &abstrObj{t, Dict{list}}, nil
}

type (
	abstrObj struct {
		typ typ.Type
		Dict
	}
)

func (a *abstrObj) Typ() typ.Type { return a.typ }
func (a *abstrObj) Idx(i int) (Lit, error) {
	_, err := idxField(a.typ, i)
	if err != nil {
		return nil, err
	}
	return a.Dict.List[i].Lit, nil
}
func (a *abstrObj) Key(key string) (Lit, error) {
	_, _, err := keyField(a.typ, key)
	if err != nil {
		return nil, err
	}
	return a.Dict.Key(key)
}
func (a *abstrObj) SetIdx(i int, el Lit) error {
	f, err := idxField(a.typ, i)
	if err != nil {
		return err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return err
		}
	}
	a.Dict.List[i].Lit = el
	return nil
}
func (a *abstrObj) SetKey(key string, el Lit) error {
	f, _, err := keyField(a.typ, key)
	if err != nil {
		return err
	}
	if el == nil {
		el = Null(f.Type)
	} else {
		el, err = Convert(el, f.Type, 0)
		if err != nil {
			return err
		}
	}
	return a.Dict.SetKey(key, el)
}
func (a *abstrObj) IterIdx(it func(int, Lit) error) error {
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


func idxField(t typ.Type, i int) (*typ.Field, error) {
	if n := t.Info; n != nil {
		if i >= 0 && i < len(n.Fields) {
			return &n.Fields[i], nil
		}
	}
	return nil, ErrIdxBounds
}
func keyField(t typ.Type, k string) (*typ.Field, int, error) {
	if n := t.Info; n != nil {
		for i, f := range n.Fields {
			if f.Key() == k {
				return &f, i, nil
			}
		}
	}
	return nil, 0, typ.ErrFieldName
}
