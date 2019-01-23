package typ

import (
	"strings"

	"github.com/mb0/xelf/bfr"
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

// Info represents the reference name and obj fields for the ref and obj types.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Fields []Field `json:"fields,omitempty"`
	Values []Value `json:"values,omitempty"`
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

// Value represents named integer constant for a flag or enum type.
type Value struct {
	Name string `json:"name,omitempty"`
	Val  int64  `json:"val,omitempty"`
	key  string
}

// Key returns the lowercase value key.
func (a Value) Key() string {
	if n := a.Name; n != "" && a.key == "" {
		a.key = strings.ToLower(n)
	}
	return a.key
}

func (a Type) IsZero() bool { return a.Kind == 0 && a.Info.IsZero() }
func (a *Info) IsZero() bool {
	return a == nil || a.Ref == "" && len(a.Fields) == 0 && len(a.Values) == 0
}

func (a Type) Equal(b Type) bool { return a.Kind == b.Kind && a.Info.Equal(b.Info) }
func (a *Info) Equal(b *Info) bool {
	if a.IsZero() {
		return b.IsZero()
	}
	if b.IsZero() ||
		len(a.Fields) != len(b.Fields) ||
		len(a.Values) != len(b.Values) ||
		a.Ref != b.Ref && a.Key() != b.Key() {
		return false
	}
	for i, af := range a.Fields {
		if !af.Equal(b.Fields[i]) {
			return false
		}
	}
	for i, av := range a.Values {
		if !av.Equal(b.Values[i]) {
			return false
		}
	}
	return true
}
func (a Field) Equal(b Field) bool {
	return (a.Name == b.Name || a.Key() == b.Key()) && a.Type.Equal(b.Type)
}
func (a Value) Equal(b Value) bool {
	return (a.Name == b.Name || a.Key() == b.Key()) && a.Val == b.Val
}

func (a Type) String() string {
	var b strings.Builder
	a.WriteBfr(bfr.Ctx{B: &b})
	return b.String()
}

func (a Type) WriteBfr(b bfr.Ctx) error {
	switch a.Last().Kind & MaskRef {
	case KindRef:
		k := a.Kind
	Loop:
		for {
			switch k & MaskElem {
			case KindArr:
				b.WriteString("arr|")
			case KindMap:
				b.WriteString("map|")
			default:
				break Loop
			}
			k = k >> SlotSize
		}
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
		err = a.Info.WriteBfr(b)
		b.WriteByte(')')
		return err
	}
	return a.Kind.WriteBfr(b)
}

func (a *Info) WriteBfr(b bfr.Ctx) error {
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
		err := f.Type.WriteBfr(b)
		if err != nil {
			return err
		}
	}
	return nil
}
