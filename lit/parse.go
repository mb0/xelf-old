package lit

import (
	"io"
	"strconv"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
)

var (
	ErrKey     = cor.StrError("expect key name")
	ErrKeySep  = cor.StrError("expect key separator")
	ErrKeyVal  = cor.StrError("expect key value")
	ErrUnknown = cor.StrError("unknown literal")
)

// Read reads and parses from r and returns a literal or an error.
func Read(r io.Reader) (Lit, error) {
	a, err := lex.Read(r)
	if err != nil {
		return nil, err
	}
	return Parse(a)
}

// Parse parses the syntax tree a and returns a literal or an error.
func Parse(a *lex.Tree) (Lit, error) {
	switch a.Tok {
	case lex.Number:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, err
		}
		return Num(n), nil
	case lex.String:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, err
		}
		return Char(txt), nil
	case lex.Symbol:
		switch a.Raw {
		case "null":
			return Nil, nil
		case "false":
			return False, nil
		case "true":
			return True, nil
		}
		return nil, a.Err(ErrUnknown)
	case '[':
		return parseList(a, &List{})
	case '{':
		return parseDict(a, &Dict{})
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func parseList(tree *lex.Tree, v *List) (*List, error) {
	v.Data = make([]Lit, 0, len(tree.Seq))
	for _, t := range tree.Seq {
		el, err := Parse(t)
		if err != nil {
			return nil, err
		}
		v.Data = append(v.Data, el)
	}
	return v, nil
}

func parseDict(tree *lex.Tree, v *Dict) (*Dict, error) {
	v.List = make([]Keyed, 0, len(tree.Seq))
	for _, t := range tree.Seq {
		if t.Tok != lex.Tag || len(t.Seq) < 2 {
			return nil, lex.ErrorAtPos(t.Pos, ErrKeySep)
		}
		a, b := t.Seq[0], t.Seq[1]
		var key string
		switch a.Tok {
		case lex.Symbol:
			key = a.Raw
		case lex.String:
			var err error
			key, err = cor.Unquote(a.Raw)
			if err != nil {
				return nil, err
			}
		default:
			return nil, a.Err(ErrKey)
		}
		el, err := Parse(b)
		if err != nil {
			return nil, err
		}
		v.List = append(v.List, Keyed{key, el})
	}
	return v, nil
}
