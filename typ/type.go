package typ

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
)

// Type represents the full type details. It consists of a kind and additional information.
type Type struct {
	Kind Kind
	*Info
}

func (Type) Typ() Type { return Typ }

type Const = cor.Const

// Info represents the reference name and type parameters or constants.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Params []Param `json:"params,omitempty"`
	Consts []Const `json:"consts,omitempty"`
}

// Key returns the lowercase ref key.
func (a *Info) Key() string {
	if a != nil {
		return cor.Keyed(a.Ref)
	}
	return ""
}

func (a *Info) ParamLen() int {
	if a == nil {
		return 0
	}
	return len(a.Params)
}

// ParamByIdx returns a pointer to the field at idx or an error.
func (a *Info) ParamByIdx(idx int) (*Param, error) {
	if a != nil && idx >= 0 && idx < len(a.Params) {
		return &a.Params[idx], nil
	}
	return nil, cor.Errorf("no param with idx %d", idx)
}

// ParamByKey returns a pointer to the field and its idex at key or an error.
func (a *Info) ParamByKey(key string) (*Param, int, error) {
	if a != nil {
		for i, f := range a.Params {
			if f.Key() == key {
				return &a.Params[i], i, nil
			}
		}
	}
	return nil, -1, cor.Errorf("no param with key %s", key)
}

// Param represents an type parameter with a name and type.
type Param struct {
	Name string `json:"name,omitempty"`
	Type `json:"typ,omitempty"`
}

// Opt returns true if the param is optional, indicated by its name ending in a question mark.
func (a Param) Opt() bool { n := a.Name; return n != "" && n[len(n)-1] == '?' }

// Key returns the lowercase param key.
func (a Param) Key() string { return cor.LastKey(a.Name) }

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
func (a *Type) UnmarshalJSON(raw []byte) error {
	var tmp struct{ Typ string }
	err := json.Unmarshal(raw, &tmp)
	if err != nil {
		return err
	}
	t, err := ParseString(tmp.Typ)
	if err != nil {
		return err
	}
	*a = t
	return nil
}

func (a Type) WriteBfr(b *bfr.Ctx) error {
	if b.JSON {
		b.WriteString(`{"typ":"`)
		bb := *b
		bb.JSON = false
		err := a.writeBfr(&bb, nil, nil)
		b.WriteString(`"}`)
		return err
	}
	return a.writeBfr(b, nil, nil)
}

func (a Type) writeBfr(b *bfr.Ctx, pre *strings.Builder, hist []*Info) error {
	var detail bool
	switch a.Kind & MaskRef {
	case KindRec, KindObj:
		for i := 0; i < len(hist); i++ {
			h := hist[len(hist)-1-i]
			if a.Info.Equal(h) {
				writeRef(b, pre, '~', strconv.Itoa(i), a)
				return nil
			}
		}
	case KindList, KindDict:
		if pre == nil {
			pre = &strings.Builder{}
		} else {
			pre.WriteByte('|')
		}
		pre.WriteString(a.Kind.String())
		return a.Elem().writeBfr(b, pre, hist)
	}
	switch a.Kind & MaskRef {
	case KindRef:
		ref := ""
		if a.Info != nil {
			ref = a.Ref
		}
		writeRef(b, pre, '@', ref, a)
		return nil
	case KindRec, KindExp, KindAlt:
		detail = true
		fallthrough
	case KindFlag, KindEnum, KindObj:
		b.WriteByte('(')
		err := writePre(b, pre, a)
		if err != nil {
			return err
		}
		err = a.Info.writeXelf(b, detail, append(hist, a.Info))
		b.WriteByte(')')
		return err
	}
	return writePre(b, pre, a)
}

func writePre(b *bfr.Ctx, pre *strings.Builder, a Type) error {
	if pre != nil {
		b.WriteString(pre.String())
		if a == Any {
			return nil
		}
		b.WriteByte('|')
	}
	return a.Kind.WriteBfr(b)
}

func writeRef(b *bfr.Ctx, pre *strings.Builder, x byte, ref string, a Type) {
	if pre != nil {
		b.WriteString(pre.String())
		b.WriteByte('|')
	}
	b.WriteByte(x)
	b.WriteString(ref)
	if a.Kind&FlagOpt != 0 {
		b.WriteByte('?')
	}
}

func (a *Info) writeXelf(b *bfr.Ctx, detail bool, hist []*Info) error {
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
		err := f.Type.writeBfr(b, nil, hist)
		if err != nil {
			return err
		}
	}
	return nil
}
