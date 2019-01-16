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
	elm := elemType(res)
	switch elm.Kind & MaskRef {
	case KindObj:
		res.Info = &Info{}
	case KindRef:
		if res.Info != nil && res.Ref != "" {
			if len(args) > 0 {
				return Void, fst.Err(ErrArgCount)
			}
			return res, nil
		}
		fallthrough
	case KindFlag, KindEnum, KindRec:
		if len(args) < 1 {
			return Void, tree.Err(ErrArgCount)
		}
		res.Info = &Info{}
		res.Ref, err = parseRef(args[0])
		if err != nil {
			return Void, err
		}
		if elm.Kind != KindRec {
			if len(args) != 1 {
				return Void, tree.Err(ErrArgCount)
			}
			return res, nil
		}
		args = args[1:]
	default:
		if len(args) > 0 {
			return Void, args[0].Err(ErrArgCount)
		}
		return res, nil
	}
	res.Fields, err = parseFields(args)
	if err != nil {
		return Void, err
	}
	return res, nil
}

func parseRef(t *lex.Tree) (string, error) {
	if t.Tok != lex.Str {
		return "", t.Err(ErrRefName)
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
		return nil, head[0].Err(ErrFieldName)
	}
	if len(tail) > 0 {
		return nil, tail[0].Err(ErrFieldName)
	}
	naked := 0
	fs := make([]Field, 0, len(keyed))
	for i, n := range keyed {
		name := n.Key[1:]
		if len(n.Seq) == 0 {
			fs = append(fs, Field{Name: name})
			naked++
			continue
		}
		if name == "" {
			if naked > 0 {
				return nil, keyed[len(keyed)-naked].Tree.Err(ErrNakedField)
			}
			for _, a := range n.Seq {
				ft, err := Parse(a)
				if err != nil {
					return nil, a.Err(err)
				}
				fs = append(fs, Field{Type: ft})
			}
			continue
		}
		if len(n.Seq) > 1 {
			n.Seq[1].Err(ErrFieldType)
		}
		ft, err := Parse(n.Seq[0])
		if err != nil {
			return nil, err
		}
		for naked > 0 {
			naked--
			fs[i-naked].Type = ft
		}
		fs = append(fs, Field{Name: name, Type: ft})
	}
	if naked > 0 {
		return nil, keyed[len(keyed)-naked].Tree.Err(ErrNakedField)
	}
	return fs, nil
}

func elemType(t Type) Type {
	for k := t.Kind; ; k = k >> SlotSize {
		switch k & MaskElem {
		case KindArr, KindMap:
			continue
		}
		return Type{k, t.Info}
	}
	return t
}
