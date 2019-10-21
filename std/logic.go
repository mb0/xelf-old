package std

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// failSpec returns an error, if c is an execution context it fails expression string as error,
// otherwise it uses ErrUnres. This is primarily useful for testing.
var failSpec = core.add(SpecDX("(form 'fail' :plain? list any)", func(x CallCtx) (exp.El, error) {
	return nil, cor.Errorf("%s", x.Call)
}))

// orSpec resolves the arguments as short-circuiting logical or to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'or' expression resolves to true.
var orSpec = core.add(SpecDXX("(form 'or' :plain? list bool)",
	func(x CallCtx) (exp.El, error) {
		args := x.Args(0)
		for i, arg := range args {
			el, err := x.Prog.Eval(x.Env, arg, typ.Void)
			if err == exp.ErrUnres {
				args = args[i:]
				x.Groups[0] = args
				if len(args) == 1 {
					x.Spec = boolSpec
					simplifyBool(x.Call)
				}
				return x.Call, err
			}
			if err != nil {
				return nil, err
			}
			a := el.(*exp.Atom)
			if !a.Lit.IsZero() {
				return &exp.Atom{Lit: lit.True, Src: x.Src}, nil
			}
		}
		return &exp.Atom{Lit: lit.False, Src: x.Src}, nil
	}))

// andSpec resolves the arguments as short-circuiting logical 'and' to a bool literal.
// The arguments must be plain literals and are considered true if not a zero value.
// An empty 'and' expression resolves to true.
var andSpec = core.add(SpecDXX("(form 'and' :plain? list bool)", resolveAnd))

func resolveAnd(x CallCtx) (exp.El, error) {
	args := x.Args(0)
	for i, arg := range args {
		el, err := x.Prog.Eval(x.Env, arg, typ.Void)
		if err == exp.ErrUnres {
			args = args[i:]
			x.Groups[0] = args
			if len(args) == 1 {
				if x.Spec.Ref == "and" {
					x.Spec = boolSpec
				}
				simplifyBool(x.Call)
			}
			return x.Call, err
		}
		if err != nil {
			return nil, err
		}
		a := el.(*exp.Atom)
		if a.Lit.IsZero() {
			return &exp.Atom{Lit: lit.False, Src: x.Src}, nil
		}
	}
	return &exp.Atom{Lit: lit.True, Src: x.Src}, nil
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
	boolSpec = core.add(SpecDXX("(form ':bool' :plain? list bool)",
		func(x CallCtx) (exp.El, error) {
			res, err := resolveAnd(x)
			if err != nil {
				return x.Call, err
			}
			a := res.(*exp.Atom)
			if len(x.Args(0)) == 0 {
				a.Lit = lit.False
			} else {
				a.Lit = lit.Bool(!a.Lit.IsZero())
			}
			return a, nil
		}))

	notSpec = core.add(SpecDXX("(form 'not' :plain? list bool)",
		func(x CallCtx) (exp.El, error) {
			res, err := resolveAnd(x)
			if err != nil {
				return x.Call, err
			}
			a := res.(*exp.Atom)
			if len(x.Args(0)) == 0 {
				a.Lit = lit.True
			} else {
				a.Lit = lit.Bool(a.Lit.IsZero())
			}
			return a, nil
		}))
}

func simplifyBool(e *exp.Call) {
	if len(e.Args(0)) != 1 {
		return
	}
	fst, ok := e.Arg(0).(*exp.Call)
	if !ok {
		return
	}
	switch fst.Spec {
	case boolSpec:
		e.Groups = fst.Groups
	case notSpec:
		e.Groups = fst.Groups
		if e.Spec == boolSpec {
			e.Spec = notSpec
		} else {
			e.Spec = boolSpec
		}
	}
}

// ifSpec resolves the arguments as condition, action pairs as part of an if-else condition.
// The odd end is the else action otherwise a zero value of the first action's type is used.
var ifSpec = core.add(SpecRX("(form 'if' ~any ~any :plain? list @)",

	func(x CallCtx) (exp.El, error) {
		// collect all possible action types in an alternative, then choose the common type
		alt := typ.NewAlt()
		var i int
		var unres bool
		all := x.All()
		for i = 0; i+1 < len(all); i += 2 {
			_, err := x.Prog.Resl(x.Env, all[i], typ.Any)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			hint := x.New()
			_, err = x.Prog.Resl(x.Env, all[i+1], hint)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			unres = unres || err == exp.ErrUnres
			alt = typ.Alt(alt, x.Prog.Apply(hint))
		}
		if i < len(all) {
			hint := x.New()
			_, err := x.Prog.Resl(x.Env, all[i], hint)
			if err != nil && err != exp.ErrUnres {
				return nil, err
			}
			unres = unres || err == exp.ErrUnres
			alt = typ.Alt(alt, x.Prog.Apply(hint))
		}
		alt, err := typ.Choose(alt)
		if err != nil {
			return nil, err
		}
		if x.Hint != typ.Void {
			alt, err = typ.Unify(x.Prog.Ctx, alt, x.Hint)
			if err != nil {
				return nil, err
			}
		}
		if alt != typ.Void {
			ps := x.Sig.Params
			ps[len(ps)-1].Type = x.Prog.Inst(alt)
		}
		return x.Call, nil
	},

	func(x CallCtx) (exp.El, error) {
		// collect all possible action types in an alternative, then choose the common type
		var i int
		all := x.All()
		res := x.Call.Res()
		for i = 0; i+1 < len(all); i += 2 {
			cond, err := x.Prog.Eval(x.Env, all[i], typ.Any)
			if err != nil {
				if err != exp.ErrUnres {
					return x.Call, err
				}
				// previous conditions did not match
				x.Groups[0] = all[i : i+1]
				x.Groups[1] = all[i+1 : i+2]
				x.Groups[2] = all[i+2:]
				return x.Call, err
			}
			a, ok := cond.(*exp.Atom)
			if ok && !a.Lit.IsZero() {
				return x.Prog.Eval(x.Env, all[i+1], res)
			}
		}
		if i < len(all) { // we have an else expression
			return x.Prog.Eval(x.Env, all[i], res)
		}
		return &exp.Atom{Lit: lit.Zero(res), Src: x.Src}, nil
	},
))
