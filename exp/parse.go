package exp

import (
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ParseString scans and parses string s and returns an element or an error.
func ParseString(env Env, s string) (El, error) {
	a, err := lex.Scan(s)
	if err != nil {
		return nil, err
	}
	return Parse(env, a)
}

// Parse parses the syntax tree a and returns an element or an error.
func Parse(env Env, a *lex.Tree) (El, error) {
	switch a.Tok {
	case lex.Number, lex.String, '[', '{':
		return lit.Parse(a)
	case lex.Symbol:
		switch a.Raw[0] {
		case '~', '@':
			return typ.Parse(a)
		case '.', '$', '/': // paths
			return &Sym{Name: a.Raw, Pos: a.Pos}, nil
		}
		switch a.Raw {
		case "void":
			return typ.Void, nil
		case "null":
			return lit.Nil, nil
		case "false":
			return lit.False, nil
		case "true":
			return lit.True, nil
		}
		resl := Lookup(env, a.Raw)
		if resl == nil {
			t, err := typ.Parse(a)
			if err == nil {
				return t, nil
			}
		}
		return &Sym{Name: a.Raw, Pos: a.Pos}, nil
	case lex.Tag, lex.Decl:
		return &Named{Name: a.Raw, Pos: a.Pos}, nil
	case '(':
		if len(a.Seq) == 0 { // empty expression is void
			return typ.Void, nil
		}
		fst, err := Parse(env, a.Seq[0])
		if err != nil {
			return nil, err
		}
		switch t := fst.(type) {
		case Type:
			if t == typ.Void {
				return t, nil
			}
			r, p := typ.NeedsInfo(t)
			if r || p {
				t, err = typ.ParseInfo(a.Seq[1:], t, nil)
				if err != nil {
					return nil, err
				}
				return t, nil
			}
		case *Named:
			els := make([]El, 0, len(a.Seq)-1)
			for _, b := range a.Seq[1:] {
				el, err := Parse(env, b)
				if err != nil {
					return nil, err
				}
				els = append(els, el)
			}
			t.El = Dyn(els)
			return t, nil
		}
		els := make([]El, 1, len(a.Seq))
		els[0] = fst
		for _, b := range a.Seq[1:] {
			el, err := Parse(env, b)
			if err != nil {
				return nil, err
			}
			if el != typ.Void {
				els = append(els, el)
			}
		}
		return Dyn(els), nil
	}
	return nil, a.Err(lex.ErrUnexpected)
}
