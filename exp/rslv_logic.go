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
	for i := range e.Args {
		args := e.Args[i:]
		err = c.ResolveAll(env, args[:1])
		if err != nil {
			e.Args = args
			return e, err
		}
		l, ok := args[0].(Lit)
		if !ok {
			return nil, fmt.Errorf("unexpected argument in 'or': %T", args[0])
		}
		if !l.IsZero() {
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
	for i := range e.Args {
		args := e.Args[i:]
		err = c.ResolveAll(env, args[:1])
		if err != nil {
			e.Args = args
			return e, err
		}
		l, ok := args[0].(Lit)
		if !ok {
			return nil, fmt.Errorf("unexpected argument in 'and': %T", args[0])
		}
		if l.IsZero() {
			return lit.False, nil
		}
	}
	return lit.True, nil
}

// rslvNot will resolve the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if a zero value.
// An empty 'not' expression resolves to true.
func rslvNot(c *Ctx, env Env, e *Expr) (El, error) {
	res, err := rslvAnd(c, env, e)
	if err != nil {
		return res, err
	}
	if len(e.Args) == 0 {
		return lit.True, nil
	}
	l, ok := res.(Lit)
	if !ok {
		return nil, fmt.Errorf("unexpected argument in 'not': %T", e.Args[0])
	}
	return lit.Bool(l.IsZero()), nil
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
		err = c.ResolveAll(env, e.Args[i:i+1])
		if err != nil {
			return e, err
		}
		cond, ok := e.Args[i].(Lit)
		if !ok {
			return nil, fmt.Errorf("unexpected condition in 'if' expression %T", e.Args[0])
		}
		if !cond.IsZero() {
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
