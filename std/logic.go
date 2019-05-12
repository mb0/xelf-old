package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// failSpec returns an error, if c is an execution context it fails expression string as error,
// otherwise it uses ErrUnres. This is primarily useful for testing.
var failSpec = core.impl("(form 'fail' :rest? list any)",
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
						x.Call = &exp.Call{Spec: boolSpec, Args: x.Call.Args}
					}
				}
				return x.Call, exp.ErrUnres
			}
			if err != nil {
				return nil, err
			}
			if !el.(lit.Lit).IsZero() {
				return lit.True, nil
			}
		}
		return lit.False, nil
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
						Args: x.Call.Args,
					}
				}
			}
			return x.Call, exp.ErrUnres
		}
		if err != nil {
			return nil, err
		}
		if el.(lit.Lit).IsZero() {
			return lit.False, nil
		}
	}
	return lit.True, nil
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
	boolSpec = core.impl("(form '(bool)' :plain? list bool)", // TODO change to ':bool' ?
		func(x exp.ReslReq) (exp.El, error) {
			res, err := resolveAnd(x)
			if err == exp.ErrUnres {
				if x.Part {
					x.Call = simplifyBool(x.Call, res.(*exp.Call).Args)
				}
				return x.Call, err
			}
			if err != nil {
				return nil, err
			}
			if len(x.Call.Args) == 0 {
				return lit.False, nil
			}
			return lit.Bool(!res.(lit.Lit).IsZero()), nil
		})

	notSpec = core.impl("(form 'not' :plain? list bool)",
		func(x exp.ReslReq) (exp.El, error) {
			res, err := resolveAnd(x)
			if err == exp.ErrUnres {
				if x.Part {
					x.Call = simplifyBool(x.Call, res.(*exp.Call).Args)
				}
				return x.Call, err
			}
			if err != nil {
				return nil, err
			}
			if len(x.Call.Args) == 0 {
				return lit.True, nil
			}
			return lit.Bool(res.(lit.Lit).IsZero()), nil
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
		// TODO check actions to find a common type
		var i int
		for i = 0; i+1 < len(x.Call.Args); i += 2 {
			cond, err := x.Ctx.Resolve(x.Env, x.Call.Args[i], typ.Any)
			if err == exp.ErrUnres {
				if x.Part {
					// previous condition turned out false
					x.Call.Args = x.Call.Args[i:]
				}
				return x.Call, err
			}
			if err != nil {
				return nil, err
			}
			if !cond.(lit.Lit).IsZero() {
				return x.Ctx.Resolve(x.Env, x.Call.Args[i+1], x.Hint)
			}
		}
		if i < len(x.Call.Args) {
			return x.Ctx.Resolve(x.Env, x.Call.Args[i], x.Hint)
		}
		act, _ := x.WithExec(false).Resolve(x.Env, x.Call.Args[1], x.Hint)
		et, err := elType(act)
		if err != nil || et == typ.Void {
			return nil, cor.Errorf("when else action is omitted then must provide type information")
		}
		return lit.Zero(et), nil
	})

func elType(el exp.El) (typ.Type, error) {
	switch t := el.Typ(); t.Kind {
	case typ.KindTyp:
		return el.(typ.Type), nil
	case typ.KindSym:
		s := el.(*exp.Sym)
		if s.Type != typ.Void {
			return s.Type, nil
		}
	case typ.KindCall:
		x := el.(*exp.Call)
		t := x.Res()
		if t != typ.Void {
			return t, nil
		}
	default:
		return t, nil
	}
	return typ.Void, exp.ErrUnres
}
