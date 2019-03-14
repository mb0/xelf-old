package exp

import (
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
		st := a.Seq[0]
		if st.Tok == lex.Sym && st.Val != "" {
			switch st.Val[0] {
			case '-', '+':
				args, err := decledArgs(nil, a.Seq[1:])
				return Decl{st.Val, args}, err
			case ':':
				args, err := plainArgs(nil, a.Seq[1:])
				return Tag{st.Val, args}, err
			}
		}
		fst, err := Parse(st)
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
	if res == nil {
		res = make([]El, 0, len(seq))
	}
	var decl bool
	var last bool
	var lasti int
	for i, t := range seq {
		if key, ok := lex.CheckSym(t, 1, lex.IsDecl); ok {
			if !decl && i > 0 {
				res, err = taggedArgs(res, seq[:i])
				if err != nil {
					return nil, err
				}
			} else if last {
				tags, err := taggedArgs(nil, seq[lasti:i])
				if err != nil {
					return nil, err
				}
				lt := res[len(res)-1].(Decl)
				lt.Args = append(lt.Args, tags...)
				res[len(res)-1] = lt
			}
			decl = true
			if lex.IsExp(t) {
				args, err := decledArgs(nil, t.Seq[1:])
				if err != nil {
					return nil, err
				}
				res = append(res, Decl{key, args})
				last = false
			} else {
				res = append(res, Decl{key, nil})
				last = true
				lasti = i + 1
			}
		}
	}
	if !decl {
		res, err = taggedArgs(res, seq)
		if err != nil {
			return nil, err
		}
	} else if last && lasti < len(seq) {
		tags, err := taggedArgs(nil, seq[lasti:])
		if err != nil {
			return nil, err
		}
		lt := res[len(res)-1].(Decl)
		lt.Args = append(lt.Args, tags...)
		res[len(res)-1] = lt
	}
	return res, err
}

func taggedArgs(res []El, seq []*lex.Tree) (_ []El, err error) {
	if res == nil {
		res = make([]El, 0, len(seq))
	}
	var tag bool
	var last bool
	for i, t := range seq {
		if key, ok := lex.CheckSym(t, 1, lex.IsTag); ok {
			tag = true
			if lex.IsExp(t) {
				args, err := decledArgs(nil, t.Seq[1:])
				if err != nil {
					return nil, err
				}
				res = append(res, Tag{key, args})
				last = false
			} else {
				res = append(res, Tag{key, nil})
				last = key != "::"
			}
		} else if last {
			args, err := plainArgs(nil, seq[i:i+1])
			if err != nil {
				return nil, err
			}
			lt := res[len(res)-1].(Tag)
			lt.Args = args
			res[len(res)-1] = lt
			last = false
		} else if tag {
			return decledArgs(res, seq[i:])
		} else {
			e, err := Parse(t)
			if err != nil {
				return nil, err
			}
			if e != typ.Void {
				res = append(res, e)
			}
		}
	}
	return res, nil
}

func plainArgs(res []El, seq []*lex.Tree) ([]El, error) {
	if res == nil {
		res = make([]El, 0, len(seq))
	}
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
