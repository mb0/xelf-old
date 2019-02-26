package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func init() {
	core.add("fail", typ.Infer, nil, rslvFail)
	core.add("or", typ.Bool, nil, rslvOr)
	core.add("and", typ.Bool, nil, rslvAnd)
	core.add("bool", typ.Bool, nil, rslvBool)
	core.add("not", typ.Bool, nil, rslvNot)
	core.add("if", typ.Infer, nil, rslvIf)
}

// rslvFail returns an error and set the expressions type to any.
// If c is an execution context it fails expression string as error, otherwise it uses ErrUnres.
//
// This is primarily useful for testing.
func rslvFail(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	e.Type = typ.Any
	if c.Exec {
		return nil, cor.Errorf("%s", e)
	}
	return e, ErrUnres
}

// rslvOr resolves the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
// (form +args? arr|any - bool)
func rslvOr(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	for i, arg := range e.Args {
		el, err := c.Resolve(env, arg, typ.Any)
		if err == ErrUnres {
			e.Type = typ.Bool
			if c.Part {
				e.Args, err = c.WithExec(false).ResolveAll(env, e.Args[i:])
				if err != nil && err != ErrUnres {
					return nil, err
				}
				if len(e.Args) == 1 {
					e = &Expr{Ref: Ref{
						Name: "bool",
						Type: typ.Bool},
						Args:     e.Args,
						Resolver: ExprResolverFunc(rslvBool),
					}
				}
			}
			return e, ErrUnres
		}
		if err != nil {
			return nil, err
		}
		if !el.(Lit).IsZero() {
			return lit.True, nil
		}
	}
	return lit.False, nil
}

// rslvAnd resolves the arguments as short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'and' expression resolves to true.
// (form +args? arr|any - bool)
func rslvAnd(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	for i, arg := range e.Args {
		el, err := c.Resolve(env, arg, typ.Any)
		if err == ErrUnres {
			e.Type = typ.Bool
			if c.Part {
				e.Args, err = c.WithExec(false).ResolveAll(env, e.Args[i:])
				if err != nil && err != ErrUnres {
					return nil, err
				}
				if len(e.Args) == 1 {
					e = &Expr{Ref: Ref{
						Name: "bool",
						Type: typ.Bool},
						Args:     e.Args,
						Resolver: ExprResolverFunc(rslvBool),
					}
				}
			}
			return e, ErrUnres
		}
		if err != nil {
			return nil, err
		}
		if el.(Lit).IsZero() {
			return lit.False, nil
		}
	}
	return lit.True, nil
}

// rslvBool resolves the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'bool' expression resolves to false.
// (form +args? arr|any - bool)
func rslvBool(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	res, err := rslvAnd(c, env, e, hint)
	if err == ErrUnres {
		if c.Part {
			e = simplifyBool(e, res.(*Expr).Args)
		}
		return e, err
	}
	if err != nil {
		return nil, err
	}
	if len(e.Args) == 0 {
		return lit.False, nil
	}
	return lit.Bool(!res.(Lit).IsZero()), nil
}

// rslvNot will resolve the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if a zero value.
// An empty 'not' expression resolves to true.
// (form +args? arr|any - bool)
func rslvNot(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	res, err := rslvAnd(c, env, e, hint)
	if err == ErrUnres {
		if c.Part {
			e = simplifyBool(e, res.(*Expr).Args)
		}
		return e, err
	}
	if err != nil {
		return nil, err
	}
	if len(e.Args) == 0 {
		return lit.True, nil
	}
	return lit.Bool(res.(Lit).IsZero()), nil
}

func simplifyBool(e *Expr, args []El) *Expr {
	e.Args = args
	if len(args) != 1 {
		return e
	}
	fst, ok := args[0].(*Expr)
	if !ok {
		return e
	}
	switch fst.Name {
	case "bool", "not":
	default:
		return e
	}
	if e.Name == "bool" {
		return fst
	}
	res := *fst
	if fst.Name == "bool" {
		res.Name = "not"
	} else {
		res.Name = "bool"
	}
	return &res
}

// rslvIf resolves the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
// (form +cond any +act any +tail? list - @)
func rslvIf(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	// TODO check actions to find a common type
	var i int
	for i = 0; i+1 < len(e.Args); i += 2 {
		cond, err := c.Resolve(env, e.Args[i], typ.Any)
		if err == ErrUnres {
			if c.Part {
				// previous condition turned out false
				e.Args = e.Args[i:]
			}
			return e, err
		}
		if err != nil {
			return nil, err
		}
		if !cond.(Lit).IsZero() {
			return c.Resolve(env, e.Args[i+1], hint)
		}
	}
	if i < len(e.Args) {
		return c.Resolve(env, e.Args[i], hint)
	}
	act, _ := c.WithExec(false).Resolve(env, e.Args[1], hint)
	et, err := elType(act)
	if err != nil || et == typ.Void {
		return nil, cor.Errorf("when else action is omitted then must provide type information")
	}
	return lit.Zero(et), nil
}
