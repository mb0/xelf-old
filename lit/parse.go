package lit

import (
	"errors"
	"strconv"

	"github.com/mb0/xelf/lex"
)

var (
	ErrKey     = errors.New("expect key name")
	ErrKeySep  = errors.New("expect key separator")
	ErrKeyVal  = errors.New("expect key value")
	ErrUnknown = errors.New("unknown literal")
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
	case lex.Num:
		n, err := strconv.ParseFloat(a.Val, 64)
		if err != nil {
			return nil, err
		}
		return Num(n), nil
	case lex.Str:
		txt, err := lex.Unquote(a.Val)
		if err != nil {
			return nil, err
		}
		return Char(txt), nil
	case lex.Sym:
		l, err := ParseSym(a.Val)
		if err != nil {
			return nil, a.Err(err)
		}
		return l, nil
	case '[':
		return parseList(a)
	case '{':
		return parseDict(a)
	}
	return nil, a.Err(lex.ErrUnexpected)
}

// ParseSym returns the literal represented by the symbol s or an error.
func ParseSym(s string) (Lit, error) {
	switch s {
	case "null":
		return Nil, nil
	case "false":
		return False, nil
	case "true":
		return True, nil
	}
	return nil, ErrUnknown
}

func parseList(tree *lex.Tree) (res List, _ error) {
	var last bool
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
	return res, nil
}

func parseDict(tree *lex.Tree) (*Dict, error) {
	var res []Keyed
	for i := 0; i < len(tree.Seq); i++ {
		var key string
		switch a := tree.Seq[i]; a.Tok {
		case lex.Sym:
			if len(a.Val) == 0 || !lex.IsName(a.Val) {
				return nil, a.Err(ErrKey)
			}
			key = a.Val
		case lex.Str:
			var err error
			key, err = lex.Unquote(a.Val)
			if err != nil {
				return nil, err
			}
		default:
			return nil, a.Err(ErrKey)
		}
		if i+1 >= len(tree.Seq) {
			return nil, &lex.Error{*tree.End, ErrKeySep, 0}
		}
		i++
		if b := tree.Seq[i]; b.Tok != ':' {
			return nil, b.Err(ErrKeySep)
		}
		if i+1 >= len(tree.Seq) {
			return nil, &lex.Error{*tree.End, ErrKeySep, 0}
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
	return &Dict{res}, nil
}
