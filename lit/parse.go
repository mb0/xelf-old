package lit

import (
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

// ParseString scans and parses string s and returns a literal or an error.
func ParseString(s string) (Lit, error) {
	a, err := lex.Scan(s)
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
		return parseList(a)
	case '{':
		return parseDict(a)
	}
	return nil, a.Err(lex.ErrUnexpected)
}

func parseList(tree *lex.Tree) (*List, error) {
	var last bool
	res := make([]Lit, 0, len(tree.Seq))
	for _, t := range tree.Seq {
		if last && t.Tok == ',' {
			last = false
			continue
		}
		last = true
		el, err := Parse(t)
		if err != nil {
			return nil, err
		}
		res = append(res, el)
	}
	return &List{Data: res}, nil
}

func parseDict(tree *lex.Tree) (*Dict, error) {
	var res []Keyed
	for i := 0; i < len(tree.Seq); i++ {
		var key string
		switch a := tree.Seq[i]; a.Tok {
		case lex.Symbol:
			if len(a.Raw) == 0 || !cor.IsName(a.Raw) {
				return nil, a.Err(ErrKey)
			}
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
		if i+1 >= len(tree.Seq) {
			return nil, lex.ErrorAtPos(tree.End, ErrKeySep)
		}
		i++
		if b := tree.Seq[i]; b.Tok != ':' {
			return nil, b.Err(ErrKeySep)
		}
		if i+1 >= len(tree.Seq) {
			return nil, lex.ErrorAtPos(tree.End, ErrKeySep)
		}
		i++
		el, err := Parse(tree.Seq[i])
		if err != nil {
			return nil, err
		}
		res = append(res, Keyed{key, el})
		if i+1 < len(tree.Seq) && tree.Seq[i+1].Tok == ',' {
			i++
		}
	}
	return &Dict{List: res}, nil
}
