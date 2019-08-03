package std

import (
	"strconv"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/typ"
)

var withSpec = core.add(SpecX("(form 'with' any :rest list|expr @)",
	func(x CallCtx) (exp.El, error) {
		dot := x.Arg(0)
		el, err := x.Ctx.Resolve(x.Env, dot, typ.Void)
		if err != nil {
			return x.Call, err
		}
		env := &exp.DataScope{x.Env, el.(*exp.Atom).Lit}
		rest := x.Args(1)
		if len(rest) == 0 {
			return nil, cor.Errorf("with must have body expressions")
		}
		rest, err = x.ResolveAll(env, rest, typ.Void)
		if err != nil {
			return x.Call, err
		}
		last := rest[len(rest)-1]
		if !x.Exec {
			ps := x.Call.Type.Params
			p := &ps[len(ps)-1]
			p.Type = last.Typ()
			return x.Call, nil
		}
		return last, nil
	}))

// letSpec declares one or more resolvers in a new scope and resolves the tailing actions.
// It returns the last actions result.
var letSpec = decl.add(SpecX("(form 'let' :tags dict|any :plain list|expr @)",
	func(x CallCtx) (exp.El, error) {
		decls, err := x.Unis(0)
		if err != nil {
			return nil, err
		}
		rest := x.Args(1)
		if len(rest) == 0 || len(decls) == 0 {
			return nil, cor.Errorf("let must have declarations and a body")
		}
		s := exp.NewScope(x.Env)
		if len(decls) > 0 {
			res, err := letDecls(x.Ctx, s, decls)
			if err != nil {
				return x.Call, err
			}
			if len(rest) == 0 {
				return res, nil
			}
		}
		rest, err = x.ResolveAll(s, rest, typ.Void)
		if err != nil {
			return x.Call, err
		}
		last := rest[len(rest)-1]
		if !x.Exec {
			ps := x.Call.Type.Params
			p := &ps[len(ps)-1]
			p.Type = last.Typ()
			return x.Call, nil
		}
		return last, nil
	}))

// fnSpec declares a function literal from its arguments.
var fnSpec = decl.add(SpecX("(form 'fn' :tags? dict|typ :plain list|expr @)",
	func(x CallCtx) (exp.El, error) {
		tags, err := x.Unis(0)
		if err != nil {
			return nil, err
		}
		rest := x.Args(1)
		if len(tags) > 0 {
			// construct sig from decls
			fs := make([]typ.Param, 0, len(tags))
			for _, d := range tags {
				l, err := x.Ctx.Resolve(x.Env, d.El, typ.Typ)
				if err != nil {
					return x.Call, err
				}
				dt, ok := l.(*exp.Atom).Lit.(typ.Type)
				if !ok {
					return nil, cor.Errorf("want type in func parameters got %T", l)
				}
				fs = append(fs, typ.Param{Name: d.Name[1:], Type: dt})
			}
			return &exp.Atom{&exp.Spec{typ.Func("", fs), &exp.ExprBody{rest, x.Env}}, x.Call.Source()}, nil
		}
		last := rest[len(rest)-1]
		// the last action's type is resolved in a mock function scope, that collects all
		// parameters on lookup and resolves them to a new type variable. the type variables
		// are then unified using the usual resolution process and can be collect for the
		// signature afterwards.
		res := x.New()
		mock := &mockScope{par: x.Env, ctx: x.Ctx.Ctx}
		x.With(true, false).Resolve(mock, last, res)
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

func letDecls(c *exp.Ctx, env *exp.Scope, decls []*exp.Named) (res exp.El, err error) {
	for _, d := range decls {
		if len(d.Name) < 2 {
			return nil, cor.Error("unnamed declaration")
		}
		if d.El == nil {
			return nil, cor.Error("naked declaration")
		}
		res, err = c.Resolve(env, d.El, typ.Void)
		if err != nil {
			return nil, err
		}
		switch a := res.(type) {
		case *exp.Atom:
			err = env.Def(d.Key(), exp.NewDef(a.Lit))
		default:
			return nil, cor.Errorf("unexpected element as declaration value %v", res)
		}
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
