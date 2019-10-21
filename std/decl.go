package std

import (
	"strconv"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/typ"
)

var withSpec = core.add(SpecRX("(form 'with' any :act expr @)",
	func(x CallCtx) (exp.El, error) {
		dot := x.Arg(0)
		el, err := x.Prog.Resl(x.Env, dot, typ.Void)
		if err != nil {
			return x.Call, err
		}
		t := elResType(el)
		env := &exp.DataScope{x.Env, exp.Def{Type: t}}
		if a, ok := el.(*exp.Atom); ok {
			env.Lit = a.Lit
		}
		act := x.Arg(1)
		if act == nil {
			return nil, cor.Errorf("with must have an action")
		}
		act, err = x.Prog.Resl(env, act, x.Hint)
		if err != nil {
			return x.Call, cor.Errorf("resl with act: %s", err)
			return x.Call, err
		}
		ps := x.Sig.Params
		p := &ps[len(ps)-1]
		p.Type = elResType(act)
		return x.Call, nil
	},
	func(x CallCtx) (exp.El, error) {
		dot := x.Arg(0)
		el, err := x.Prog.Eval(x.Env, dot, typ.Void)
		if err != nil {
			return x.Call, err
		}
		dl := el.(*exp.Atom).Lit
		env := &exp.DataScope{x.Env, exp.Def{Type: dl.Typ(), Lit: dl}}
		act := x.Arg(1)
		if act == nil {
			return nil, cor.Errorf("with must have an action")
		}
		act, err = x.Prog.Eval(env, act, x.Hint)
		if err != nil {
			return x.Call, err
		}
		return act, nil
	}))

// letSpec declares one or more resolvers in a new scope and resolves the tailing actions.
// It returns the last actions result.
var letSpec = decl.add(SpecRX("(form 'let' :tags dict|any :action expr @)",
	func(x CallCtx) (exp.El, error) {
		act := x.Arg(1)
		decls := x.Tags(0)
		if act == nil || len(decls) == 0 {
			return nil, cor.Errorf("let must have tags and an action")
		}
		s := exp.NewScope(x.Env)
		_, err := reslLetDecls(x.Prog, s, decls)
		if err != nil {
			return x.Call, err
		}
		act, err = x.Prog.Resl(s, act, typ.Void)
		if err != nil {
			return x.Call, err
		}
		ps := x.Sig.Params
		p := &ps[len(ps)-1]
		p.Type = elResType(act)
		return x.Call, nil
	},
	func(x CallCtx) (exp.El, error) {
		act := x.Arg(1)
		decls := x.Tags(0)
		s := exp.NewScope(x.Env)
		_, err := evalLetDecls(x.Prog, s, decls)
		if err != nil {
			return x.Call, err
		}
		res, err := x.Prog.Eval(s, act, typ.Void)
		if err != nil {
			return x.Call, err
		}
		return res, nil
	}))

func elResType(el exp.El) typ.Type {
	switch v := el.(type) {
	case *exp.Sym:
		return v.Type
	case *exp.Atom:
		return v.Lit.Typ()
	case *exp.Call:
		return v.Res()
	}
	return typ.Void
}

// fnSpec declares a function literal from its arguments.
var fnSpec = decl.add(SpecXX("(form 'fn' :tags? dict|typ :plain list|expr @)",
	func(x CallCtx) (exp.El, error) {
		tags := x.Tags(0)
		rest := x.Args(1)
		if len(tags) > 0 {
			// construct sig from decls
			fs := make([]typ.Param, 0, len(tags))
			var naked int
			for _, d := range tags {
				p := typ.Param{Name: d.Name[1:]}
				if d.El == nil {
					naked++
				} else {
					l, err := x.Prog.Resl(x.Env, d.El, typ.Typ)
					if err != nil {
						return x.Call, err
					}
					dt, ok := l.(*exp.Atom).Lit.(typ.Type)
					if !ok {
						return nil, cor.Errorf("want type in func parameters got %T", l)
					}
					for naked > 0 {
						fs[len(fs)-naked].Type = dt
						naked--
					}
					p.Type = dt
				}
				fs = append(fs, p)
			}
			for naked > 0 {
				fs[len(fs)-naked].Type = typ.Any
				naked--
			}
			return &exp.Atom{&exp.Spec{typ.Func("", fs), &exp.ExprBody{rest, x.Env}}, x.Src}, nil
		}
		last := rest[len(rest)-1]
		// the last action's type is resolved in a mock function scope, that collects all
		// parameters on lookup and resolves them to a new type variable. the type variables
		// are then unified using the usual resolution process and can be collect for the
		// signature afterwards.
		res := x.New()
		mock := &mockScope{par: x.Env, ctx: x.Prog.Ctx}
		x.Resl(mock, last, res)
		ps, err := mock.params()
		if err != nil {
			return nil, err
		}
		ps = append(ps, typ.Param{Type: x.Apply(res)})
		spec := &exp.Spec{typ.Func("", ps), &exp.ExprBody{rest, x.Env}}
		return &exp.Atom{Lit: spec}, nil
	}))

type mockScope struct {
	par  exp.Env
	ctx  *typ.Ctx
	syms []string
	smap map[string]typ.Type
}

func (ms *mockScope) Parent() exp.Env      { return ms.par }
func (ms *mockScope) Supports(x byte) bool { return x == '.' }
func (ms *mockScope) Get(s string) *exp.Def {
	if s == "_" {
		s = ".0"
	} else if len(s) < 2 || s[0] != '.' || s[1] == '.' {
		return nil
	}
	t, ok := ms.smap[s]
	if !ok {
		ms.syms = append(ms.syms, s)
		t = ms.ctx.New()
		if ms.smap == nil {
			ms.smap = make(map[string]typ.Type)
		}
		ms.smap[s] = t
	}
	return &exp.Def{Type: t}
}

func (ms *mockScope) params() ([]typ.Param, error) {
	ps := make([]typ.Param, len(ms.syms))
	for _, s := range ms.syms {
		if b := s[1]; b < '0' || b > '9' {
			continue // fill named params later
		}
		idx, err := strconv.Atoi(s[1:])
		if err != nil {
			return nil, err
		}
		ps[idx] = typ.Param{Type: ms.ctx.Apply(ms.smap[s])}
	}
	var search int
	for _, s := range ms.syms {
		if b := s[1]; b >= '0' && b <= '9' {
			continue // already filled index params
		}
		// find next free param slot
		for ; search < len(ps); search++ {
			p := &ps[search]
			search++
			if p.Type.Kind == 0 {
				p.Type = ms.ctx.Apply(ms.smap[s])
				break
			}
		}
	}
	return ps, nil
}

func reslLetDecls(p *exp.Prog, env *exp.Scope, decls []*exp.Named) (res exp.El, err error) {
	for _, d := range decls {
		if len(d.Name) < 2 {
			return nil, cor.Error("unnamed declaration")
		}
		if d.El == nil {
			return nil, cor.Error("naked declaration")
		}
		res, err = p.Resl(env, d.El, typ.Void)
		if err != nil {
			return nil, err
		}
		// TODO remove literal definition
		switch a := res.(type) {
		case *exp.Atom:
			err = env.Def(d.Key(), exp.NewDef(a.Lit))
		default:
			r := elResType(res)
			err = env.Def(d.Key(), &exp.Def{Type: r})
		}
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
func evalLetDecls(p *exp.Prog, env *exp.Scope, decls []*exp.Named) (res exp.El, err error) {
	for _, d := range decls {
		res, err = p.Eval(env, d.El, typ.Void)
		if err != nil {
			return nil, err
		}
		switch a := res.(type) {
		case *exp.Atom:
			err = env.Def(d.Key(), exp.NewDef(a.Lit))
		default:
			err = exp.ErrUnres
		}
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
