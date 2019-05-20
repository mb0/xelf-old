package std

import (
	"bytes"
	"strings"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var (
	errCatLit   = cor.StrError("unexpected argument in 'cat' expression")
	errSetKeyer = cor.StrError("expected keyer literal as first argument in 'set' expression")
)

// catSpec concatenates one or more arguments to a str, raw or idxer literal.
var catSpec = core.impl("(form 'cat' (@1:alt str raw idxr) :plain list @2)",
	func(x exp.ReslReq) (exp.El, error) {
		err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
		if err != nil {
			return x.Call, err
		}
		fst := x.Arg(0).(*exp.Atom)
		t, opt := fst.Typ().Deopt()
		var res lit.Lit
		switch t.Kind & typ.MaskRef {
		case typ.KindChar, typ.KindStr:
			var b strings.Builder
			err = catChar(&b, false, fst.Lit, x.Args(1))
			if err != nil {
				return nil, err
			}
			res = lit.Str(b.String())
		case typ.KindRaw:
			var b bytes.Buffer
			err = catChar(&b, true, fst.Lit, x.Args(1))
			if err != nil {
				return nil, err
			}
			res = lit.Raw(b.Bytes())
		default:
			apd, ok := fst.Lit.(lit.Appender)
			if !ok {
				break
			}
			for _, arg := range x.Args(1) {
				aa, ok := arg.(*exp.Atom)
				idxr, ok := aa.Lit.(lit.Indexer)
				if !ok {
					return nil, errCatLit
				}
				err = idxr.IterIdx(func(i int, l lit.Lit) error {
					apd, err = apd.Append(l)
					return err
				})
				if err != nil {
					return nil, err
				}
			}
			return &exp.Atom{Lit: apd}, nil
		}
		if res == nil {
			return nil, cor.Errorf("cannot cat %s", t)
		}
		if opt {
			res = lit.Some{res}
		}
		return &exp.Atom{Lit: res}, nil
	})

// apdSpec appends the rest literal arguments to the first literal appender argument.
var apdSpec = core.impl("(form 'apd' @1:list|@2 :plain list|@2 @1)",
	func(x exp.ReslReq) (exp.El, error) {
		err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
		if err != nil {
			return x.Call, err
		}
		apd, ok := x.Arg(0).(*exp.Atom).Lit.(lit.Appender)
		if !ok {
			return nil, cor.Errorf("cannot append to %T", x.Arg(0))
		}
		for _, arg := range x.Args(1) {
			if a, ok := arg.(*exp.Atom); ok {
				apd, err = apd.Append(a.Lit)
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, cor.Errorf("cannot append arg %T", arg)
		}
		return &exp.Atom{Lit: apd}, nil
	})

// setSpec sets the first keyer literal with the following declaration arguments.
var setSpec = core.impl("(form 'set' @1:keyr|@2 :plain? list|keyr|@2 :tags? dict|@2 @1)",
	func(x exp.ReslReq) (exp.El, error) {
		err := x.Layout.Resolve(x.Ctx, x.Env, x.Hint)
		if err != nil {
			return x.Call, err
		}
		fst := x.Arg(0).(*exp.Atom)
		res, ok := deopt(fst.Lit).(lit.Keyer)
		if !ok {
			return nil, errSetKeyer
		}
		opt := res != fst.Lit
		if len(x.Call.Args) == 1 {
			return fst, nil
		}
		decls, err := x.Unis(2)
		if err != nil {
			return nil, err
		}
		for _, d := range decls {
			el, ok := d.Arg().(*exp.Atom)
			if !ok {
				return nil, cor.Errorf("want literal in declaration got %v", d.El)
			}
			_, err = res.SetKey(d.Key(), el.Lit)
			if err != nil {
				return nil, err
			}
		}
		a := &exp.Atom{Lit: res}
		if opt {
			a.Lit = lit.Some{a.Lit}
		}
		return a, nil
	})

func catChar(b bfr.B, raw bool, fst lit.Lit, args []exp.El) error {
	err := writeChar(b, fst)
	if err != nil {
		return err
	}
	for _, arg := range args {
		a, ok := arg.(*exp.Atom)
		if !ok {
			return cor.Errorf("%s not a literal: %w", arg, errCatLit)
		}
		err := writeChar(b, a.Lit)
		if err != nil {
			return err
		}
	}
	return nil
}
func writeChar(b bfr.B, l lit.Lit) (err error) {
	l = deopt(l)
	c, ok := l.(lit.Character)
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
