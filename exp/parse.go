package exp

import (
	"errors"

	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ParseString scans and parses string s and returns an element or an error.
func ParseString(s string) (El, error) {
	a, err := lex.Scan(s)
	if err != nil {
		return nil, err
	}
	return Parse(a)
}

// Parse parses the syntax tree a and returns an element or an error.
func Parse(a *lex.Tree) (El, error) {
	switch a.Tok {
	case lex.Num, lex.Str, '[', '{':
		return lit.Parse(a)
	case lex.Sym:
		l, err := lit.ParseSym(a.Val)
		if err == nil {
			return l, nil
		}
		t, err := typ.ParseSym(a.Val)
		if err == nil {
			return t, nil
		}
		return &Ref{Sym{Name: a.Val}}, nil
	case '(':
		if len(a.Seq) == 0 { // reuse void type as void expressions
			return typ.Void, nil
		}
		fst, err := Parse(a.Seq[0])
		if err != nil {
			return nil, err
		}
		var sym Sym
		args := make([]El, 0, len(a.Seq))
		switch f := fst.(type) {
		case Type:
			// check if type definition
			if nr, nf := typ.NeedsInfo(f); nr || nf {
				n, err := typ.ParseInfo(a, nr, nf)
				if err != nil {
					return nil, err
				}
				f.Info = n
				return f, nil
			}
			// or a conversion
			sym = Sym{Name: "as", Type: f}
			args = append(args, f)
		case Lit:
			if len(a.Seq) == 1 {
				return fst, nil
			}
			// default literal resolver
			sym = Sym{Name: "combine", Type: f.Typ()}
			args = append(args, f)
		case *Ref:
			sym = f.Sym
		default:
			args = append(args, fst)
		}
		args, err = decledArgs(args, a.Seq[1:])
		if err != nil {
			return nil, err
		}
		if sym.Name == "" {
			return Dyn(args), nil
		}
		return &Expr{Sym: sym, Args: args}, nil
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func decledArgs(res []El, seq []*lex.Tree) (_ []El, err error) {
	seq, tail := lex.SplitAfter(seq, lex.SymPred(0, func(s string) bool { return s == "-" }))
	head, rest, decls := lex.SplitKeyed(seq, false, lex.IsDecl)
	if len(rest) != 0 {
		return nil, errors.New("unexpected decl tail")
	}
	res, err = taggedArgs(res, head)
	for _, decl := range decls {
		args, err := taggedArgs(nil, decl.Seq)
		if err != nil {
			return nil, err
		}
		res = append(res, Decl{Name: decl.Key, Args: args})
	}
	res, err = taggedArgs(res, tail)
	return res, err
}

func taggedArgs(res []El, seq []*lex.Tree) (_ []El, err error) {
	head, tail, tags := lex.SplitKeyed(seq, true, lex.IsTag)
	res, err = plainArgs(res, head)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		args, err := plainArgs(nil, tag.Seq)
		if err != nil {
			return nil, err
		}
		res = append(res, Tag{Name: tag.Key, Args: args})
	}
	res, err = plainArgs(res, tail)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func plainArgs(res []El, seq []*lex.Tree) ([]El, error) {
	for _, t := range seq {
		e, err := Parse(t)
		if err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}
