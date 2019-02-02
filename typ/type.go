package typ

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
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

type Const = cor.Const

// Info represents the reference name and obj fields for the ref and obj types.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Fields []Field `json:"fields,omitempty"`
	Consts []Const `json:"consts,omitempty"`
	key    string
}

// Key returns the lowercase ref key
func (a *Info) Key() string {
	if a == nil {
		return ""
	}
	if a.Ref != "" && a.key == "" {
		a.key = strings.ToLower(a.Ref)
	}
	return a.key
}

var errFieldNotFound = errors.New("field not found")

// FieldByIdx returns a pointer to the field at idx or an error
func (a *Info) FieldByIdx(idx int) (*Field, error) {
	if a != nil && idx >= 0 && idx < len(a.Fields) {
		return &a.Fields[idx], nil
	}
	return nil, errFieldNotFound
}

// FieldByKey returns a pointer to the field with key or an error
func (a *Info) FieldByKey(key string) (*Field, int, error) {
	if a != nil {
		for i, f := range a.Fields {
			if f.Key() == key {
				return &a.Fields[i], i, nil
			}
		}
	}
	return nil, -1, errFieldNotFound
}

// Field represents an obj field with a name and type.
type Field struct {
	Name string `json:"name,omitempty"`
	Type
	key string
}

// Opt returns true if the field is optional, indicated by its name ending in a question mark.
func (a Field) Opt() bool { n := a.Name; return n != "" && n[len(n)-1] == '?' }

// Key returns the lowercase field key.
func (a Field) Key() string {
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
	return a == nil || a.Ref == "" && len(a.Fields) == 0 && len(a.Consts) == 0
}

type infoPair = struct{ a, b *Info }

func (a Type) Equal(b Type) bool { return a.equal(b, nil) }
func (a Type) equal(b Type, hist []infoPair) bool {
	return a.Kind == b.Kind && a.Info.equal(b.Info, hist)
}
func (a *Info) Equal(b *Info) bool { return a.equal(b, nil) }
func (a *Info) equal(b *Info, hist []infoPair) bool {
	if a == b {
		return true
	}
	if a.IsZero() {
		return b.IsZero()
	}
	if b.IsZero() ||
		len(a.Fields) != len(b.Fields) ||
		len(a.Consts) != len(b.Consts) ||
		a.Ref != b.Ref && a.Key() != b.Key() {
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
	for i, af := range a.Fields {
		if !af.equal(b.Fields[i], append(hist, p)) {
			return false
		}
	}
	return true
}

func (a Field) Equal(b Field) bool { return a.equal(b, nil) }
func (a Field) equal(b Field, hist []infoPair) bool {
	return (a.Name == b.Name || a.Key() == b.Key()) && a.Type.equal(b.Type, hist)
}

func (a Type) String() string {
	var b strings.Builder
	a.WriteBfr(bfr.Ctx{B: &b})
	return b.String()
}

func (a Type) WriteBfr(b bfr.Ctx) error {
	return a.writeBfr(b, nil)
}

func (a Type) writeBfr(b bfr.Ctx, hist []*Info) error {
	last := a.Last()
	switch last.Kind & MaskRef {
	case KindObj, KindRec:
		for i := 0; i < len(hist); i++ {
			h := hist[len(hist)-1-i]
			if a.Info.Equal(h) {
				writeArrAndMap(b, a.Kind)
				b.WriteByte('@')
				b.WriteString(strconv.Itoa(i))
				if a.Kind&FlagOpt != 0 {
					b.WriteByte('?')
				}
				return nil
			}
		}
	}

	switch last.Kind & MaskRef {
	case KindRef:
		k := writeArrAndMap(b, a.Kind)
		b.WriteByte('@')
		if a.Info != nil {
			b.WriteString(a.Info.Ref)
		}
		if k&FlagOpt != 0 {
			b.WriteByte('?')
		}
		return nil
	case KindFlag, KindEnum, KindObj, KindRec:
		b.WriteByte('(')
		err := a.Kind.WriteBfr(b)
		if err != nil {
			return err
		}
		for i := 0; i < len(hist); i++ {
			h := hist[len(hist)-1-i]
			if a.Info.Equal(h) {
				b.WriteString(" + @")
				b.WriteString(strconv.Itoa(i))
				return b.WriteByte(')')
			}
		}
		err = a.Info.writeBfr(b, append(hist, a.Info))
		b.WriteByte(')')
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

func (a *Info) writeBfr(b bfr.Ctx, hist []*Info) error {
	if a == nil {
		return nil
	}
	if a.Ref != "" {
		b.WriteByte(' ')
		name, err := lex.Quote(a.Ref, '\'')
		if err != nil {
			return err
		}
		b.WriteString(name)
	}
	for i := 0; i < len(a.Fields); i++ {
		b.WriteByte(' ')
		f := a.Fields[i]
		b.WriteByte('+')
		b.WriteString(f.Name)
		b.WriteByte(' ')
		for _, o := range a.Fields[i+1:] {
			if !f.Type.Equal(o.Type) {
				break
			}
			b.WriteByte('+')
			b.WriteString(o.Name)
			b.WriteByte(' ')
			i++
		}
		err := f.Type.writeBfr(b, hist)
		if err != nil {
			return err
		}
	}
	return nil
}
