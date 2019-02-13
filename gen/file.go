package gen

import (
	"fmt"
	"go/format"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/typ"
	"github.com/pkg/errors"
)

// GoImport takes a qualified name of the form 'pkg.Decl', looks up a path from context packages
// map if available, otherwise the name is used as path. If the package path is the same as the
// context package it returns the 'Decl' part. Otherwise it adds the package path to the import
// list and returns a substring starting with last package path segment: 'pkg.Decl'.
func GoImport(c *Ctx, name string) string {
	idx := strings.LastIndexByte(name, '.')
	var ns string
	if idx > -1 {
		ns = name[:idx]
	}
	if ns != "" && c != nil {
		if path, ok := c.Pkgs[ns]; ok {
			ns = path
		}
		if ns != c.Pkg {
			c.Imports.Add(ns)
		} else {
			return name[idx+1:]
		}
	}
	if idx := strings.LastIndexByte(name, '/'); idx != -1 {
		return name[idx+1:]
	}
	return name
}

// WriteGoFile writes the elements to a go file with package and import declarations.
//
// For now only flag, enum and rec type definitions are supported
func WriteGoFile(c *Ctx, els []exp.El) error {
	b := bfr.Get()
	defer bfr.Put(b)
	// swap new buffer with context buffer
	f := c.B
	c.B = b
	for _, el := range els {
		c.WriteString("\n")
		switch v := el.(type) {
		case typ.Type:
			err := DeclareGoType(c, v)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported element %s", el)
		}
	}
	// swap back
	c.B = f
	f.WriteString(c.Header)
	f.WriteString("package ")
	f.WriteString(pkgName(c.Pkg))
	f.WriteString("\n")
	if len(c.Imports.List) > 0 {
		f.WriteString("\nimport (\n")
		for _, im := range c.Imports.List {
			f.WriteString("\t\"")
			f.WriteString(im)
			f.WriteString("\"\n")
		}
		f.WriteString(")\n")
	}
	res, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}
	for len(res) > 0 {
		n, err := f.Write(res)
		if err != nil {
			return err
		}
		res = res[n:]
	}
	return nil
}

// DeclareGoType writes a type declaration for flag, enum and rec types.
// For flag and enum types the declaration includes the constant declarations.
func DeclareGoType(c *Ctx, t typ.Type) (err error) {
	ref := refName(t)
	switch k := t.Kind; k & typ.MaskRef {
	case typ.KindFlag:
		c.WriteString("type ")
		c.WriteString(ref)
		c.WriteString(" uint64\n\n")
		writeFlagConsts(c, t, ref)
	case typ.KindEnum:
		c.WriteString("type ")
		c.WriteString(ref)
		c.WriteString(" string\n\n")
		writeEnumConsts(c, t, ref)
	case typ.KindRec:
		c.WriteString("type ")
		c.WriteString(refName(t))
		c.WriteByte(' ')
		t.Kind &^= typ.FlagRef
		err = WriteGoType(c, t)
		c.WriteByte('\n')
	default:
		err = errors.Errorf("type %s cannot be declared", t)
	}
	return err
}

func pkgName(pkg string) string {
	if idx := strings.LastIndexByte(pkg, '/'); idx != -1 {
		pkg = pkg[idx+1:]
	}
	if idx := strings.IndexByte(pkg, '.'); idx != -1 {
		pkg = pkg[:idx]
	}
	return pkg
}

func refName(t typ.Type) string {
	if t.Info == nil {
		return ""
	}
	n := t.Ref
	if i := strings.LastIndexByte(n, '.'); i >= 0 {
		n = n[i+1:]
	}
	if len(n) > 0 {
		if c := n[0]; c < 'A' || -c > 'Z' {
			n = strings.ToUpper(n[:1]) + n[1:]
		}
	}
	return n
}

func writeFlagConsts(c *Ctx, t typ.Type, ref string) {
	mono := true
	c.WriteString("const (")
	for i, cst := range t.Consts {
		c.WriteString("\n\t")
		c.WriteString(ref)
		c.WriteString(cst.Name)
		mask := uint64(cst.Val)
		mono = mono && mask == (1<<uint64(i))
		if mono {
			if i == 0 {
				c.WriteByte(' ')
				c.WriteString(ref)
				c.WriteString(" = 1 << iota")
			}
		} else {
			c.WriteString(" = ")
			for j, cr := range cor.GetFlags(t.Consts[:i], mask) {
				if j != 0 {
					c.WriteString(" | ")
				}
				c.WriteString(ref)
				c.WriteString(cr.Name)
			}
		}
	}
	c.WriteString("\n)\n")
}

func writeEnumConsts(c *Ctx, t typ.Type, ref string) {
	c.WriteString("const (")
	for _, cst := range t.Consts {
		c.WriteString("\n\t")
		c.WriteString(ref)
		c.WriteString(cst.Name)
		c.WriteByte(' ')
		c.WriteString(ref)
		c.WriteString(" = \"")
		c.WriteString(strings.ToLower(cst.Name))
		c.WriteByte('"')
	}
	c.WriteString("\n)\n")
}
