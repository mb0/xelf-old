package utl

import (
	"errors"
	"fmt"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

// ParseNode will parse a node form expression and assign the tag values to node.
// It returns the node as a proxy literal and the tail elements or an error.
func ParseNode(c *exp.Ctx, env exp.Env, e *exp.Expr, node interface{}) (lit.Lit, []exp.El, error) {
	p, err := lit.Proxy(node)
	if err != nil {
		return nil, nil, err
	}
	o, ok := p.(lit.Obj)
	if !ok {
		return nil, nil, fmt.Errorf("node must be an object is %T", p)
	}
	tags, tail, err := exp.NodeForm(e.Args)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range tags {
		path, err := lit.ReadPath(v.Name[1:])
		if err != nil {
			return nil, nil, err
		}
		last := path[len(path)-1]
		target, ok := o, true

		if len(path) > 1 {
			pl, err := lit.SelectPath(target, path[:len(path)-1])
			if err != nil {
				return nil, nil, err
			}
			target, ok = pl.(lit.Obj)
			if !ok {
				return nil, nil, errors.New("tag path expects object")
			}
		}
		var x exp.El
		if len(v.Args) > 1 {
			x, err = c.Resolve(env, exp.Dyn(v.Args))
		} else if len(v.Args) == 1 {
			x, err = c.Resolve(env, v.Args[0])
		} else {
			x = lit.Nil
		}
		if err != nil {
			return nil, nil, err
		}
		el, ok := x.(lit.Lit)
		if !ok {
			return nil, nil, fmt.Errorf("expect literal got %T", x)
		}
		err = target.SetKey(last.Key, el)
		if err != nil {
			tt := target.Typ()
			return nil, nil, fmt.Errorf("set key error for typ %s: %v", tt, err)
		}
	}
	return p, tail, nil
}
