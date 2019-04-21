package exp

import (
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ParseString scans and parses string s and returns an element or an error.
func ParseString(env Env, s string) (Expr, error) {
	a, err := lex.Scan(s)
	if err != nil {
		return nil, err
	}
	return Parse(env, a)
}

// Parse parses the syntax tree a and returns an element or an error.
// It needs a static environment to distinguish elements.
func Parse(env Env, a *lex.Tree) (Expr, error) {
	switch a.Tok {
	case lex.Number, lex.String, '[', '{':
		l, err := lit.Parse(a)
		if err != nil {
			return nil, err
		}
		return &Atom{Lit: l, Src: a.Src}, nil
	case lex.Symbol:
		switch a.Raw[0] {
		case '~', '@':
			t, err := typ.Parse(a)
			if err != nil {
				return nil, err
			}
			return &Atom{Lit: t, Src: a.Src}, nil
		case '.', '$', '/': // paths
			return &Sym{Name: a.Raw, Src: a.Src}, nil
		}
		switch a.Raw {
		case "void":
			return &Atom{Lit: typ.Void, Src: a.Src}, nil
		case "null":
			return &Atom{Lit: lit.Nil, Src: a.Src}, nil
		case "false":
			return &Atom{Lit: lit.False, Src: a.Src}, nil
		case "true":
			return &Atom{Lit: lit.True, Src: a.Src}, nil
		}
		def := Lookup(env, a.Raw)
		if def == nil {
			t, err := typ.Parse(a)
			if err == nil {
				return &Atom{Lit: t, Src: a.Src}, nil
			}
		}
		return &Sym{Name: a.Raw, Src: a.Src, Def: def}, nil
	case lex.Tag, lex.Decl:
		return &Named{Name: a.Raw, Src: a.Src}, nil
	case '(':
		if len(a.Seq) == 0 { // empty expression is void
			return nil, nil
		}
		fst, err := Parse(env, a.Seq[0])
		if err != nil || fst == nil {
			return nil, err
		}
		switch t := fst.(type) {
		case *Atom:
			if t.Typ() != typ.Typ {
				break
			}
			tt := t.Lit.(Type)
			if tt == typ.Void {
				return nil, nil
			}
			r, p := typ.NeedsInfo(tt)
			if r || p {
				tt, err = typ.ParseInfo(a.Seq[1:], tt, nil)
				if err != nil {
					return nil, err
				}
				return &Atom{tt, a.Src}, nil
			}
		case *Named:
			dyn, err := parseDyn(env, a.Seq[1:], nil)
			if err != nil {
				return nil, err
			}
			t.El = dyn
			t.Src = a.Src
			return t, nil
		case *Sym:
			if t.Def != nil && t.Def.Spec != nil {
				els, _, err := parseArgs(env, a.Seq[1:], nil)
				if err != nil {
					return nil, err
				}
				return &Call{Def: t.Def, Args: els, Src: a.Src}, nil
			}
		}
		d, err := parseDyn(env, a.Seq[1:], fst)
		if err != nil {
			return nil, err
		}
		d.Src = a.Src
		return d, nil
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func parseDyn(env Env, seq []*lex.Tree, el Expr) (_ *Dyn, err error) {
	args, src, err := parseArgs(env, seq, el)
	if err != nil {
		return nil, err
	}
	return &Dyn{Els: args, Src: src}, nil
}

func parseArgs(env Env, seq []*lex.Tree, el Expr) (args []El, src lex.Src, err error) {
	args = make([]El, 0, len(seq)+1)
	if el != nil {
		args = append(args, el)
		src.Pos = el.Source().Pos
	}
	for i, t := range seq {
		if i == 0 && el == nil {
			src.Pos = t.Pos
		}
		el, err = Parse(env, t)
		if err != nil {
			return nil, src, err
		}
		if el != nil {
			args = append(args, el)
			src.End = t.End
		}
	}
	return args, src, nil
}
