package utl

import (
	"errors"
	"fmt"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

// ParseNode will parse a node form expression and assign the tag values to node.
// It returns the node as a proxy literal and the tail elements or an error.
func ParseNode(c *exp.Ctx, env exp.Env, e *exp.Expr, node interface{}) (lit.Assignable, []exp.El, error) {
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
	err = AssignTagsObj(c, env, tags, o, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, tail, nil
}

// AssignTagsObj assigns tags to object o or returns an error.
// Unnamed tags are treated as positional arguments and point to field at index+offset
func AssignTagsObj(c *exp.Ctx, env exp.Env, tags []exp.Tag, o lit.Obj, offset int) (err error) {
	for i, v := range tags {
		var x exp.El
		if len(v.Args) > 1 {
			x, err = c.Resolve(env, exp.Dyn(v.Args))
		} else if len(v.Args) == 1 {
			x, err = c.Resolve(env, v.Args[0])
		} else {
			x = lit.Nil
		}
		if err != nil {
			return err
		}
		el, ok := x.(lit.Lit)
		if !ok {
			return fmt.Errorf("expect literal got %T", x)
		}
		if v.Name == "" {
			err = o.SetIdx(i+offset, el)
			if err != nil {
				return fmt.Errorf("set idx error for typ %s: %v", o.Typ(), err)
			}
			return nil
		}
		path, err := lit.ReadPath(v.Name[1:])
		if err != nil {
			return err
		}
		last := path[len(path)-1]
		target, ok := o, true
		if len(path) > 1 {
			pl, err := lit.SelectPath(target, path[:len(path)-1])
			if err != nil {
				return err
			}
			target, ok = pl.(lit.Obj)
			if !ok {
				return errors.New("tag path expects object")
			}
		}
		err = target.SetKey(last.Key, el)
		if err != nil {
			tt := target.Typ()
			return fmt.Errorf("set key error for typ %s: %v", tt, err)
		}
	}
	return nil
}
