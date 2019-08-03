package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// failSpec returns an error, if c is an execution context it fails expression string as error,
// otherwise it uses ErrUnres. This is primarily useful for testing.
var failSpec = core.add(SpecDX("(form 'fail' :rest? list any)",
	func(x CallCtx) (exp.El, error) {
		if x.Exec {
			return nil, cor.Errorf("%s", x.Call)
		}
		return x.Call, exp.ErrUnres
	}))

func reslLogic(x CallCtx) (exp.El, error) {
	_, err := x.Ctx.WithExec(false).ResolveAll(x.Env, x.Args(0), typ.Any)
	t := x.Call.Type
	r := &t.Params[len(t.Params)-1]
	r.Type = typ.Bool
	if x.Hint != typ.Void {
		typ.Unify(x.Ctx.Ctx, x.Hint, typ.Bool)
	}
	return x.Call, err
}

// orSpec resolves the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
var orSpec = core.add(SpecRXX("(form 'or' :plain? list bool)", reslLogic,
	func(x CallCtx) (exp.El, error) {
		args := x.Args(0)
		for i, arg := range args {
			el, err := x.Ctx.Resolve(x.Env, arg, typ.Any)
			if err == exp.ErrUnres {
				if x.Part {
					x.Call.Args = args[i:]
					if len(x.Call.Args) == 1 {
						x.Call.Spec = boolSpec
						simplifyBool(x.Call)
					}
				}
				return x.Call, err
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
	}))

// andSpec resolves the arguments as short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'and' expression resolves to true.
var andSpec = core.add(SpecRXX("(form 'and' :plain? list bool)", reslLogic, resolveAnd))

func resolveAnd(x CallCtx) (exp.El, error) {
	args := x.Layout.Args(0)
	for i, arg := range args {
		el, err := x.Ctx.Resolve(x.Env, arg, typ.Any)
		if err == exp.ErrUnres {
			if x.Part {
				x.Call.Args = args[i:]
				if len(x.Call.Args) == 1 {
					if x.Call.Spec.Ref == "and" {
						x.Call.Spec = boolSpec
					}
					simplifyBool(x.Call)
				}
			}
			return x.Call, err
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
	boolSpec = core.add(SpecRXX("(form ':bool' :plain? list bool)", reslLogic,
		func(x CallCtx) (exp.El, error) {
			res, err := resolveAnd(x)
			if err != nil {
				return x.Call, err
			}
			a := res.(*exp.Atom)
			if len(x.Call.Args) == 0 {
				a.Lit = lit.False
			} else {
				a.Lit = lit.Bool(!a.Lit.IsZero())
			}
			return a, nil
		}))

	notSpec = core.add(SpecRXX("(form 'not' :plain? list bool)", reslLogic,
		func(x CallCtx) (exp.El, error) {
			res, err := resolveAnd(x)
			if err != nil {
				return x.Call, err
			}
			a := res.(*exp.Atom)
			if len(x.Call.Args) == 0 {
				a.Lit = lit.True
			} else {
				a.Lit = lit.Bool(a.Lit.IsZero())
			}
			return a, nil
		}))
}

func simplifyBool(e *exp.Call) {
	if len(e.Args) != 1 {
		return
	}
	fst, ok := e.Args[0].(*exp.Call)
	if !ok {
		return
	}
	switch fst.Spec {
	case boolSpec:
		e.Args = fst.Args
	case notSpec:
		e.Args = fst.Args
		if e.Spec == boolSpec {
			e.Spec = notSpec
		} else {
			e.Spec = boolSpec
		}
	}
}

// ifSpec resolves the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
var ifSpec = core.add(SpecX("(form 'if' ~any ~any :plain? list : @)",
	func(x CallCtx) (exp.El, error) {
		// collect all possible action types in an alternative, then choose the common type
		alt := typ.NewAlt()
		ctx := x.WithExec(false)
		var i int
		var unres bool
		for i = 0; i+1 < len(x.Call.Args); i += 2 {
			hint := x.New()
			_, err := ctx.Resolve(x.Env, x.Call.Args[i+1], hint)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			unres = unres || err == exp.ErrUnres
			alt = typ.Alt(alt, ctx.Apply(hint))
		}
		if i < len(x.Call.Args) {
			hint := x.New()
			_, err := ctx.Resolve(x.Env, x.Call.Args[i], hint)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			unres = unres || err == exp.ErrUnres
			alt = typ.Alt(alt, ctx.Apply(hint))
		}
		alt, err := typ.Choose(alt)
		if err == nil && x.Hint != typ.Void {
			alt, err = typ.Unify(ctx.Ctx, alt, x.Hint)
			if alt != typ.Void {
				ps := x.Call.Type.Params
				ps[len(ps)-1].Type = ctx.Inst(alt)
			}
		}
		if err != nil {
			return nil, err
		}
		exec := x.Part || x.Exec
		for i = 0; i+1 < len(x.Call.Args); i += 2 {
			cond, err := x.Ctx.Resolve(x.Env, x.Call.Args[i], typ.Any)
			if err != nil {
				if !x.Part || err != exp.ErrUnres {
					return x.Call, err
				}
				// previous conditions did not match
				x.Call.Args = x.Call.Args[i:]
				return x.Call, err
			}
			if exec {
				a := cond.(*exp.Atom)
				if !a.Lit.IsZero() {
					return x.Ctx.Resolve(x.Env, x.Call.Args[i+1], alt)
				}
			}
		}
		if !exec {
			if unres {
				return x.Call, exp.ErrUnres
			}
			return x.Call, nil
		}
		if i < len(x.Call.Args) { // we have an else expression
			return x.Ctx.Resolve(x.Env, x.Call.Args[i], alt)
		}
		return &exp.Atom{Lit: lit.Zero(alt)}, nil
	}))
