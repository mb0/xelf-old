package exp

import (
	"bytes"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var (
	errCatLit   = cor.StrError("unexpected argument in 'cat' expression")
	errSetKeyer = cor.StrError("expected keyer literal as first argument in 'set' expression")
)

func init() {
	core.add("cat", typ.Infer, nil, rslvCat)
	core.add("apd", typ.Infer, nil, rslvApd)
	core.add("set", typ.Infer, nil, rslvSet)
}

// rslvCat concatenates one or more arguments to a str, raw or idxer literal.
// (form +a (poly str raw list) +b? list - @a)
func rslvCat(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 1)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
	if err != nil {
		if err != ErrUnres {
			return nil, err
		}
		t, err := elType(args[0])
		if err == nil {
			if t.Kind&typ.MaskElem == typ.BaseChar {
				t.Kind |= typ.KindStr
			}
			e.Type = t
		}
		return e, ErrUnres
	}
	t := args[0].(Lit).Typ()
	t, opt := t.Deopt()
	var res Lit
	switch t.Kind & typ.MaskRef {
	case typ.BaseChar, typ.KindStr:
		var b strings.Builder
		err = catChar(&b, false, args)
		if err != nil {
			return nil, err
		}
		res = lit.Str(b.String())
	case typ.KindRaw:
		var b bytes.Buffer
		err = catChar(&b, true, args)
		if err != nil {
			return nil, err
		}
		res = lit.Raw(b.Bytes())
	default:
		e.Type = typ.Raw
		apd, ok := args[0].(lit.Appender)
		if !ok {
			break
		}
		for _, arg := range args[1:] {
			idxr, ok := arg.(lit.Idxer)
			if !ok {
				return nil, errCatLit
			}
			err = idxr.IterIdx(func(i int, l Lit) error {
				apd, err = apd.Append(l)
				return err
			})
			if err != nil {
				return nil, err
			}
		}
		return apd, nil
	}
	if res == nil {
		return nil, cor.Errorf("cannot cat %s", t)
	}
	if opt {
		res = lit.Some{res}
	}
	return res, nil
}

// rslvApd appends the rest literal arguments to the first literal appender argument.
// (form +a (poly list) +b? @a - @a)
func rslvApd(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 1)
	if err != nil {
		return nil, err
	}
	args, err := c.ResolveAll(env, e.Args)
	if err != nil {
		return e, err
	}
	apd, ok := args[0].(lit.Appender)
	if !ok {
		return nil, cor.Errorf("cannot append to %T", args[0])
	}
	for _, arg := range args[1:] {
		if l, ok := arg.(Lit); ok {
			apd, err = apd.Append(l)
			if err != nil {
				return nil, err
			}
			continue
		}
		return nil, cor.Errorf("cannot append arg %T", arg)
	}
	return apd, nil
}

// rslvSet sets the first keyer literal with the following declaration arguments.
// (form +a (poly dict) +decls? dict - @a)
func rslvSet(c *Ctx, env Env, e *Expr) (El, error) {
	if len(e.Args) == 0 {
		return nil, errSetKeyer
	}
	fst, err := c.Resolve(env, e.Args[0])
	if err != nil {
		if err != ErrUnres {
			return nil, err
		}
		t, err := elType(fst)
		if err == nil {
			e.Type = t
		}
		return e, ErrUnres
	}
	res, ok := deopt(fst).(lit.Keyer)
	if !ok {
		return nil, errSetKeyer
	}
	opt := res != fst
	if len(e.Args) == 1 {
		return fst, nil
	}
	// resolve all arguments
	args, err := c.ResolveAll(env, e.Args[1:])
	if err != nil {
		return e, err
	}
	decls, err := UniDeclForm(args)
	if err != nil {
		return nil, err
	}
	for _, d := range decls {
		el, ok := d.Args[0].(Lit)
		if !ok {
			return nil, cor.Error("want literal in declaration argument")
		}
		err = res.SetKey(d.Name[1:], el)
		if err != nil {
			return nil, err
		}
	}
	if opt {
		return lit.Some{res}, nil
	}
	return res, nil
}

func catChar(b bfr.B, raw bool, args []El) error {
	for _, arg := range args {
		l, ok := arg.(Lit)
		if !ok {
			return cor.Errorf("%v, not a literal", errCatLit)
		}
		k := l.Typ().Kind
		if raw && k&typ.MaskElem == typ.KindRaw {
			v, ok := deopt(arg).(lit.Raw)
			if !ok {
				return cor.Errorf("%v, not a raw literal", errCatLit)
			}
			b.Write(v.Val().([]byte))
		} else if k&typ.BaseChar != 0 {
			v, ok := deopt(arg).(lit.Charer)
			if !ok {
				return cor.Errorf("%v, not a char literal", errCatLit)
			}
			b.WriteString(v.Char())
		} else {
			b.WriteString(l.String())
		}
	}
	return nil
}
