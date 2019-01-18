package typ

// Cmp is a bit-set for detailed type comparisons results
type Cmp uint64

// Exclusive level bits define the meaning of the other bits. Levels are check, conv, wrap and equal.
// These level bits are the most significant in the set and allow simple compatibility tests.
//     res < LvlConv
const (
	LvlCheck Cmp = 1 << (iota + 24)
	LvlConv
	LvlEqual
	LvlMask Cmp = 0x7 << 24
	CmpNone Cmp = 0
)

const (
	// wrap src as some opt
	BitWrap Cmp = 1 << (iota + 20)
	// unwrap src opt
	BitUnwrap
)

const (
	// both types are the same
	CmpSame = LvlEqual | (1 << iota)
	// infer type from src, dst is a unnamed ref
	CmpInfer
)
const (
	// convert src to the any type
	CmpConvAny = LvlConv | (1 << iota)
	// convert src from spec to base type
	CmpConvBase
	// convert src from base to spec type
	CmpConvSpec
	// convert idxer to list
	CmpConvList
	// convert keyer to dict
	CmpConvDict
	// convert from arr to another arr
	CmpConvArr
	// convert from map to another map
	CmpConvMap
	// convert from obj to another obj
	CmpConvObj
)

const (
	// resolve unresolved type references in either src or dst
	CmpCheckRef = LvlCheck | (1 << iota)
	// compare src any value type to dst
	CmpCheckAny
	// parse base to spec type
	CmpCheckSpec
	// compare list element value types to dst arr or obj
	CmpCheckList
	// compare dict element value types to dst map or obj
	CmpCheckDict
	// try to convert arr to another arr
	CmpCheckArr
	// try to convert map to another map
	CmpCheckMap
	// try to convert obj to another obj
	CmpCheckObj
)

// Compare returns compatibility details for a source and destination type
// When the result is not none or equal, it informs what to do with a value
// of the source type to arrive at the destination type.
// The flag, enum and rec are treated as their corresponding literal type
func Compare(src, dst Type) Cmp {
	s, so := deopt(src)
	d, do := deopt(dst)
	res := compare(s, d)
	if res > LvlCheck && so != do {
		if so {
			res |= BitUnwrap
		} else {
			res |= BitWrap
		}
	}
	return res
}

func deopt(t Type) (_ Type, ok bool) {
	if ok = isOpt(t); ok {
		t.Kind = t.Kind &^ FlagOpt
	}
	return t, ok
}
func isOpt(t Type) bool {
	return t.Kind&FlagOpt != 0 && t.Kind&MaskRef != 0
}

func compare(src, dst Type) Cmp {
	s, d := src.Kind, dst.Kind
	if s == d {
		if src.Info.Equal(dst.Info) {
			return CmpSame
		}
	}
	if d == KindRef {
		if dst.Info == nil || dst.Info.Ref == "" {
			// infer dst type from src
			return CmpInfer
		}
		// ref needs to be resolved first
		return CmpCheckRef
	}
	if s == KindRef {
		return CmpCheckRef
	}
	// rule out special types, which have strict equality
	if m := Kind(MaskBase | FlagOpt); s&m == 0 || d&m == 0 {
		return CmpNone
	}
	// handle any, type
	if d == KindAny {
		return CmpConvAny
	}
	if s == KindAny {
		return CmpCheckAny
	}
	// handle base types starting with primitives
	if d == BaseNum || d == BaseChar {
		if s&d == 0 {
			return CmpNone
		}
		return CmpConvBase
	}
	if s == BaseNum || s == BaseChar {
		if d&s == 0 {
			return CmpNone
		}
		if s == BaseChar {
			switch d & MaskElem {
			case KindRaw, KindUUID, KindTime, KindSpan:
				return CmpCheckSpec
			}
		}
		return CmpConvSpec
	}
	// handle container base type list and dict
	if d == BaseList {
		if s&BaseList == 0 {
			return CmpNone
		}
		return CmpConvList
	}
	if d == BaseDict {
		if s&BaseDict == 0 {
			return CmpNone
		}
		return CmpConvDict
	}
	if s == BaseList {
		switch d & MaskElem {
		case KindArr, KindObj:
		default:
			return CmpNone
		}
		return CmpCheckList
	}
	if s == BaseDict {
		switch d & MaskElem {
		case KindMap, KindObj:
		default:
			return CmpNone
		}
		return CmpCheckDict
	}
	// handle specific container src type
	switch s & MaskElem {
	case KindArr:
		if s&MaskElem == KindArr {
			sub := Compare(src.Sub(), dst.Sub())
			if sub > LvlConv {
				return CmpConvArr
			}
			if sub > LvlCheck {
				return CmpCheckArr
			}
		}
	case KindMap:
		if s&MaskElem == KindMap {
			sub := Compare(src.Sub(), dst.Sub())
			if sub > LvlConv {
				return CmpConvMap
			}
			if sub > LvlCheck {
				return CmpCheckArr
			}
		}
	case KindObj:
		if s&MaskElem == KindObj {
			sub := compareInfo(src.Info, dst.Info)
			if sub > LvlConv {
				return CmpConvObj
			}
			if sub > LvlCheck {
				return CmpCheckObj
			}
		}
	}
	return CmpNone
}

func compareInfo(src, dst *Info) Cmp {
	if src.IsZero() {
		if dst.IsZero() {
			return CmpSame
		}
		return CmpNone
	}
	if dst.IsZero() {
		return CmpNone
	}
	if src.Key() != dst.Key() {
		return CmpNone
	}
	if len(src.Fields) == 0 {
		if len(dst.Fields) == 0 {
			return CmpSame
		}
		return CmpNone
	}
	if len(dst.Fields) == 0 {
		return CmpNone
	}
	var res Cmp
	for _, df := range dst.Fields {
		sf := findField(src.Fields, df.Name)
		if sf == nil {
			if !df.Opt() { // field is required
				return CmpNone
			}
			continue
		}
		c := Compare(sf.Type, df.Type)
		if c < LvlCheck { // incompatible field
			return c
		}
		if res == CmpNone || c&LvlMask == res&LvlMask {
			res = res | c
		} else if c < res {
			res = c
		}
	}
	return res
}

func findField(fs []Field, key string) *Field {
	for _, f := range fs {
		if f.Name == key {
			return &f
		}
	}
	return nil
}
