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

func ParseString(s string) (Type, error) {
	a, err := lex.Scan(s)
	if err != nil {
		return Void, err
	}
	return Parse(a)
}

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

func ParseSym(str string) (Type, error) {
	if len(str) == 0 {
		return Void, ErrInvalid
	}
	if str[0] == '@' {
		if str[len(str)-1] == '?' {
			return Opt(Ref(str[1 : len(str)-1])), nil
		}
		return Ref(str[1:]), nil
	}
	if len(str) > 4 && str[3] == '|' {
		t, err := ParseSym(str[4:])
		switch str[:3] {
		case "arr":
			return Arr(t), err
		case "map":
			return Map(t), err
		}
	}
	k, err := ParseKind(str)
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

func ParseInfo(t *lex.Tree, ref, fields bool) (n *Info, err error) {
	if !(ref || fields) {
		return nil, nil
	}
	args := t.Seq[1:]
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
