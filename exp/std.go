package exp

var StdEnv = Builtin{Core, Std}

func Core(sym string) Resolver {
	var f ExprResolverFunc
	switch sym {
	case "fail":
		f = rslvFail
	case "if":
		f = rslvIf
	case "dyn":
		f = rslvDyn
	case "as":
		f = rslvAs
	case "or":
		f = rslvOr
	case "and":
		f = rslvAnd
	case "not":
		f = rslvNot
	case "eq", "ne":
		f = rslvEq
	case "equal":
		f = rslvEqual
	case "lt", "ge":
		f = rslvLt
	case "gt", "le":
		f = rslvGt
	case "add":
		f = rslvAdd
	case "sub":
		f = rslvSub
	case "mul":
		f = rslvMul
	case "div":
		f = rslvDiv
	case "rem":
		f = rslvRem
	case "cat":
		f = rslvCat
	case "apd":
		f = rslvApd
	case "set":
		f = rslvSet
	}
	if f == nil {
		return nil
	}
	return f
}

func Std(sym string) Resolver {
	var f ExprResolverFunc
	switch sym {
	case "let":
		f = rslvLet
	case "with":
		f = rslvWith
	case "reduce":
		f = rslvReduce
	}
	if f == nil {
		return nil
	}
	return f
}

type ExprResolverFunc func(*Ctx, Env, *Expr) (El, error)

func (rf ExprResolverFunc) Resolve(c *Ctx, env Env, e El) (El, error) {
	xp, ok := e.(*Expr)
	if !ok {
		return e, ErrUnres
	}
	return rf(c, env, xp)
}
