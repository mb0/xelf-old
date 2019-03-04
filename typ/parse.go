package typ

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
)

var (
	ErrInvalid    = cor.StrError("invalid type")
	ErrArgCount   = cor.StrError("wrong type argument count")
	ErrRefName    = cor.StrError("expect ref name")
	ErrParamName  = cor.StrError("expect param name")
	ErrParamType  = cor.StrError("expect param type")
	ErrNakedParam = cor.StrError("naked param declaration")
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
	return parse(a, nil)
}
func parse(a *lex.Tree, hist []Type) (Type, error) {
	if len(a.Seq) > 0 && a.Tok == '(' {
		return parseSeq(a, hist)
	}
	if a.Tok == lex.Sym {
		t, err := parseSym(a.Val, hist)
		if err != nil {
			return Void, a.Err(err)
		}
		return t, nil
	}
	return Void, a.Err(ErrInvalid)
}

// ParseSym returns the type represented by the symbol s or an error.
func ParseSym(s string) (Type, error) {
	return parseSym(s, nil)
}

func parseSym(s string, hist []Type) (res Type, _ error) {
	if len(s) == 0 {
		return Void, ErrInvalid
	}
	opt := s[len(s)-1] == '?'
	switch s[0] {
	case '~':
		ref := s[1:]
		i := strings.IndexByte(ref, '.')
		if i == 0 { // explicit type
			return parseSym(ref, hist)
		} // else schema reference
		if opt {
			return Opt(Ref(ref[:len(ref)-1])), nil
		}
		return Ref(ref), nil
	case '@':
		ref := s[1:]
		if opt {
			ref = ref[:len(ref)-1]
		}
		res = Ref(ref)
		if len(ref) > 0 {
			if c := ref[0]; c >= '0' && c <= '9' { // self reference by index
				idx, err := strconv.Atoi(ref)
				if err != nil {
					return Void, cor.Errorf("self ref index must be a number: %w", err)
				}
				if idx < 0 || idx >= len(hist) {
					return Void, cor.Error("self ref index out of bounds")
				}
				res = hist[len(hist)-1-idx]
			}
		}
		if opt {
			res = Opt(res)
		}
		return res, nil

	}
	if len(s) > 4 && s[3] == '|' {
		t, err := parseSym(s[4:], hist)
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

func parseSeq(tree *lex.Tree, hist []Type) (Type, error) {
	fst, args := tree.Seq[0], tree.Seq[1:]
	if fst.Tok != lex.Sym {
		return Void, fst.Err(ErrInvalid)
	}
	res, err := parseSym(fst.Val, hist)
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
	res, err = parseInfo(res, tree, needRef, needFields, hist)
	if err != nil {
		return Void, err
	}
	return res, nil
}

// NeedsInfo returns whether type t is missing reference or params information.
func NeedsInfo(t Type) (ref, params bool) {
	switch k := t.Last().Kind & MaskRef; k {
	case KindFlag, KindEnum, KindRec:
		ref = t.Info == nil || len(t.Ref) == 0
		return ref, false
	case KindObj:
		return false, t.Info == nil || len(t.Params) == 0
	case KindExp:
		switch t.Kind {
		case ExpForm:
			return t.Info == nil || len(t.Ref) == 0,
				t.Info == nil || len(t.Params) == 0
		case ExpFunc:
			return false, t.Info == nil || len(t.Params) == 0
		}
	}
	return false, false
}

// ParseInfo parses arguments of a for ref and params information and returns it or an error.
func ParseInfo(t Type, a *lex.Tree, ref, params bool) (Type, error) {
	return parseInfo(t, a, ref, params, nil)
}
func parseInfo(t Type, a *lex.Tree, ref, params bool, hist []Type) (_ Type, err error) {
	if a == nil || !(ref || params) {
		return Void, nil
	}
	if len(a.Seq) < 2 {
		return Void, a.Err(ErrArgCount)
	}
	args := a.Seq[1:]
	t.Info = &Info{}
	if ref {
		t.Ref, err = parseRef(args[0])
		if err != nil {
			return Void, args[0].Err(err)
		}
		args = args[1:]
	}
	if params {
		dt, _ := t.Deopt()
		t.Params, err = parseParams(args, append(hist, dt))
		if err != nil {
			return Void, a.Seq[0].Err(err)
		}
	}
	return t, nil
}

func parseRef(t *lex.Tree) (string, error) {
	if t.Tok != lex.Str {
		return "", ErrRefName
	}
	return lex.Unquote(t.Val)
}

func isParamDecl(s string) bool { return s != "" && s[0] == '+' }

func parseParams(seq []*lex.Tree, hist []Type) ([]Param, error) {
	if len(seq) == 0 {
		return nil, ErrArgCount
	}
	head, tail, keyed := lex.SplitKeyed(seq, true, isParamDecl)
	if len(head) > 0 {
		return nil, ErrParamName
	}
	if len(tail) > 0 {
		return nil, ErrParamName
	}
	naked := 0
	ps := make([]Param, 0, len(keyed))
	for _, n := range keyed {
		name := n.Key[1:]
		if len(n.Seq) == 0 {
			ps = append(ps, Param{Name: name})
			naked++
			continue
		}
		if name == "" && naked == 0 {
			for _, a := range n.Seq {
				ft, err := parse(a, hist)
				if err != nil {
					return nil, err
				}
				ps = append(ps, Param{Type: ft})
			}
			continue
		}
		if len(n.Seq) > 1 {
			return nil, ErrParamType
		}
		ft, err := parse(n.Seq[0], hist)
		if err != nil {
			return nil, err
		}
		for naked > 0 {
			ps[len(ps)-naked].Type = ft
			naked--
		}
		ps = append(ps, Param{Name: name, Type: ft})
	}
	if naked > 0 {
		return nil, ErrNakedParam
	}
	return ps, nil
}
