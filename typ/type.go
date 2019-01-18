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

// Info represents the reference name and obj fields for the ref and obj types.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

// Field represents an obj field with a name and type.
type Field struct {
	Name string `json:"name,omitempty"`
	Type
}

func (a Type) IsZero() bool  { return a.Kind == 0 && a.Info.IsZero() }
func (a *Info) IsZero() bool { return a == nil || a.Ref == "" && len(a.Fields) == 0 }
func (a Field) IsZero() bool { return a.Name == "" && a.Type.IsZero() }

func (a Type) Equal(b Type) bool { return a.Kind == b.Kind && a.Info.Equal(b.Info) }
func (a *Info) Equal(b *Info) bool {
	if a.IsZero() {
		return b.IsZero()
	}
	if b.IsZero() || a.Ref != b.Ref || len(a.Fields) != len(b.Fields) {
		return false
	}
	for i, af := range a.Fields {
		if !af.Equal(b.Fields[i]) {
			return false
		}
	}
	return true
}
func (a Field) Equal(b Field) bool { return a.Name == b.Name && a.Type.Equal(b.Type) }

func (a Type) String() string {
	var b strings.Builder
	a.WriteBfr(bfr.Ctx{B: &b})
	return b.String()
}

func (a Type) WriteBfr(b bfr.Ctx) error {
	switch elemType(a).Kind & MaskRef {
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
	for _, f := range a.Fields {
		b.WriteByte(' ')
		err := f.WriteBfr(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a Field) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('+')
	b.WriteString(a.Name)
	b.WriteByte(' ')
	return a.Type.WriteBfr(b)
}
