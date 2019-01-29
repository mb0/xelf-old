package exp

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var (
	errCatLit   = errors.New("unexpected argument in 'cat' expression")
	errSetKeyer = errors.New("expected keyer literal as first argument in 'set' expression")
)

// rslvCat concatenates one or more arguments to a str, raw or idxer literal.
func rslvCat(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 1)
	if err != nil {
		return nil, err
	}
	err = c.ResolveAll(env, e.Args)
	if err != nil {
		return e, err
	}
	t := e.Args[0].(Lit).Typ()
	t, opt := t.Deopt()
	var res Lit
	switch t.Kind & typ.MaskRef {
	case typ.BaseChar, typ.KindStr:
		var b strings.Builder
		err = catChar(&b, false, e.Args)
		if err != nil {
			return nil, err
		}
		res = lit.Str(b.String())
	case typ.KindRaw:
		var b bytes.Buffer
		err = catChar(&b, true, e.Args)
		if err != nil {
			return nil, err
		}
		res = lit.Raw(b.Bytes())
	default:
		apd, ok := e.Args[0].(lit.Appender)
		if !ok {
			break
		}
		for _, arg := range e.Args[1:] {
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
		return nil, fmt.Errorf("cannot cat %s", t)
	}
	if opt {
		res = lit.Some{res}
	}
	return res, nil
}

// rslvApd appends the rest literal arguments to the first literal appender argument.
func rslvApd(c *Ctx, env Env, e *Expr) (El, error) {
	err := ArgsMin(e.Args, 1)
	if err != nil {
		return nil, err
	}
	err = c.ResolveAll(env, e.Args)
	if err != nil {
		return e, err
	}
	apd, ok := e.Args[0].(lit.Appender)
	if !ok {
		return nil, fmt.Errorf("cannot append to %T", e.Args[0])
	}
	for _, arg := range e.Args[1:] {
		if l, ok := arg.(Lit); ok {
			apd, err = apd.Append(l)
			if err != nil {
				return nil, err
			}
			continue
		}
		return nil, fmt.Errorf("cannot append arg %T", arg)
	}
	return apd, nil
}

// rslvSet sets the first keyer literal arguments with the following declaration arguments.
func rslvSet(c *Ctx, env Env, e *Expr) (El, error) {
	if len(e.Args) == 0 {
		return nil, errSetKeyer
	}
	fst, err := c.Resolve(env, e.Args[0])
	if err != nil {
		return e, err
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
	err = c.ResolveAll(env, e.Args[1:])
	if err != nil {
		return e, err
	}
	decls, err := UniDeclForm(e.Args[1:])
	if err != nil {
		return nil, err
	}
	for _, d := range decls {
		el, ok := d.Args[0].(Lit)
		if !ok {
			return nil, errors.New("want literal in declaration argument")
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
			return fmt.Errorf("%v, not a literal", errCatLit)
		}
		k := l.Typ().Kind
		if raw && k&typ.MaskElem == typ.KindRaw {
			v, ok := deopt(arg).(lit.Raw)
			if !ok {
				return fmt.Errorf("%v, not a raw literal", errCatLit)
			}
			b.Write(v.Val().([]byte))
		} else if k&typ.BaseChar != 0 {
			v, ok := deopt(arg).(lit.Charer)
			if !ok {
				return fmt.Errorf("%v, not a char literal", errCatLit)
			}
			b.WriteString(v.Char())
		} else {
			b.WriteString(l.String())
		}
	}
	return nil
}
