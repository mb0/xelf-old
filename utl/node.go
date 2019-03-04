package utl

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// Node is an interface for assignable object literals.
type Node interface {
	lit.Obj
	Ptr() interface{}
	Assign(lit.Lit) error
}

// GenNode returns a node for val or an error. It tries to proxy if val is not a Node.
func GetNode(val interface{}) (Node, error) {
	n, ok := val.(Node)
	if !ok {
		p, err := lit.Proxy(val)
		if err != nil {
			return nil, err
		}
		n, ok = p.(Node)
		if !ok {
			return nil, cor.Errorf("want node got %T", p)
		}
	}
	return n, nil
}

var layoutNode = []typ.Param{{Name: "tags"}, {Name: "rest"}}
var layoutFull = []typ.Param{{Name: "args"}, {Name: "decls"}, {Name: "rest"}}

// ParseNode parses args as node form and sets them to v using rules or returns an error.
func ParseNode(c *exp.Ctx, env exp.Env, args []exp.El, v interface{}, rules NodeRules) error {
	n, err := GetNode(v)
	if err != nil {
		return err
	}
	lo, err := exp.LayoutArgs(layoutNode, args)
	if err != nil {
		return err
	}
	tags, err := lo.Tags(0)
	if err != nil {
		return err
	}
	rest := lo.Args(1)
	if rules.Tags.IdxKeyer == nil {
		rules.Tags.IdxKeyer = ZeroKeyer
	}
	return rules.Resolve(c, env, tags, nil, rest, n)
}

// ParseDeclNode parses args as full form and sets them to v using rules or returns an error.
func ParseDeclNode(c *exp.Ctx, env exp.Env, args []exp.El, v interface{}, rules NodeRules) error {
	n, err := GetNode(v)
	if err != nil {
		return err
	}
	lo, err := exp.LayoutArgs(layoutFull, args)
	if err != nil {
		return err
	}
	tags, err := lo.Tags(0)
	if err != nil {
		return err
	}
	decls, err := lo.Decls(1)
	if err != nil {
		return err
	}
	rest := lo.Args(2)
	if rules.Tags.IdxKeyer == nil {
		rules.Tags.IdxKeyer = ZeroKeyer
	}
	return rules.Resolve(c, env, tags, decls, rest, n)
}

// NodeRules is a configurable helper for assigning tags and tail elements to nodes.
type NodeRules struct {
	Tags TagRules
	Decl KeyRule
	Tail KeyRule
}

// Resolve resolves tags, decls and tail and assigns them to node or returns an error.
func (nr NodeRules) Resolve(c *exp.Ctx, env exp.Env,
	tags []exp.Tag, decls []exp.Decl, tail []exp.El, node Node) error {
	err := nr.Tags.Resolve(c, env, tags, node)
	if err != nil {
		return err
	}
	for _, d := range decls {
		l, err := nr.Decl.prepper(KeyRule{})(c, env, d.Name[1:], d.Args)
		if err != nil {
			return err
		}
		if nr.Decl.KeySetter == nil {
			return cor.Errorf("unexpected decl %s", d)
		}
		err = nr.Decl.KeySetter(node, d.Name[1:], l)
		if err != nil {
			return err
		}
	}
	if nr.Tail.KeySetter != nil {
		l, err := nr.Tail.prepper(KeyRule{})(c, env, "::", tail)
		if err != nil {
			return err
		}
		return nr.Tail.KeySetter(node, "::", l)
	}
	return ParseTags(c, env, tail, node, nr.Tags)
}
