package typ

import (
	"strconv"
	"strings"

	"github.com/mb0/xelf/cor"
)

var (
	ErrInvalid   = cor.StrError("invalid type")
	ErrArgCount  = cor.StrError("wrong type argument count")
	ErrRefName   = cor.StrError("expect ref name")
	ErrParamType = cor.StrError("expect param type")
)

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
