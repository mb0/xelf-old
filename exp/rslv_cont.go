package exp

import (
	"reflect"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func init() {
	std.add("reduce", []typ.Param{
		{Name: "a", Type: typ.Any},
		{Name: "unis", Type: typ.Void},
		{Name: "rest", Type: typ.Void},
		{Type: typ.Infer},
	}, rslvReduce)
}

// rslvReduce reduces a container to a single element. The first argument is the initial reducer
// value, after that comes an iterator declaration as documented in iterDecl and one or more
// action elements.
//    (eq (str 'hello alice, bob')
//        (reduce 'hello' +e +i ['alice' 'bob']
//            (cat _ (if i ',') ' ' e)))
// (form 'reduce' +a any +decls +rest - @)
func rslvReduce(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := LayoutArgs(e.Rslv.Arg(), e.Args)
	if err != nil {
		return nil, err
	}
	err = lo.Resolve(c, env)
	if err != nil {
		return e, err
	}
	red := lo.Arg(0)
	decls, err := lo.Unis(1)
	if err != nil {
		return nil, err
	}
	tail := lo.Args(2)
	it, cont, err := iterDecls(c, env, decls)
	if err != nil {
		return e, err
	}
	it.red = "_"
	c = c.WithPart(false)
	switch a := cont.(type) {
	case lit.Keyer:
		i := 0
		err = a.IterKey(func(k string, el Lit) error {
			s := it.newScope(env, el, i, k, red)
			i++
			var res El
			for _, act := range tail {
				res, err = c.Resolve(s, act, typ.Void)
				if err != nil {
					return err
				}
			}
			red = res
			return nil
		})
	case lit.Idxer:
		err = a.IterIdx(func(i int, el Lit) error {
			s := it.newScope(env, el, i, "", red)
			var res El
			for _, act := range tail {
				res, err = c.Resolve(s, act, typ.Void)
				if err != nil {
					return err
				}
			}
			red = res
			return nil
		})
	}
	return red, err
}

// iter is a helper type that holds declaration names for each iteration
type iter struct {
	red, el, idx, key string
}

// newScope returns a scope with the parent env and given literal if their names are not empty.
func (it *iter) newScope(env Env, l Lit, i int, k string, red El) *Scope {
	s := NewScope(env)
	if it.red != "" {
		switch v := red.(type) {
		case Lit:
			s.Def(it.red, LitResolver{v})
		}
	}
	if it.el != "" {
		s.Def(it.el, LitResolver{l})
	}
	if it.idx != "" {
		s.Def(it.idx, LitResolver{lit.Int(i)})
	}
	if it.key != "" {
		s.Def(it.key, LitResolver{lit.Str(k)})
	}
	return s
}

// iterDecls returns an iterator for a declaration slice. it must have at least one and at most
// two for idxer or three for keyer declarations. The declaration refers to the element literal,
// index integer and key string in that order. All declarations must point to the same idxer or
// keyer literal, usually by using naked declarations, decls can be nameless if unused:
//    (reduce '' + +idx +key myobj (cat _ (if idx ', ') key))
func iterDecls(c *Ctx, env Env, decls []Decl) (*iter, Lit, error) {
	if n := len(decls); n < 1 || n > 3 {
		return nil, nil, cor.Errorf("expect one, two or three declaration")
	}
	res := iter{el: decls[0].Name[1:]}
	args := decls[0].Args
	for i, d := range decls[1:] {
		if !reflect.DeepEqual(d.Args[0], args[0]) {
			return nil, nil, cor.Errorf("expect same declaration expression")
		}
		if i == 0 {
			res.idx = d.Name[1:]
		} else {
			res.key = d.Name[1:]
		}
	}
	args, err := c.ResolveAll(env, args, typ.List)
	if err != nil {
		return nil, nil, err
	}
	switch a := args[0].(type) {
	case lit.Keyer:
		return &res, a, nil
	case lit.Idxer:
		if len(decls) > 2 {
			return nil, nil, cor.Errorf("expect one or two declarations for list and arr")
		}
		return &res, a, nil
	}
	return nil, nil, cor.Errorf("unexpect declaration expression in 'each' %T", args[0])
}
