package typ

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
)

// Type represents the full type details. It consists of a kind and additional information.
type Type struct {
	Kind Kind `json:"typ"`
	*Info
}

func (a Type) Sub() Type {
	a.Kind = a.Kind >> SlotSize
	return a
}
func (Type) Typ() Type { return Typ }

type Const = cor.Const

// Info represents the reference name and type parameters or constants.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Params []Param `json:"params,omitempty"`
	Consts []Const `json:"consts,omitempty"`
	key    string
}

// Key returns the lowercase ref key.
func (a *Info) Key() string {
	if a == nil {
		return ""
	}
	if a.Ref != "" && a.key == "" {
		a.key = strings.ToLower(a.Ref)
	}
	return a.key
}

var errFieldNotFound = cor.StrError("field not found")

// FieldByIdx returns a pointer to the field at idx or an error.
func (a *Info) FieldByIdx(idx int) (*Param, error) {
	if a != nil && idx >= 0 && idx < len(a.Params) {
		return &a.Params[idx], nil
	}
	return nil, errFieldNotFound
}

// FieldByKey returns a pointer to the field and its idex at key or an error.
func (a *Info) FieldByKey(key string) (*Param, int, error) {
	if a != nil {
		for i, f := range a.Params {
			if f.Key() == key {
				return &a.Params[i], i, nil
			}
		}
	}
	return nil, -1, errFieldNotFound
}

// Param represents an type parameter with a name and type.
type Param struct {
	Name string `json:"name,omitempty"`
	Type
	key string
}

// Opt returns true if the param is optional, indicated by its name ending in a question mark.
func (a Param) Opt() bool { n := a.Name; return n != "" && n[len(n)-1] == '?' }

// Key returns the lowercase param key.
func (a Param) Key() string {
	if n := a.Name; n != "" && a.key == "" {
		if a.Opt() {
			n = n[:len(n)-1]
		}
		a.key = strings.ToLower(n)
	}
	return a.key
}

func (a Type) IsZero() bool { return a.Kind == 0 && a.Info.IsZero() }
func (a *Info) IsZero() bool {
	return a == nil || a.Ref == "" && len(a.Params) == 0 && len(a.Consts) == 0
}

type infoPair = struct{ a, b *Info }

func (a Type) Equal(b Type) bool { return a.equal(b, nil) }
func (a Type) equal(b Type, hist []infoPair) bool {
	return a.Kind == b.Kind && a.Info.equal(b.Info, a.Kind&FlagRef != 0, hist)
}
func (a *Info) Equal(b *Info) bool { return a.equal(b, false, nil) }
func (a *Info) equal(b *Info, ref bool, hist []infoPair) bool {
	if a == b {
		return true
	}
	if a.IsZero() {
		return b.IsZero()
	}
	if b.IsZero() ||
		a.Ref != b.Ref && a.Key() != b.Key() {
		return false
	}
	if ref {
		return true
	}
	if len(a.Params) != len(b.Params) ||
		len(a.Consts) != len(b.Consts) {
		return false
	}
	for i, av := range a.Consts {
		if av != b.Consts[i] {
			return false
		}
	}
	p := infoPair{a, b}
	for _, h := range hist {
		if h == p {
			return true
		}
	}
	for i, af := range a.Params {
		if !af.equal(b.Params[i], append(hist, p)) {
			return false
		}
	}
	return true
}

func (a Param) Equal(b Param) bool { return a.equal(b, nil) }
func (a Param) equal(b Param, hist []infoPair) bool {
	return (a.Name == b.Name || a.Key() == b.Key()) && a.Type.equal(b.Type, hist)
}

func (a Type) String() string               { return bfr.String(a) }
func (a Type) MarshalJSON() ([]byte, error) { return bfr.JSON(a) }

func (a Type) WriteBfr(b bfr.Ctx) error {
	if b.JSON {
		b.WriteByte('{')
		err := a.writeBfr(b, nil)
		b.WriteByte('}')
		return err
	}
	return a.writeBfr(b, nil)
}

func (a Type) writeBfr(b bfr.Ctx, hist []*Info) error {
	last := a.Last()
	switch last.Kind & MaskRef {
	case KindObj, KindRec:
		for i := 0; i < len(hist); i++ {
			h := hist[len(hist)-1-i]
			if a.Info.Equal(h) {
				writeRef(b, strconv.Itoa(i), a.Kind)
				return nil
			}
		}
	}
	var detail bool
	switch last.Kind & MaskRef {
	case KindRef:
		ref := ""
		if a.Info != nil {
			ref = a.Ref
		}
		writeRef(b, ref, a.Kind)
		return nil
	case KindObj, KindExp:
		detail = true
		fallthrough
	case KindFlag, KindEnum, KindRec:
		if b.JSON {
			b.WriteString(`"typ":"`)
		} else {
			b.WriteByte('(')
		}
		err := a.Kind.WriteBfr(b)
		if err != nil {
			return err
		}
		if b.JSON {
			b.WriteByte('"')
			err = a.Info.writeJSON(b, detail, append(hist, a.Info))
		} else {
			err = a.Info.writeXelf(b, detail, append(hist, a.Info))
			b.WriteByte(')')
		}
		return err
	}
	if b.JSON {
		b.WriteString(`"typ":"`)
		err := a.Kind.WriteBfr(b)
		b.WriteByte('"')
		return err
	}
	return a.Kind.WriteBfr(b)
}

func writeArrAndMap(b bfr.Ctx, k Kind) Kind {
	for {
		switch k & MaskElem {
		case KindArr:
			b.WriteString("arr|")
		case KindMap:
			b.WriteString("map|")
		default:
			return k
		}
		k = k >> SlotSize
	}
}

func writeRef(b bfr.Ctx, ref string, k Kind) {
	if b.JSON {
		b.WriteString(`"typ":"`)
		k = writeArrAndMap(b, k)
		b.WriteString("ref")
		if k&FlagOpt != 0 {
			b.WriteByte('?')
		}
		if ref != "" {
			b.WriteString(`","ref":"`)
			b.WriteString(ref)
		}
		b.WriteByte('"')
	} else {
		k = writeArrAndMap(b, k)
		b.WriteByte('@')
		b.WriteString(ref)
		if k&FlagOpt != 0 {
			b.WriteByte('?')
		}
	}
}

func (a *Info) writeXelf(b bfr.Ctx, detail bool, hist []*Info) error {
	if a == nil {
		return nil
	}
	if a.Ref != "" {
		b.WriteByte(' ')
		b.Quote(a.Ref)
	}
	if !detail {
		return nil
	}
	for i := 0; i < len(a.Params); i++ {
		f := a.Params[i]
		b.WriteString(" +")
		b.WriteString(f.Name)
		for _, o := range a.Params[i+1:] {
			if !f.Type.Equal(o.Type) {
				break
			}
			b.WriteString(" +")
			b.WriteString(o.Name)
			i++
		}
		b.WriteByte(' ')
		err := f.Type.writeBfr(b, hist)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Info) writeJSON(b bfr.Ctx, detail bool, hist []*Info) error {
	if a == nil {
		return nil
	}
	if a.Ref != "" {
		b.WriteString(`,"ref":`)
		b.Quote(a.Ref)
	}
	if !detail || len(a.Params) == 0 {
		return nil
	}
	b.WriteString(`,"params":[`)
	for i := 0; i < len(a.Params); i++ {
		f := a.Params[i]
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{`)
		if f.Name != "" {
			b.WriteString(`"name":`)
			b.Quote(f.Name)
			b.WriteByte(',')
		}
		err := f.Type.writeBfr(b, hist)
		if err != nil {
			return err
		}
		b.WriteByte('}')
	}
	b.WriteByte(']')
	return nil
}
