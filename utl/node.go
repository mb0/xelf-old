package utl

import (
	"reflect"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/prx"
	"github.com/mb0/xelf/typ"
)

// Node is an interface for assignable object literals.
type Node interface {
	lit.Record
	Ptr() interface{}
	Assign(lit.Lit) error
}

// GenNode returns a node for val or an error. It tries to proxy if val is not a Node.
func GetNode(val interface{}) (Node, error) {
	n, ok := val.(Node)
	if !ok {
		p, err := prx.NewProxy(val)
		if err != nil {
			return nil, cor.Errorf("proxy %T: %w", val, err)
		}
		n, ok = p.(Node)
		if !ok {
			return nil, cor.Errorf("want node got %T", p)
		}
	}
	return n, nil
}

// NodeRules is a configurable helper for assigning tags and tail elements to nodes.
type NodeRules struct {
	Tags TagRules
	Decl KeyRule
	Tail KeyRule
}

// NodeResolver is a form resolver that returns nodes of a specific type.
type NodeResolver struct {
	NodeRules
	def   reflect.Value
	reuse bool
}

// NewNodeResolver returns a node resolver using rules and returning new nodes based on v.
func NewNodeResolver(rules NodeRules, v interface{}) *NodeResolver {
	def := reflect.ValueOf(v)
	t := def.Type()
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic(cor.Error("node resolver only works with a pointer to a struct type"))
	}
	return &NodeResolver{rules, def, false}
}

func (r *NodeResolver) getNode() (Node, error) {
	v := r.def
	if !r.reuse {
		v = reflect.New(v.Type().Elem())
		if !r.def.IsNil() {
			v.Elem().Set(r.def.Elem())
		}
	}
	p, err := prx.ProxyValue(v)
	if err != nil {
		return nil, cor.Errorf("proxy for %s: %w", v.Type(), err)
	}
	return p.(Node), nil
}

func (r *NodeResolver) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.El, error) {
	node, err := r.getNode()
	if err != nil {
		return nil, err
	}
	var decls []*exp.Named
	// associate to arguments using rules
	fps := c.Spec.Arg()
	for i, fp := range fps {
		switch fp.Key() {
		case "plain", "tags", "args":
			err = r.Tags.Resolve(p, env, c.Tags(i), node)
			if err != nil {
				return nil, err
			}
		case "rest", "tail":
			if r.Tail.KeySetter != nil {
				tail := c.Args(i)
				named := &exp.Named{Name: "::", El: &exp.Dyn{Els: tail}}
				l, err := r.Tail.prepper(KeyRule{})(p, env, named)
				if err != nil {
					return nil, err
				}
				err = r.Tail.KeySetter(node, named.Key(), l)
				if err != nil {
					return nil, err
				}
			} else {
				err = r.Tags.Resolve(p, env, c.Tags(i), node)
				if err != nil {
					return nil, err
				}
			}
		case "unis":
			decls, err = c.Unis(i)
			if err != nil {
				return nil, err
			}
		case "decls":
			decls, err = c.Decls(i)
			if err != nil {
				return nil, err
			}
		default:
			t := &exp.Named{Name: fp.Name, El: &exp.Dyn{Els: c.Args(i)}}
			r.Tags.ResolveTag(p, env, t, i, node)
		}
	}
	for _, d := range decls {
		l, err := r.Decl.prepper(KeyRule{})(p, env, d)
		if err != nil {
			return nil, err
		}
		if r.Decl.KeySetter == nil {
			return nil, cor.Errorf("unexpected decl %s", d)
		}
		err = r.Decl.KeySetter(node, d.Key(), l)
		if err != nil {
			return nil, err
		}
	}
	return &exp.Atom{Lit: node}, nil
}

func (r *NodeResolver) Eval(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.El, error) {
	return r.Resl(p, env, c, h)
}

var refNode = reflect.TypeOf((*Node)(nil)).Elem()
