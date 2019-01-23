package lit

import (
	"github.com/mb0/xelf/typ"
)

// MakeArr returns a new abstract arr literal with the given type and len or an error.
func MakeArr(t typ.Type, len int) (Arr, error) {
	return MakeArrCap(t, len, len)
}

// MakeArrCap returns a new abstract arr literal with the given type, len and cap or an error.
func MakeArrCap(t typ.Type, len, cap int) (Arr, error) {
	if t.Kind&typ.MaskElem != typ.KindArr {
		return nil, typ.ErrInvalid
	}
	res := abstrArr{t.Next(), make(List, len, cap)}
	for i := range res.List {
		res.List[i] = Null(res.elem)
	}
	return &res, nil
}

type (
	abstrArr struct {
		elem typ.Type
		List
	}
)

func (a abstrArr) Typ() typ.Type  { return typ.Arr(a.elem) }
func (a abstrArr) Elem() typ.Type { return a.elem }
func (a abstrArr) SetIdx(i int, el Lit) (err error) {
	if el == nil {
		el = Null(a.elem)
	} else {
		el, err = Convert(el, a.elem, 0)
		if err != nil {
			return err
		}
	}
	return a.List.SetIdx(i, el)
}
