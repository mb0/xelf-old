package exp

import (
	"github.com/mb0/xelf/cor"
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
		if def != nil {
			return &Sym{Name: a.Raw, Src: a.Src, Type: def.Type, Lit: def.Lit}, nil
		}
		t, err := typ.Parse(a)
		if err == nil && t.Kind.Prom() {
			return &Atom{Lit: t, Src: a.Src}, nil
		}
		return &Sym{Name: a.Raw, Src: a.Src}, nil
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
			tt := t.Lit.(typ.Type)
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
			if t.Name[0] == ':' {
				def := Lookup(env, t.Name)
				if def == nil {
					return nil, errStartingTag(t.Name)
				}
				spec, ok := def.Lit.(*Spec)
				if !ok {
					return nil, errStartingTag(t.Name)
				}
				return &Call{Spec: spec, Args: t.Args(), Src: t.Src}, nil
			}
			dyn, err := parseDyn(env, a.Seq[1:], nil)
			if err != nil {
				return nil, err
			}
			t.El = dyn
			t.Src = a.Src
			return t, nil
		case *Sym:
			if spec, ok := t.Lit.(*Spec); ok {
				els, _, err := parseArgs(env, a.Seq[1:], nil)
				if err != nil {
					return nil, err
				}
				return &Call{Spec: spec, Args: els, Src: a.Src}, nil
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

func errStartingTag(name string) error {
	return cor.Errorf("expressions starting with a tag must resolve to a built-in type "+
		"conversion spec, got %v", name)
}

func parseDyn(env Env, seq []*lex.Tree, el El) (_ *Dyn, err error) {
	args, src, err := parseArgs(env, seq, el)
	if err != nil {
		return nil, err
	}
	return &Dyn{Els: args, Src: src}, nil
}

func parseArgs(env Env, seq []*lex.Tree, el El) (args []El, src lex.Src, err error) {
	args = make([]El, 0, len(seq)+1)
	if el != nil {
		args = append(args, el)
		src.Pos = el.Source().Pos
	}
	var tag *Named
	for i, t := range seq {
		if i == 0 && el == nil {
			src.Pos = t.Pos
		}
		src.End = t.End
		el, err = Parse(env, t)
		if err != nil {
			return nil, src, err
		}
		switch v := el.(type) {
		case nil:
			continue
		case *Named:
			if tag != nil {
				args = append(args, tag)
				tag = nil
			}
			if v.IsTag() && v.El == nil {
				tag = v
				continue
			}
		}
		if tag != nil {
			tag.El = el
			el, tag = tag, nil
		}
		args = append(args, el)
	}
	if tag != nil {
		args = append(args, tag)
	}
	return args, src, nil
}
