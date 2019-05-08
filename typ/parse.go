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
func ParseSym(s string, hist []Type) (res Type, err error) {
	if len(s) == 0 {
		return Void, ErrInvalid
	}
	opt := s[len(s)-1] == '?'
	switch s[0] {
	case '~':
		ref := s[1:]
		if opt {
			ref = ref[:len(ref)-1]
		}
		if len(ref) > 0 && cor.Digit(rune(ref[0])) {
			// self reference by index
			idx, err := strconv.Atoi(ref)
			if err != nil {
				return Void, cor.Errorf("self ref index must be a number: %w", err)
			}
			if idx < 0 || idx >= len(hist) {
				return Void, cor.Error("self ref index out of bounds")
			}
			res = hist[len(hist)-1-idx]
		} else if strings.IndexByte(ref, '.') == -1 { // explicit type
			k, err := ParseKind(s)
			return Type{Kind: k}, err
		} else { // schema type
			res = Sch(ref)
		}
		if opt {
			return Opt(res), nil
		}
		return res, nil
	case '@':
		ref := s[1:]
		if opt {
			ref = ref[:len(ref)-1]
		}
		var cs []Type
		if i := strings.IndexByte(ref, ':'); i >= 0 {
			c, err := ParseSym(ref[i+1:], nil)
			if err != nil {
				return Void, cor.Errorf("invalid type var constraint %s", s)
			}
			cs = []Type{c}
			ref = ref[:i]
		}
		if len(ref) == 0 {
			res = Var(0, cs...)
		} else if cor.Digit(rune(ref[0])) {
			id, err := strconv.ParseUint(ref, 10, 64)
			if err != nil {
				return Void, cor.Errorf("invalid type var %s", s)
			}
			res = Var(id, cs...)
		} else {
			res = Ref(ref)
		}
		if opt {
			res = Opt(res)
		}
		return res, nil
	}
	if len(s) == 4 || len(s) > 4 && s[4] == '|' {
		t := Any
		if len(s) > 4 {
			t, err = ParseSym(s[5:], hist)
			if err != nil {
				return Void, err
			}
		}
		switch s[:4] {
		case "cont":
			return Cont(t), nil
		case "idxr":
			return Idxr(t), nil
		case "keyr":
			return Keyr(t), nil
		case "list":
			return List(t), nil
		case "dict":
			return Dict(t), nil
		}
	}
	k, err := ParseKind(s)
	if err != nil {
		return Void, err
	}
	t := Type{Kind: k}
	switch t.Kind & MaskRef {
	case KindBits, KindEnum, KindObj, KindRec, KindAlt, KindFunc, KindForm:
		t.Info = &Info{}
	}
	return Type{Kind: k}, nil
}

// NeedsInfo returns whether type t is missing reference or params information.
func NeedsInfo(t Type) (ref, params bool) {
	switch t = t.Last(); t.Kind & MaskRef {
	case KindBits, KindEnum, KindObj:
		return !t.HasRef(), false
	case KindRec, KindAlt:
		return false, !t.HasParams()
	case KindForm:
		return !t.HasRef(), !t.HasParams()
	case KindFunc:
		return false, !t.HasParams()
	case KindVar:
		if t.HasParams() {
			return NeedsInfo(t.Params[0].Type)
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
		return t, cor.Errorf("for %s %d: %v", t, len(args), ErrArgCount)
	}
	if t.Info == nil {
		t.Info = &Info{}
	}
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
		return t, cor.Errorf("for %s: %v", t, ErrArgCount)
	}
	dt, _ := t.Deopt()
	hist = append(hist, dt)
	group := dt.Kind != KindForm
	res := make([]Param, 0, len(args))
	var naked int
	for len(args) > 0 {
		a := args[0]
		var name string
		if a.Tok == lex.Tag {
			name = a.Raw[1:]
			args = args[1:]
		} // else unnamed parameter
		res = append(res, Param{Name: name})
		if group {
			naked++
		} else {
			naked = 1
		}
		if len(args) > 0 && args[0].Tok != lex.Tag {
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
	if t.Kind == KindList || t.Kind == KindDict {
		e := t.Elem()
		for e.Kind == KindList || e.Kind == KindDict {
			t = e
			e = t.Elem()
		}
		if e.Info != nil {
			e.Params = res
		} else {
			e.Info = &Info{Params: res}
			t.Params[0].Type = e
		}
	} else {
		t.Params = res
	}
	return t, nil
}
