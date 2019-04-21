package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var (
	formAnd  *Spec
	formBool *Spec
	formNot  *Spec
)

func init() {
	core.add("fail", []typ.Param{{Name: "plain"}, {Type: typ.Any}}, rslvFail)
	plainBool := []typ.Param{{Name: "plain", Type: typ.List}, {Type: typ.Bool}}
	core.add("or", plainBool, rslvOr)
	formAnd = core.add("and", plainBool, rslvAnd)
	formBool = core.add("(bool)", plainBool, rslvBool)
	formNot = core.add("not", plainBool, rslvNot)
	core.add("if", []typ.Param{
		{Name: "cond", Type: typ.Any}, {Name: "act"}, {Name: "plain"}, {Type: typ.Infer},
	}, rslvIf)
}

// rslvFail returns an error, if c is an execution context it fails expression string as error,
// otherwise it uses ErrUnres. This is primarily useful for testing.
// (form 'fail' +plain - any)
func rslvFail(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	if c.Exec {
		return nil, cor.Errorf("%s", e)
	}
	return e, ErrUnres
}

// rslvOr resolves the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
// (form 'or' +plain list - bool)
func rslvOr(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := LayoutArgs(e.Spec.Arg(), e.Args)
	if err != nil {
		return nil, err
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
					e = &Call{Def: DefSpec(formBool), Args: e.Args}
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
// (form 'and' +plain - bool)
func rslvAnd(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	lo, err := LayoutArgs(e.Spec.Arg(), e.Args)
	if err != nil {
		return nil, err
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
					e = &Call{Def: DefSpec(formBool), Args: e.Args}
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
// (form 'bool' +plain - bool)
func rslvBool(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	res, err := rslvAnd(c, env, e, hint)
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
}

// rslvNot will resolve the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if a zero value.
// An empty 'not' expression resolves to true.
// (form 'not' +plain - bool)
func rslvNot(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	res, err := rslvAnd(c, env, e, hint)
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
	case formBool:
		if e.Spec == formBool {
			return fst
		}
		f = formNot
	case formNot:
		if e.Spec == formBool {
			return fst
		}
		f = formBool
	default:
		return e
	}
	return &Call{Def: DefSpec(f), Args: fst.Args}
}

// rslvIf resolves the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
// (form +cond any +act any +tail? list - @)
func rslvIf(c *Ctx, env Env, e *Call, hint Type) (El, error) {
	_, err := LayoutArgs(e.Spec.Arg(), e.Args)
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
