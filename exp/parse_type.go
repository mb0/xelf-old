package exp

import (
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func parseType(el El, hist []Type) (Type, error) {
	switch v := el.(type) {
	case Type:
		return v, nil
	case *Sym:
		return typ.ParseSym(v.Name, hist)
	case Dyn:
		return parseTypeDyn(v, hist)
	}
	return typ.Void, typ.ErrInvalid
}

func parseTypeDyn(args []El, hist []Type) (Type, error) {
	fst, args := args[0], args[1:]
	sym, ok := fst.(*Sym)
	if !ok {
		return typ.Void, typ.ErrInvalid
	}
	res, err := typ.ParseSym(sym.Name, hist)
	if err != nil {
		return typ.Void, err
	}
	needRef, needFields := typ.NeedsInfo(res)
	if !needRef && !needFields {
		if len(args) > 0 {
			return typ.Void, typ.ErrArgCount
		}
		return res, nil
	}
	res, err = parseTypeInfo(res, args, needRef, needFields, hist)
	if err != nil {
		return typ.Void, err
	}
	return res, nil
}

func parseTypeInfo(t Type, args []El, ref, param bool, hist []Type) (_ Type, err error) {
	if len(args) == 0 {
		return t, typ.ErrArgCount
	}
	t.Info = &typ.Info{}
	if ref {
		c, ok := args[0].(lit.Charer)
		if !ok {
			return t, typ.ErrRefName
		}
		t.Ref = c.Char()
		args = args[1:]
	}
	if param {
		dt, _ := t.Deopt()
		return parseTypeParams(t, args, append(hist, dt))
	}
	return t, nil
}

func parseTypeParams(t Type, args []El, hist []Type) (typ.Type, error) {
	if len(args) == 0 {
		return t, typ.ErrArgCount
	}
	group := t.Kind != typ.ExpForm
	decls, err := layoutParamArgs(args, group)
	if err != nil {
		return t, err
	}
	ps := make([]typ.Param, 0, len(decls))
	for _, d := range decls {
		p := typ.Param{Name: d.Name[1:]}
		if d.El != nil {
			p.Type, err = parseType(d.El, hist)
			if err != nil {
				return t, err
			}
		} else if group {
			return t, typ.ErrParamType
		}
		ps = append(ps, p)
	}
	t.Params = ps
	return t, nil
}

var layoutParams = []typ.Param{{Name: "unis"}}

func layoutParamArgs(args []El, group bool) ([]*Named, error) {
	lo, err := LayoutArgs(layoutParams, args)
	if err != nil {
		return nil, err
	}
	if group {
		return lo.Unis(0)
	}
	return lo.Decls(0)
}
