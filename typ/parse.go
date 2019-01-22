package typ

import (
	"errors"

	"github.com/mb0/xelf/lex"
)

var (
	ErrInvalid    = errors.New("invalid type")
	ErrArgCount   = errors.New("wrong type argument count")
	ErrRefName    = errors.New("expect ref name")
	ErrFieldName  = errors.New("expect field name")
	ErrFieldType  = errors.New("expect field type")
	ErrNakedField = errors.New("naked field declaration")
)

// ParseString scans and parses string s and returns a type or an error.
func ParseString(s string) (Type, error) {
	a, err := lex.Scan(s)
	if err != nil {
		return Void, err
	}
	return Parse(a)
}

// Parse parses the syntax tree a and returns a type or an error.
func Parse(a *lex.Tree) (Type, error) {
	if len(a.Seq) > 0 && a.Tok == '(' {
		return parseSeq(a)
	}
	if a.Tok == lex.Sym {
		t, err := ParseSym(a.Val)
		if err != nil {
			return Void, a.Err(err)
		}
		return t, nil
	}
	return Void, a.Err(ErrInvalid)
}

// ParseSym returns the type represented by the symbol s or an error.
func ParseSym(s string) (Type, error) {
	if len(s) == 0 {
		return Void, ErrInvalid
	}
	if s[0] == '@' {
		if s[len(s)-1] == '?' {
			return Opt(Ref(s[1 : len(s)-1])), nil
		}
		return Ref(s[1:]), nil
	}
	if len(s) > 4 && s[3] == '|' {
		t, err := ParseSym(s[4:])
		switch s[:3] {
		case "arr":
			return Arr(t), err
		case "map":
			return Map(t), err
		}
	}
	k, err := ParseKind(s)
	return Type{Kind: k}, err
}

func parseSeq(tree *lex.Tree) (Type, error) {
	fst, args := tree.Seq[0], tree.Seq[1:]
	if fst.Tok != lex.Sym {
		return Void, fst.Err(ErrInvalid)
	}
	res, err := ParseSym(fst.Val)
	if err != nil {
		return Void, fst.Err(err)
	}
	needRef, needFields := NeedsInfo(res)
	if !needRef && !needFields {
		if len(args) > 0 {
			return Void, args[0].Err(ErrArgCount)
		}
		return res, nil
	}
	res.Info, err = ParseInfo(tree, needRef, needFields)
	if err != nil {
		return Void, err
	}
	return res, nil
}

// NeedsInfo returns whether type t is missing reference or field information.
func NeedsInfo(t Type) (ref, fields bool) {
	switch t.Last().Kind & MaskRef {
	case KindFlag, KindEnum:
		return t.Info == nil || len(t.Ref) == 0, false
	case KindObj:
		return false, t.Info == nil || len(t.Fields) == 0
	case KindRec:
		return t.Info == nil || len(t.Ref) == 0, t.Info == nil || len(t.Fields) == 0
	}
	return false, false
}

// ParseInfo parses arguments of a for ref and field information and returns it or an error.
func ParseInfo(a *lex.Tree, ref, fields bool) (n *Info, err error) {
	if !(ref || fields) {
		return nil, nil
	}
	args := a.Seq[1:]
	n = &Info{}
	if ref {
		if len(args) < 1 {
			return nil, ErrArgCount
		}
		n.Ref, err = parseRef(args[0])
		if err != nil {
			return nil, args[0].Err(err)
		}
		args = args[1:]
	}
	if fields {
		n.Fields, err = parseFields(args)
		if err != nil {
			return nil, args[0].Err(err)
		}
	}
	return n, nil
}

func parseRef(t *lex.Tree) (string, error) {
	if t.Tok != lex.Str {
		return "", ErrRefName
	}
	return lex.Unquote(t.Val)
}

func isFieldDecl(s string) bool { return s != "" && s[0] == '+' }

func parseFields(seq []*lex.Tree) ([]Field, error) {
	if len(seq) == 0 {
		return nil, ErrArgCount
	}
	head, tail, keyed := lex.SplitKeyed(seq, true, isFieldDecl)
	if len(head) > 0 {
		return nil, ErrFieldName
	}
	if len(tail) > 0 {
		return nil, ErrFieldName
	}
	naked := 0
	fs := make([]Field, 0, len(keyed))
	for _, n := range keyed {
		name := n.Key[1:]
		if len(n.Seq) == 0 {
			fs = append(fs, Field{Name: name})
			naked++
			continue
		}
		if name == "" {
			if naked > 0 {
				return nil, ErrNakedField
			}
			for _, a := range n.Seq {
				ft, err := Parse(a)
				if err != nil {
					return nil, err
				}
				fs = append(fs, Field{Type: ft})
			}
			continue
		}
		if len(n.Seq) > 1 {
			return nil, ErrFieldType
		}
		ft, err := Parse(n.Seq[0])
		if err != nil {
			return nil, err
		}
		for naked > 0 {
			fs[len(fs)-naked].Type = ft
			naked--
		}
		fs = append(fs, Field{Name: name, Type: ft})
	}
	if naked > 0 {
		return nil, ErrNakedField
	}
	return fs, nil
}
