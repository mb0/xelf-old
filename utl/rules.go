package utl

import (
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var layoutTags = []typ.Param{{Name: "args"}}

// ParseTags parses args as tags and sets them to v using rules or returns an error.
func ParseTags(c *exp.Ctx, env exp.Env, args []exp.El, v interface{}, rules TagRules) error {
	n, err := GetNode(v)
	if err != nil {
		return err
	}
	lo, err := exp.LayoutArgs(layoutTags, args)
	if err != nil {
		return err
	}
	tags, err := lo.Tags(0)
	if err != nil {
		return err
	}
	return rules.Resolve(c, env, tags, n)
}

type (
	// IdxKeyer returns a key for an unnamed tag at idx.
	IdxKeyer = func(n Node, idx int) string
	// KeyPrepper resolves els and returns a literal for key or an error.
	KeyPrepper = func(c *exp.Ctx, env exp.Env, key string, els []exp.El) (lit.Lit, error)
	// KeySetter sets l to node with key or returns an error.
	KeySetter = func(node Node, key string, l lit.Lit) error
)

// KeyRule is a configurable helper for assigning tags or decls to nodes.
type KeyRule struct {
	KeyPrepper
	KeySetter
}

// TagRules is a configurable helper for assigning tags to nodes.
type TagRules struct {
	// Rules holds optional per key rules
	Rules map[string]KeyRule
	// IdxKeyer will map unnamed tags to a key, when null unnamed tags result in an error
	IdxKeyer
	// KeyRule holds optional default rules.
	// If neither specific nor default rules are found DynPrepper and PathSetter are used.
	KeyRule
}

// WithOffset return a with an offset keyer.
func (tr TagRules) WithOffset(off int) *TagRules {
	tr.IdxKeyer = OffsetKeyer(off)
	return &tr
}

// Resolve resolves tags using c and env and assigns them to node or returns an error
func (tr *TagRules) Resolve(c *exp.Ctx, env exp.Env, tags []exp.Tag, node Node) (err error) {
	for i, t := range tags {
		err = tr.ResolveTag(c, env, t, i, node)
		if err != nil {
			return cor.Errorf("resolve tag %s for %T: %w", t.Name, node.Typ(), err)
		}
	}
	return nil
}

// ResolveTag resolves tag using c and env and assigns them to node or returns an error
func (tr *TagRules) ResolveTag(c *exp.Ctx, env exp.Env, tag exp.Tag, idx int, node Node) (err error) {
	var key string
	if tag.Name != "" {
		key = strings.ToLower(tag.Name[1:])
	} else if tr.IdxKeyer != nil {
		key = tr.IdxKeyer(node, idx)
	}
	if key == "" {
		return cor.Errorf("unrecognized tag %s", tag)
	}
	r := tr.Rules[key]
	l, err := tr.prepper(r)(c, env, key, tag.Args)
	if err != nil {
		return err
	}
	return tr.setter(r)(node, key, l)
}

// ZeroKeyer is an index keyer without offset.
var ZeroKeyer = OffsetKeyer(0)

// OffsetKeyer returns an index keyer that looks up a field at the index plus the offset.
func OffsetKeyer(offset int) IdxKeyer {
	return func(n Node, i int) string {
		f, err := n.Typ().ParamByIdx(i + offset)
		if err != nil {
			return ""
		}
		return f.Key()
	}
}

// ListPrepper resolves args using c and env and returns a list or an error.
func ListPrepper(c *exp.Ctx, env exp.Env, _ string, args []exp.El) (lit.Lit, error) {
	args, err := c.ResolveAll(env, args, typ.Any)
	if err != nil {
		return nil, err
	}
	res := make(lit.List, 0, len(args))
	for _, arg := range args {
		res = append(res, arg.(lit.Lit))
	}
	return res, nil
}

// DynPrepper resolves args using c and env and returns a literal or an error.
// Empty args return a untyped null literal. Multiple args are resolved as dyn expression.
func DynPrepper(c *exp.Ctx, env exp.Env, _ string, args []exp.El) (_ lit.Lit, err error) {
	if len(args) == 0 {
		return lit.Nil, nil
	}
	var x exp.El
	if len(args) == 1 {
		x, err = c.Resolve(env, args[0], typ.Void)
	} else {
		x, err = c.Resolve(env, exp.Dyn(args), typ.Void)
	}
	if err != nil {
		return nil, err
	}
	return x.(lit.Lit), nil
}

// PathSetter sets el to n using key as path or returns an error.
func PathSetter(n Node, key string, el lit.Lit) error {
	path, err := lit.ReadPath(key)
	if err != nil {
		return cor.Errorf("read path %s: %w", key, err)
	}
	_, err = lit.SetPath(n, path, el, true)
	if err != nil {
		return cor.Errorf("set path %s: %w", key, err)
	}
	return nil
}

// ExtraMapSetter returns a key setter that tries to add to a node map field with key.
func ExtraMapSetter(mapkey string) KeySetter {
	return func(n Node, key string, el lit.Lit) error {
		err := PathSetter(n, key, el)
		if err != nil && key != mapkey {
			if el == lit.Nil {
				el = lit.True
			}
			err = PathSetter(n, mapkey+"."+key, el)
		}
		return err
	}
}

// FlagPrepper returns a key prepper that tries to resolve a flag constant.
func FlagPrepper(consts []cor.Const) KeyPrepper {
	return func(c *exp.Ctx, env exp.Env, key string, args []exp.El) (lit.Lit, error) {
		l, err := DynPrepper(c, env, key, args)
		if err != nil {
			return l, err
		}
		if l == lit.Nil {
			for _, b := range consts {
				if strings.EqualFold(key, b.Name) {
					return lit.Int(b.Val), nil
				}
			}
			return nil, cor.Errorf("no constant named %q", key)
		}
		n, ok := l.(lit.Numer)
		if !ok {
			return nil, cor.Errorf("expect numer for %q got %T", key, l)
		}
		return lit.Int(n.Num()), nil
	}
}

// FlagSetter returns a key setter that tries to add to a node flag field with key.
func FlagSetter(key string) KeySetter {
	return func(n Node, _ string, el lit.Lit) error {
		f, err := n.Key(key)
		if err != nil {
			return err
		}
		v, ok := f.(lit.Numer)
		if !ok {
			return cor.Errorf("expect int field for %q got %T", key, f)
		}
		w, ok := el.(lit.Int)
		if !ok {
			return cor.Errorf("expect int lit for %q got %T", key, el)
		}
		_, err = n.SetKey(key, lit.Int(uint64(v.Num())|uint64(w)))
		return err
	}
}

func (a KeyRule) prepper(r KeyRule) KeyPrepper {
	if r.KeyPrepper != nil {
		return r.KeyPrepper
	}
	if a.KeyPrepper != nil {
		return a.KeyPrepper
	}
	return DynPrepper
}
func (a KeyRule) setter(r KeyRule) KeySetter {
	if r.KeySetter != nil {
		return r.KeySetter
	}
	if a.KeySetter != nil {
		return a.KeySetter
	}
	return PathSetter
}
