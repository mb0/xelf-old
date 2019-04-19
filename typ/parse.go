package typ

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
)

var (
	ErrInvalid   = cor.StrError("invalid type")
	ErrArgCount  = cor.StrError("wrong type argument count")
	ErrRefName   = cor.StrError("expect ref name")
	ErrParamType = cor.StrError("expect param type")
)

// ParseString scans and parses string s and returns a type or an error.
func ParseString(s string) (Type, error) {
	a, err := lex.Scan(s)
	if err != nil {
		return Void, err
	}
	return Parse(a)
}

// Parse parses the element as type and returns it or an error.
func Parse(a *lex.Tree) (Type, error) {
	t, err := parse(a, nil)
	return t, err
}

// ParseSym returns the type represented by the symbol s or an error.
func ParseSym(s string, hist []Type) (res Type, _ error) {
	if len(s) == 0 {
		return Void, ErrInvalid
	}
	opt := s[len(s)-1] == '?'
	switch s[0] {
	case '~':
		ref := s[1:]
		if c := ref[0]; c >= '0' && c <= '9' {
			id, err := strconv.ParseUint(ref, 10, 64)
			if err != nil {
				return Void, cor.Errorf("invalid id: %w", err)
			}
			return Var(id), nil
		}
		i := strings.IndexByte(ref, '.')
		if i == 0 { // explicit type
			return ParseSym(ref, hist)
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
		t, err := ParseSym(s[4:], hist)
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

func parse(a *lex.Tree, hist []Type) (t Type, err error) {
	switch a.Tok {
	case lex.Symbol:
		t, err = ParseSym(a.Raw, hist)
	case '(':
		if len(a.Seq) == 0 { // empty expression is void
			return
		}
		t, err = parse(a.Seq[0], hist)
		if err != nil {
			return
		}
		t, err = ParseInfo(a.Seq[1:], t, hist)
		if err != nil {
			return
		}
	default:
		return Void, ErrInvalid
	}
	return t, err
}

func ParseInfo(args []*lex.Tree, t Type, hist []Type) (Type, error) {
	needRef, needParams := NeedsInfo(t)
	if !needRef && !needParams || len(args) == 0 {
		return t, ErrArgCount
	}
	t.Info = &Info{}
	if needRef {
		ref := args[0]
		if ref.Tok != lex.String {
			return t, ErrRefName
		}
		name, err := cor.Unquote(ref.Raw)
		if err != nil {
			return t, err
		}
		t.Ref = name
		args = args[1:]
	}
	if !needParams {
		return t, nil
	} else if len(args) == 0 {
		return t, ErrArgCount
	}
	dt, _ := t.Deopt()
	hist = append(hist, dt)
	group := dt.Kind != ExpForm
	res := make([]Param, 0, len(args))
	var naked int
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		if !isDecl(a) {
			return t, cor.Errorf("want param start got %s", a)
		}
		res = append(res, Param{Name: a.Raw[1:]})
		if group {
			naked++
		} else {
			naked = 1
		}
		if len(args) > 0 && !isDecl(args[0]) {
			t, err := parse(args[0], hist)
			args = args[1:]
			if err != nil {
				return t, err
			}
			for naked > 0 {
				res[len(res)-naked].Type = t
				naked--
			}
		}
	}
	t.Params = res
	return t, nil
}
func isDecl(a *lex.Tree) bool {
	return a.Tok == lex.Decl && strings.IndexByte("+-", a.Raw[0]) != -1
}
