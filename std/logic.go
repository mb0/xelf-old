package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// failSpec returns an error, if c is an execution context it fails expression string as error,
// otherwise it uses ErrUnres. This is primarily useful for testing.
var failSpec = core.implResl("(form 'fail' :rest? list any)",
	func(x exp.ReslReq) (exp.El, error) {
		if x.Exec {
			return nil, cor.Errorf("%s", x.Call)
		}
		return x.Call, exp.ErrUnres
	})

// orSpec resolves the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
var orSpec = core.impl("(form 'or' :plain? list bool)",
	func(x exp.ReslReq) (exp.El, error) {
		if x.Hint != typ.Void {
			typ.Unify(x.Ctx.Ctx, x.Hint, typ.Bool)
		}
		args := x.Args(0)
		for i, arg := range args {
			el, err := x.Ctx.Resolve(x.Env, arg, typ.Any)
			if err == exp.ErrUnres {
				if x.Part {
					cc := x.WithExec(false)
					x.Call.Args, err = cc.ResolveAll(x.Env, args[i:], typ.Any)
					if err != nil && err != exp.ErrUnres {
						return nil, err
					}
					if len(x.Call.Args) == 1 {
						x.Call = &exp.Call{
							Spec: boolSpec,
							Type: boolSpec.Type,
							Args: x.Call.Args,
						}
					}
				}
				return x.Call, exp.ErrUnres
			}
			if err != nil {
				return nil, err
			}
			a := el.(*exp.Atom)
			if !a.Lit.IsZero() {
				return &exp.Atom{Lit: lit.True}, nil
			}
		}
		return &exp.Atom{Lit: lit.False}, nil
	})

// andSpec resolves the arguments as short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'and' expression resolves to true.
var andSpec = core.impl("(form 'and' :plain? list bool)", resolveAnd)

func resolveAnd(x exp.ReslReq) (exp.El, error) {
	if x.Hint != typ.Void {
		typ.Unify(x.Ctx.Ctx, x.Hint, typ.Bool)
	}
	args := x.Layout.Args(0)
	for i, arg := range args {
		el, err := x.Ctx.Resolve(x.Env, arg, typ.Any)
		if err == exp.ErrUnres {
			if x.Part {
				c := x.WithExec(false)
				x.Call.Args, err = c.ResolveAll(x.Env, args[i:], typ.Any)
				if err != nil && err != exp.ErrUnres {
					return nil, err
				}
				if len(x.Call.Args) == 1 {
					x.Call = &exp.Call{
						Spec: boolSpec,
						Type: boolSpec.Type,
						Args: x.Call.Args,
					}
				}
			}
			return x.Call, exp.ErrUnres
		}
		if err != nil {
			return nil, err
		}
		a := el.(*exp.Atom)
		if a.Lit.IsZero() {
			return &exp.Atom{Lit: lit.False}, nil
		}
	}
	return &exp.Atom{Lit: lit.True}, nil
}

// boolSpec resolves the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'bool' expression resolves to false.
var boolSpec *exp.Spec

// notSpec will resolve the arguments similar to short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if a zero value.
// An empty 'not' expression resolves to true.
var notSpec *exp.Spec

func init() {
	boolSpec = core.impl("(form ':bool' :plain? list bool)",
		func(x exp.ReslReq) (exp.El, error) {
			res, err := resolveAnd(x)
			if err != nil {
				if err == exp.ErrUnres && x.Part {
					x.Call = simplifyBool(x.Call, res.(*exp.Call).Args)
				}
				return x.Call, err
			}
			a := res.(*exp.Atom)
			if len(x.Call.Args) == 0 {
				a.Lit = lit.False
			} else {
				a.Lit = lit.Bool(!a.Lit.IsZero())
			}
			return a, nil
		})

	notSpec = core.impl("(form 'not' :plain? list bool)",
		func(x exp.ReslReq) (exp.El, error) {
			res, err := resolveAnd(x)
			if err != nil {
				if err == exp.ErrUnres && x.Part {
					x.Call = simplifyBool(x.Call, res.(*exp.Call).Args)
				}
				return x.Call, err
			}
			a := res.(*exp.Atom)
			if len(x.Call.Args) == 0 {
				a.Lit = lit.True
			} else {
				a.Lit = lit.Bool(a.Lit.IsZero())
			}
			return a, nil
		})
}

func simplifyBool(e *exp.Call, args []exp.El) *exp.Call {
	e.Args = args
	if len(args) != 1 {
		return e
	}
	fst, ok := args[0].(*exp.Call)
	if !ok {
		return e
	}
	var f *exp.Spec
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
	return &exp.Call{Spec: f, Args: fst.Args}
}

// ifSpec resolves the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
var ifSpec = core.impl("(form 'if' bool @ :plain? : @)",
	func(x exp.ReslReq) (exp.El, error) {
		// collect all possible action types in an alternative, then choose the common type
		alt := typ.NewAlt()
		ctx := x.WithExec(false)
		var i int
		for i = 0; i+1 < len(x.Call.Args); i += 2 {
			hint := x.New()
			_, err := ctx.Resolve(x.Env, x.Call.Args[i+1], hint)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			alt = typ.Alt(alt, ctx.Apply(hint))
		}
		if i < len(x.Call.Args) {
			hint := x.New()
			_, err := ctx.Resolve(x.Env, x.Call.Args[i], hint)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			alt = typ.Alt(alt, ctx.Apply(hint))
		}
		alt, err := typ.Choose(alt)
		if err == nil && x.Hint != typ.Void {
			alt, err = typ.Unify(ctx.Ctx, alt, x.Hint)
		}
		if err != nil {
			return nil, err
		}
		for i = 0; i+1 < len(x.Call.Args); i += 2 {
			cond, err := x.Ctx.Resolve(x.Env, x.Call.Args[i], typ.Any)
			if err != nil {
				if !x.Part || err != exp.ErrUnres {
					return x.Call, err
				}
				// previous conditions did not match
				x.Call.Args = x.Call.Args[i:]
				x.Ctx = x.WithExec(false)
				return x.Call, err
			}
			a := cond.(*exp.Atom)
			if !a.Lit.IsZero() {
				return x.Ctx.Resolve(x.Env, x.Call.Args[i+1], alt)
			}
		}
		if i < len(x.Call.Args) { // we have an else expression
			return x.Ctx.Resolve(x.Env, x.Call.Args[i], alt)
		}
		return &exp.Atom{Lit: lit.Zero(alt)}, nil
	})
