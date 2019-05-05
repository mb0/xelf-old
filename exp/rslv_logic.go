package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// failSpec returns an error, if c is an execution context it fails expression string as error,
// otherwise it uses ErrUnres. This is primarily useful for testing.
var failSpec = core.impl("(form 'fail' :rest? : any)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		if c.Exec {
			return nil, cor.Errorf("%s", e)
		}
		return e, ErrUnres
	})

// orSpec resolves the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
var orSpec = core.impl("(form 'or' :plain? : bool)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
		if hint != typ.Void {
			typ.Unify(&c.Ctx, hint, typ.Bool)
		}
		args := lo.Args(0)
		for i, arg := range args {
			el, err := c.Resolve(env, arg, typ.Any)
			if err == ErrUnres {
				if c.Part {
					cc := c.WithExec(false)
					e.Args, err = cc.ResolveAll(env, args[i:], typ.Any)
					if err != nil && err != ErrUnres {
						return nil, err
					}
					if len(e.Args) == 1 {
						e = &Call{Spec: boolSpec, Args: e.Args}
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
	})

// andSpec resolves the arguments as short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'and' expression resolves to true.
var andSpec = core.impl("(form 'and' :plain? : bool)", resolveAnd)

func resolveAnd(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
	if hint != typ.Void {
		typ.Unify(&c.Ctx, hint, typ.Bool)
	}
	args := lo.Args(0)
	for i, arg := range args {
		el, err := c.Resolve(env, arg, typ.Any)
		if err == ErrUnres {
			if c.Part {
				e.Args, err = c.WithExec(false).ResolveAll(env, args[i:], typ.Any)
				if err != nil && err != ErrUnres {
					return nil, err
				}
				if len(e.Args) == 1 {
					e = &Call{Spec: boolSpec, Type: e.Type, Args: e.Args}
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

// boolSpec resolves the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'bool' expression resolves to false.
var boolSpec *Spec

// notSpec will resolve the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if a zero value.
// An empty 'not' expression resolves to true.
var notSpec *Spec

func init() {
	boolSpec = core.impl("(form '(bool)' :plain? : bool)", // TODO change to ':bool' ?
		func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
			res, err := resolveAnd(c, env, e, lo, hint)
			if err == ErrUnres {
				if c.Part {
					e = simplifyBool(e, res.(*Call).Args)
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
		})

	notSpec = core.impl("(form 'not' :plain? : bool)",
		func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
			res, err := resolveAnd(c, env, e, lo, hint)
			if err == ErrUnres {
				if c.Part {
					e = simplifyBool(e, res.(*Call).Args)
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
		})
}

func simplifyBool(e *Call, args []El) *Call {
	e.Args = args
	if len(args) != 1 {
		return e
	}
	fst, ok := args[0].(*Call)
	if !ok {
		return e
	}
	var f *Spec
	switch fst.Spec {
	case boolSpec:
		if e.Spec == boolSpec {
			return fst
		}
		f = notSpec
	case notSpec:
		if e.Spec == boolSpec {
			return fst
		}
		f = boolSpec
	default:
		return e
	}
	return &Call{Spec: f, Type: e.Type, Args: fst.Args}
}

// ifSpec resolves the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
var ifSpec = core.impl("(form 'if' bool @ :plain? : @)",
	func(c *Ctx, env Env, e *Call, lo *Layout, hint Type) (El, error) {
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
	})
