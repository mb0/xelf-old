package exp

import (
	"github.com/mb0/xelf/cor"
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
		switch a.Val {
		case "void":
			return typ.Void, nil
		case "null":
			return lit.Nil, nil
		case "false":
			return lit.False, nil
		case "true":
			return lit.True, nil
		}
		return &Sym{Name: a.Val}, nil
	case '(':
		if len(a.Seq) == 0 { // empty expression is void
			return typ.Void, nil
		}
		fst, err := Parse(a.Seq[0])
		if err != nil {
			return nil, err
		}
		if fst == typ.Void { // empty expression is void
			return typ.Void, nil
		}
		args := make([]El, 0, len(a.Seq))
		args = append(args, fst)
		args, err = decledArgs(args, a.Seq[1:])
		return Dyn(args), err
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func decledArgs(res []El, seq []*lex.Tree) (_ []El, err error) {
	seq, tail := lex.SplitAfter(seq, lex.SymPred(0, func(s string) bool { return s == "-" }))
	head, rest, decls := lex.SplitKeyed(seq, false, lex.IsDecl)
	if len(rest) != 0 {
		return nil, cor.Error("unexpected decl tail")
	}
	res, err = taggedArgs(res, head)
	for _, decl := range decls {
		args, err := decledArgs(nil, decl.Seq)
		if err != nil {
			return nil, err
		}
		res = append(res, Decl{Name: decl.Key, Args: args})
	}
	res, err = taggedArgs(res, tail)
	return res, err
}

func taggedArgs(res []El, seq []*lex.Tree) (_ []El, err error) {
	seq, rest := lex.SplitAfter(seq, lex.SymPred(0, func(s string) bool { return s == "::" }))
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
	if len(tail) > 0 || len(rest) > 0 {
		args := make([]El, 0, len(tail)+len(rest))
		if len(tail) > 0 {
			args, err = plainArgs(args, tail)
			if err != nil {
				return nil, err
			}
		}
		if len(rest) > 0 {
			args, err = plainArgs(args, rest)
			if err != nil {
				return nil, err
			}
		}
		res = append(res, Tag{Name: "::", Args: args})
	}
	return res, nil
}

func plainArgs(res []El, seq []*lex.Tree) ([]El, error) {
	for _, t := range seq {
		e, err := Parse(t)
		if err != nil {
			return nil, err
		}
		if e != typ.Void {
			res = append(res, e)
		}
	}
	return res, nil
}
