package exp

import (
	"io"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var ErrVoid = cor.StrError("void")

// Read scans and parses from r and returns an element or an error.
func Read(r io.Reader) (El, error) {
	a, err := lex.Read(r)
	if err != nil {
		return nil, err
	}
	return Parse(a)
}

// Parse parses the syntax tree a and returns an element or an error.
// It needs a static environment to distinguish elements.
func Parse(a *lex.Tree) (El, error) {
	switch a.Tok {
	case lex.Number, lex.String, '[', '{':
		l, err := lit.Parse(a)
		if err != nil {
			return nil, err
		}
		return &Atom{Lit: l, Src: a.Src}, nil
	case lex.Symbol:
		switch a.Raw {
		case "null":
			return &Atom{Lit: lit.Nil, Src: a.Src}, nil
		case "false":
			return &Atom{Lit: lit.False, Src: a.Src}, nil
		case "true":
			return &Atom{Lit: lit.True, Src: a.Src}, nil
		}
		switch a.Raw[0] {
		case '~', '@':
			t, err := typ.Parse(a)
			if err != nil {
				return nil, err
			}
			return &Atom{Lit: t, Src: a.Src}, nil
		}
		t, err := typ.Parse(a)
		if err == nil && t.Kind.Prom() {
			return &Atom{Lit: t, Src: a.Src}, nil
		}
		return &Sym{Name: a.Raw, Src: a.Src}, nil
	case ':', ';':
		return &Sym{Name: string(a.Tok), Src: a.Src}, nil
	case lex.Tag:
		if len(a.Seq) == 0 {
			return nil, cor.Errorf("invalid tag %q", a.String())
		}
		fst := a.Seq[0]
		res := &Tag{Src: a.Src}
		switch a.Raw {
		case ":":
			if len(a.Seq) > 2 {
				break
			}
			switch fst.Tok {
			case lex.Symbol:
				res.Name = fst.Raw
			case lex.String:
				name, err := cor.Unquote(fst.Raw)
				if err != nil {
					return nil, err
				}
				res.Name = name
			}
			if len(a.Seq) > 1 {
				snd, err := Parse(a.Seq[1])
				if err != nil {
					return nil, err
				}
				res.El = snd
			}
			return res, nil
		case ";":
			if len(a.Seq) > 1 || fst.Tok != lex.Symbol {
				break
			}
			res.Name = cor.LastName(fst.Raw)
			return res, nil
		}
		return nil, cor.Errorf("invalid tag %q", a.String())
	case '<':
		t, err := typ.Parse(a)
		if err != nil {
			return nil, err
		}
		return &Atom{Lit: t, Src: a.Src}, nil
	case '(':
		// TODO move comment and named handling to resl
		if len(a.Seq) == 0 { // empty expression is void
			return nil, ErrVoid
		}
		ftok := a.Seq[0]
		switch ftok.Tok {
		case lex.Symbol:
			if ftok.Raw == "void" {
				return nil, ErrVoid
			}
		}
		fst, err := Parse(ftok)
		if err != nil || fst == nil {
			return nil, err
		}
		d, err := parseDyn(a.Seq[1:], fst)
		if err != nil {
			return nil, err
		}
		d.Src = a.Src
		return d, nil
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func parseDyn(seq []*lex.Tree, el El) (_ *Dyn, err error) {
	args, src, err := parseArgs(seq, el)
	if err != nil {
		return nil, err
	}
	return &Dyn{Els: args, Src: src}, nil
}

func parseArgs(seq []*lex.Tree, el El) (args []El, src lex.Src, err error) {
	args = make([]El, 0, len(seq)+1)
	if el != nil {
		args = append(args, el)
		src.Pos = el.Source().Pos
	}
	for i, t := range seq {
		if i == 0 && el == nil {
			src.Pos = t.Pos
		}
		src.End = t.End
		el, err = Parse(t)
		if err == ErrVoid {
			continue
		}
		if err != nil {
			return nil, src, err
		}
		args = append(args, el)
	}
	return args, src, nil
}
