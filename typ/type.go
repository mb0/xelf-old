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

func (t Type) Typ() Type { return Typ }

// Info represents the reference name and type parameters or constants.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Params []Param `json:"params,omitempty"`
	Consts Consts  `json:"consts,omitempty"`
}

// Key returns the lowercase ref key.
func (a *Info) Key() string {
	if a != nil {
		return cor.Keyed(a.Ref)
	}
	return ""
}
func (a *Info) HasRef() bool    { return a != nil && len(a.Ref) > 0 }
func (a *Info) HasParams() bool { return a != nil && len(a.Params) > 0 }
func (a *Info) HasConsts() bool { return a != nil && len(a.Consts) > 0 }

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
func (p Param) Opt() bool { n := p.Name; return n != "" && n[len(n)-1] == '?' }

// Key returns the lowercase param key.
func (p Param) Key() string { return cor.LastKey(p.Name) }

func (t Type) IsZero() bool { return t.Kind == 0 && t.Info.IsZero() }
func (a *Info) IsZero() bool {
	return a == nil || a.Ref == "" && len(a.Params) == 0 && len(a.Consts) == 0
}

type infoPair = struct{ a, b *Info }

func (t Type) Equal(o Type) bool { return t.equal(o, nil) }
func (t Type) equal(o Type, hist []infoPair) bool {
	return t.Kind == o.Kind && t.Info.equal(o.Info, t.Kind&KindCtx != 0, hist)
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

func (p Param) Equal(o Param) bool { return p.equal(o, nil) }
func (p Param) equal(o Param, hist []infoPair) bool {
	return (p.Name == o.Name || p.Key() == o.Key()) && p.Type.equal(o.Type, hist)
}

func (t Type) String() string               { return bfr.String(t) }
func (t Type) MarshalJSON() ([]byte, error) { return bfr.JSON(t) }
func (t *Type) UnmarshalJSON(raw []byte) error {
	var tmp struct{ Typ string }
	err := json.Unmarshal(raw, &tmp)
	if err != nil {
		return err
	}
	r, err := Read(strings.NewReader(tmp.Typ))
	if err != nil {
		return err
	}
	*t = r
	return nil
}

func (t Type) WriteBfr(b *bfr.Ctx) error {
	if b.JSON {
		b.WriteString(`{"typ":"`)
		bb := *b
		bb.JSON = false
		err := t.writeBfr(&bb, nil, nil, false)
		b.WriteString(`"}`)
		return err
	}
	fst := !t.Kind.Prom() && (t.Kind&KindMeta == 0 || t.Kind == KindAlt)
	return t.writeBfr(b, nil, nil, fst)
}

func (t Type) writeBfr(b *bfr.Ctx, pre *strings.Builder, hist []*Info, qual bool) error {
	switch t.Kind & MaskRef {
	case KindRec, KindObj:
		for i := 0; i < len(hist); i++ {
			h := hist[len(hist)-1-i]
			if t.Info == h {
				writeRef(b, pre, '~', strconv.Itoa(i), t)
				return nil
			}
		}
	case KindCont, KindIdxr, KindKeyr, KindList, KindDict:
		if pre == nil {
			pre = &strings.Builder{}
		} else {
			pre.WriteByte('|')
		}
		pre.WriteString(t.Kind.String())
		return t.Elem().writeBfr(b, pre, hist, false)
	}
	var detail bool
	switch t.Kind & MaskRef {
	case KindVar:
		n := t.ParamLen()
		if n == 0 {
			break
		}
		if n > 1 {
			b.WriteByte('<')
		}
		err := writePre(b, pre, t, qual)
		if err != nil {
			return err
		}
		if n == 1 {
			b.WriteByte('|')
			c := t.Params[0].Type
			return c.writeBfr(b, nil, nil, false)
		}
		b.WriteString("|alt")
		err = t.Info.writeXelf(b, true, hist)
		b.WriteByte('>')
		return err
	case KindRef, KindSch:
		ref := ""
		if t.HasRef() {
			ref = t.Ref
		}
		var x byte = '@'
		if t.Kind&KindSch == KindSch {
			x = '~'
		}
		writeRef(b, pre, x, ref, t)
		return nil
	case KindRec, KindFunc, KindForm, KindAlt:
		detail = true
		fallthrough
	case KindBits, KindEnum, KindObj:
		b.WriteByte('<')
		err := writePre(b, pre, t, qual)
		if err != nil {
			return err
		}
		err = t.Info.writeXelf(b, detail, append(hist, t.Info))
		b.WriteByte('>')
		return err
	}
	return writePre(b, pre, t, qual)
}

func writePre(b *bfr.Ctx, pre *strings.Builder, t Type, qual bool) error {
	if qual {
		b.WriteByte('~')
	}
	if pre != nil {
		b.WriteString(pre.String())
		if t == Any {
			return nil
		}
		b.WriteByte('|')
	}
	return t.Kind.WriteBfr(b)
}

func writeRef(b *bfr.Ctx, pre *strings.Builder, x byte, ref string, t Type) {
	if pre != nil {
		b.WriteString(pre.String())
		b.WriteByte('|')
	}
	b.WriteByte(x)
	b.WriteString(ref)
	if t.Kind&KindOpt != 0 {
		b.WriteByte('?')
	}
}

func (a *Info) writeXelf(b *bfr.Ctx, detail bool, hist []*Info) error {
	if a == nil {
		return nil
	}
	if a.Ref != "" {
		b.WriteByte(' ')
		b.WriteString(a.Ref)
	}
	if !detail {
		return nil
	}
	var i int
	for ; i < len(a.Params); i++ {
		b.WriteByte(' ')
		f := a.Params[i]
		if f.Name != "" {
			b.WriteString(f.Name)
			if f.Type == Void {
				b.WriteByte(';')
				continue
			}
			b.WriteByte(':')
		}
		err := f.Type.writeBfr(b, nil, hist, false)
		if err != nil {
			return err
		}
	}
	return nil
}
