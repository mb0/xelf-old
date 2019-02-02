package gen

import (
	"strings"

	"github.com/mb0/xelf/typ"
	"github.com/pkg/errors"
)

// WriteGoType writes the native go type for t to c or returns an error.
func WriteGoType(c *Ctx, t typ.Type) error {
	k := t.Kind
	if k == typ.KindAny {
		_, err := c.WriteString("interface{}")
		return err
	}
	var r string
	switch k & typ.MaskRef {
	case typ.BaseNum:
		r = "float64"
	case typ.KindBool:
		r = "bool"
	case typ.KindInt:
		r = "int64"
	case typ.KindReal:
		r = "float64"
	case typ.BaseChar, typ.KindStr:
		r = "string"
	case typ.KindRaw:
		r = "[]byte"
	case typ.KindUUID:
		r = "[16]byte"
	case typ.KindTime:
		r = GoImport(c, "time.Time")
	case typ.KindSpan:
		r = GoImport(c, "time.Duration")
	case typ.BaseList:
		_, err := c.WriteString("[]interface{}")
		return err
	case typ.KindArr:
		c.WriteString("[]")
		return WriteGoType(c, t.Next())
	case typ.BaseDict:
		c.WriteString("map[string]interface{}")
		return nil
	case typ.KindMap:
		c.WriteString("map[string]")
		return WriteGoType(c, t.Next())
	case typ.KindObj:
		if k&typ.FlagOpt != 0 {
			c.WriteByte('*')
		}
		if t.Info == nil {
			return typ.ErrInvalid
		}
		c.WriteString("struct {\n")
		for _, f := range t.Info.Fields {
			name, opt := f.Name, f.Opt()
			if opt {
				name = name[:len(name)-1]
			}
			c.WriteByte('\t')
			c.WriteString(name)
			c.WriteByte(' ')
			err := WriteGoType(c, f.Type)
			if err != nil {
				return err
			}
			c.WriteString(" `json:\"")
			c.WriteString(strings.ToLower(name))
			if opt {
				c.WriteString(",omitempty")
			}
			c.WriteString("\"`\n")
		}
		c.WriteRune('}')
		return nil
	case typ.KindFlag, typ.KindEnum, typ.KindRec:
		r = GoImport(c, t.Ref)
	}
	if r == "" {
		return errors.Errorf("type %s cannot be represented in go", t)
	}
	if k&typ.FlagOpt != 0 {
		c.WriteByte('*')
	}
	c.WriteString(r)
	return nil
}
