package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// rslvLet declares one or more resolvers in the existing scope.
func rslvLet(c *Ctx, env Env, e *Expr) (El, error) {
	decls, err := UniDeclForm(e.Args)
	if err != nil {
		return nil, err
	}
	res, err := letDecls(c, env, decls)
	if err != nil {
		return e, err
	}
	return res, nil
}

// rslvWith declares one or more resolvers in a new scope and resolves the tailing actions.
// It returns the last actions result.
func rslvWith(c *Ctx, env Env, e *Expr) (El, error) {
	decls, tail, err := UniDeclRest(e.Args)
	if err != nil {
		return nil, err
	}
	err = ArgsMin(tail, 1)
	if err != nil {
		return nil, cor.Errorf("%v %v %v", err, decls, tail)
	}
	s := NewScope(env)
	_, err = letDecls(c, s, decls)
	if err != nil {
		return e, err
	}
	tail, err = c.ResolveAll(s, tail)
	if err != nil {
		return e, err
	}
	return tail[len(tail)-1], nil
}

func letDecls(c *Ctx, env Env, decls []Decl) (El, error) {
	var res El
	for _, d := range decls {
		if len(d.Name) < 2 {
			return nil, cor.Error("unnamed declaration")
		}
		args, err := c.ResolveAll(env, d.Args)
		if err != nil {
			return nil, err
		}
		var rslv Resolver
		switch dv := args[0].(type) {
		case Type:
			res = dv
			rslv = typeResolver(dv)
		case Lit:
			res = dv
			rslv = litResolver{dv}
		default:
			return nil, cor.Errorf("unexpected element as declaration value %v", d.Args[0])
		}
		err = env.Def(d.Name[1:], rslv)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// typeResolver is a resolver that wraps a type from package typ.
//
// The wrapped type is expected to be fully resolved.
type typeResolver Type

func (r typeResolver) Resolve(c *Ctx, env Env, e El) (El, error) {
	switch v := e.(type) {
	case *Ref:
		return Type(r), nil
	case Type:
		// must by a type ref
		if v.Last().Kind&typ.MaskRef != typ.KindRef {
			return nil, cor.Errorf("expected type reference got %s", v)
		}
		return Type(r), nil
	}
	return nil, cor.Errorf("expected type or expression reference %s", e)
}

// litResolver is a resolver that wraps a literal from package lit.
type litResolver struct{ Lit }

func (r litResolver) Resolve(c *Ctx, env Env, e El) (El, error) {
	switch e.(type) {
	case *Ref:
		return r.Lit, nil
	}
	return nil, cor.Errorf("expected expression reference %s", e)
}
