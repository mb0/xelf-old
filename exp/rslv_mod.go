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
	plainInfer := []typ.Param{
		{Name: "a", Type: typ.Any},
		{Name: "plain", Type: typ.List},
		{Type: typ.Infer},
	}
	core.add("cat", plainInfer, rslvCat)
	core.add("apd", plainInfer, rslvApd)
	core.add("set", []typ.Param{
		{Name: "a", Type: typ.Dict},
		{Name: "plain", Type: typ.List},
		{Name: "unis", Type: typ.Dict},
		{Type: typ.Infer},
	}, rslvSet)
}

// rslvCat concatenates one or more arguments to a str, raw or idxer literal.
// (form 'cat' +a any +rest list - @)
func rslvCat(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		if err == ErrUnres {
			e.Type, _ = elType(lo.Arg(0))
		}
		return e, err
	}
	fst := lo.Arg(0).(Lit)
	t, opt := fst.Typ().Deopt()
	var res Lit
	switch t.Kind & typ.MaskRef {
	case typ.BaseChar, typ.KindStr:
		var b strings.Builder
		err = catChar(&b, false, fst, lo.Args(1))
		if err != nil {
			return nil, err
		}
		res = lit.Str(b.String())
	case typ.KindRaw:
		var b bytes.Buffer
		err = catChar(&b, true, fst, lo.Args(1))
		if err != nil {
			return nil, err
		}
		res = lit.Raw(b.Bytes())
	default:
		apd, ok := fst.(lit.Appender)
		if !ok {
			break
		}
		for _, arg := range lo.Args(1) {
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
// (form 'apd' +a list +rest list - @)
func rslvApd(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		if err == ErrUnres {
			e.Type, _ = elType(lo.Arg(0))
		}
		return e, err
	}
	apd, ok := lo.Arg(0).(lit.Appender)
	if !ok {
		return nil, cor.Errorf("cannot append to %T", lo.Arg(0))
	}
	for _, arg := range lo.Args(1) {
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
// (form 'set' +a +plain +unis - @)
func rslvSet(c *Ctx, env Env, e *Expr, hint Type) (El, error) {
	lo, err := ResolveArgs(c, env, e)
	if err != nil {
		if err == ErrUnres {
			e.Type, _ = elType(lo.Arg(0))
		}
		return e, err
	}
	fst := lo.Arg(0)
	res, ok := deopt(fst).(lit.Keyer)
	if !ok {
		return nil, errSetKeyer
	}
	opt := res != fst
	if len(e.Args) == 1 {
		return fst, nil
	}
	decls, err := lo.Unis(2)
	if err != nil {
		return nil, err
	}
	for _, d := range decls {
		el, ok := d.Args[0].(Lit)
		if !ok {
			return nil, cor.Error("want literal in declaration argument")
		}
		_, err = res.SetKey(d.Name[1:], el)
		if err != nil {
			return nil, err
		}
	}
	if opt {
		return lit.Some{res}, nil
	}
	return res, nil
}

func catChar(b bfr.B, raw bool, fst Lit, args []El) error {
	err := writeChar(b, fst)
	if err != nil {
		return err
	}
	for _, arg := range args {
		l, ok := arg.(Lit)
		if !ok {
			return cor.Errorf("%s not a literal: %w", arg, errCatLit)
		}
		err := writeChar(b, l)
		if err != nil {
			return err
		}
	}
	return nil
}
func writeChar(b bfr.B, l Lit) (err error) {
	l = deopt(l)
	c, ok := l.(lit.Charer)
	if ok {
		switch v := c.Val().(type) {
		case []byte:
			_, err = b.Write(v)
		case string:
			_, err = b.WriteString(v)
		default:
			_, err = b.WriteString(c.Char())
		}
	} else {
		_, err = b.WriteString(l.String())
	}
	return err
}
