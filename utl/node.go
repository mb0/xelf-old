package utl

import (
	"fmt"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
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
			return nil, fmt.Errorf("want node got %T", p)
		}
	}
	return n, nil
}

// ParseNode parses args as node form and sets them to v using rules or returns an error.
func ParseNode(c *exp.Ctx, env exp.Env, args []exp.El, v interface{}, rules NodeRules) error {
	n, err := GetNode(v)
	if err != nil {
		return err
	}
	tags, tail, err := exp.NodeForm(args)
	if err != nil {
		return err
	}

	if rules.Tags.IdxKeyer == nil {
		rules.Tags.IdxKeyer = ZeroKeyer
	}
	return rules.Resolve(c, env, tags, tail, n)
}

// NodeRules is a configurable helper for assigning tags and tail elements to nodes.
type NodeRules struct {
	Tags TagRules
	KeyRule
}

// Resolve resolves tags and tail using c and env and assigns them to node or returns an error
func (nr NodeRules) Resolve(c *exp.Ctx, env exp.Env, tags []exp.Tag, tail []exp.El, node Node) error {
	err := nr.Tags.Resolve(c, env, tags, node)
	if err != nil {
		return err
	}
	l, err := nr.prepper(KeyRule{})(c, env, "::", tail)
	if err != nil {
		return err
	}
	if nr.KeySetter == nil {
		return fmt.Errorf("unexpected tail %s", l)
	}
	return nr.KeySetter(node, "::", l)
}
