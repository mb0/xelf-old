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
// It needs a static environment to distinguish elements.
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
		def := Lookup(env, a.Raw)
		if def == nil {
			t, err := typ.Parse(a)
			if err == nil {
				return t, nil
			}
		}
		return &Sym{Name: a.Raw, Pos: a.Pos, Def: def}, nil
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
			els, err := parseArgs(env, a.Seq[1:], nil)
			if err != nil {
				return nil, err
			}
			t.El = Dyn(els)
			return t, nil
		case *Sym:
			if t.Def != nil && t.Def.Spec != nil {
				els, err := parseArgs(env, a.Seq[1:], nil)
				if err != nil {
					return nil, err
				}
				return &Call{Def: t.Def, Args: els, Src: a.Src}, nil
			}
		}
		return parseArgs(env, a.Seq[1:], fst)
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func parseArgs(env Env, seq []*lex.Tree, el El) (args Dyn, err error) {
	args = make(Dyn, 0, len(seq)+1)
	if el != nil {
		args = append(args, el)
	}
	for _, t := range seq {
		el, err = Parse(env, t)
		if err != nil {
			return nil, err
		}
		if el != typ.Void {
			args = append(args, el)
		}
	}
	return args, nil
}
