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
		dyn := make(Dyn, 0, len(a.Seq))
		for i, t := range a.Seq {
			e, err := Parse(t)
			if err != nil {
				return nil, err
			}
			if i == 0 && e == typ.Void { // empty expression is void
				return typ.Void, nil
			}
			if e != typ.Void {
				dyn = append(dyn, e)
			}
		}
		return dyn, nil
	}
	return nil, a.Err(lex.ErrUnexpected)
}
