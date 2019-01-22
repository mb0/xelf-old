package lit

import (
	"github.com/mb0/xelf/typ"
)

// MakeMap returns a new abstract map literal with the given type or an error.
func MakeMap(t typ.Type) (Map, error) {
	return MakeMapCap(t, 0)
}

// MakeMapCap returns a new abstract map literal with the given type and cap or an error.
func MakeMapCap(t typ.Type, cap int) (Map, error) {
	if t.Kind&typ.MaskElem != typ.KindMap {
		return nil, typ.ErrInvalid
	}
	list := make([]Keyed, 0, cap)
	return &abstractMap{t.Next(), Dict{list}}, nil
}

type abstractMap struct {
	elem typ.Type
	Dict
}

func (a *abstractMap) Typ() typ.Type  { return typ.Map(a.elem) }
func (a *abstractMap) Elem() typ.Type { return a.elem }
func (a *abstractMap) Key(k string) (Lit, error) {
	l, err := a.Dict.Key(k)
	if err != nil {
		return nil, err
	}
	if l == Nil {
		l = Null(a.elem)
	}
	return l, nil
}
func (a *abstractMap) SetKey(key string, el Lit) (err error) {
	if el != nil {
		el, err = Convert(el, a.elem, 0)
		if err != nil {
			return err
		}
	}
	return a.Dict.SetKey(key, el)
}
