package exp

import (
	"fmt"

	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// rslvFail will return an error and set the expressions type to any.
// If c is an execution context it fails expression string as error, otherwise it uses ErrUnres.
//
// This is primarily useful for testing.
func rslvFail(c *Ctx, env Env, e *Expr) (El, error) {
	e.Type = typ.Any
	if c.Exec {
		return nil, fmt.Errorf("%s", e)
	}
	return e, ErrUnres
}

// rslvOr will resolve the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
func rslvOr(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	for i, arg := range e.Args {
		el, err := c.Resolve(env, arg)
		if err == ErrUnres {
			e.Type = typ.Bool
			if c.Part {
				e.Args = e.Args[i:]
				if len(e.Args) == 1 {
					e = &Expr{Sym: Sym{Name: "bool"}, Args: e.Args}
				}
			}
			return e, err
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

// rslvAnd will resolve the arguments as short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'and' expression resolves to true.
func rslvAnd(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsForm(e.Args)
	if err != nil {
		return nil, err
	}
	for i, arg := range e.Args {
		el, err := c.Resolve(env, arg)
		if err == ErrUnres {
			e.Type = typ.Bool
			if c.Part {
				e.Args = e.Args[i:]
				if len(e.Args) == 1 {
					e = &Expr{Sym: Sym{Name: "bool"}, Args: e.Args}
				}
			}
			return e, err
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

// rslvBool will resolve the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'bool' expression resolves to false.
func rslvBool(c *Ctx, env Env, e *Expr) (El, error) {
	res, err := rslvAnd(c, env, e)
	if err == ErrUnres {
		if c.Part {
			e.Args = res.(*Expr).Args
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
func rslvNot(c *Ctx, env Env, e *Expr) (El, error) {
	res, err := rslvAnd(c, env, e)
	if err == ErrUnres {
		if c.Part {
			e.Args = res.(*Expr).Args
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

// rslvIf will resolve the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
func rslvIf(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 2)
	if err != nil {
		return nil, err
	}
	// TODO check actions to find a common type
	var i int
	for i = 0; i+1 < len(e.Args); i += 2 {
		cond, err := c.Resolve(env, e.Args[i])
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
			return c.Resolve(env, e.Args[i+1])
		}
	}
	if i < len(e.Args) {
		return c.Resolve(env, e.Args[i])
	}
	org := c.Exec
	c.Exec = false
	act, _ := c.Resolve(env, e.Args[1])
	c.Exec = org
	et, err := elType(act)
	if err != nil || et == typ.Void {
		return nil, fmt.Errorf("when else action is omitted then must provide type information")
	}
	return lit.Zero(et), nil
}
