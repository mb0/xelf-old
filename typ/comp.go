package typ

// Cmp is a bit-set for detailed type comparisons results
type Cmp uint64

// Exclusive level bits define the meaning of the other bits. Levels are check, conv, wrap and equal.
// These level bits are the most significant in the set and allow simple compatibility tests.
//     res < LvlConv
const (
	LvlCheck Cmp = 1 << (iota + 24)
	LvlConv
	LvlComp
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
	CmpCompAny = LvlComp | (1 << iota)
	// convert src from spec to base type
	CmpCompBase
	// convert src from base to spec type
	CmpCompSpec
	// convert idxer to list
	CmpCompList
	// convert keyer to dict
	CmpCompDict
)

const (
	// convert from arr to another arr
	CmpConvArr = LvlConv | (1 << iota)
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

// Compare returns compatibility details for a source and destination type.
// When the result is not none or equal, it informs what to do with a value
// of the source type to arrive at the destination type.
// The flag, enum and rec are treated as their corresponding literal type.
func Compare(src, dst Type) Cmp {
	s, so := src.Deopt()
	d, do := dst.Deopt()
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

// Mirror returns the approximate mirror result of a comparison with switched arguments.
func (c Cmp) Mirror() (m Cmp) {
	switch m = c &^ (BitWrap | BitUnwrap); m {
	case CmpCompAny:
		m = CmpCheckAny
	case CmpCompBase:
		m = CmpCheckSpec
	case CmpCompSpec:
		m = CmpCompBase
	case CmpCompList:
		m = CmpCheckList
	case CmpCompDict:
		m = CmpCheckDict
	case CmpCheckAny:
		m = CmpCompAny
	case CmpCheckSpec:
		m = CmpCompBase
	case CmpCheckList:
		m = CmpCompList
	case CmpCheckDict:
		m = CmpCompDict
	}
	if c&BitWrap != 0 {
		m = m | BitUnwrap
	}
	if c&BitUnwrap != 0 {
		m = m | BitWrap
	}
	return m
}

func compare(src, dst Type) Cmp {
	s, d := src.Kind, dst.Kind
	if s == d && s != KindRef {
		if src.Info.equal(dst.Info, s&FlagRef != 0, nil) {
			return CmpSame
		}
	}
	if s == KindRef {
		return CmpCheckRef
	}
	if d == KindRef {
		if dst.Info == nil || dst.Info.Ref == "" {
			// infer dst type from src
			return CmpInfer
		}
		// ref needs to be resolved first
		return CmpCheckRef
	}
	// we can work with flags and enums as is but rec must be resolved
	if d == KindRec && (dst.Info == nil || len(dst.Params) == 0) {
		return CmpCheckRef
	}
	if s == KindRec && (src.Info == nil || len(src.Params) == 0) {
		return CmpCheckRef
	}
	// rule out special types, which have strict equality
	if m := Kind(MaskBase | FlagOpt); s&m == 0 || d&m == 0 {
		return CmpNone
	}
	// handle any, type
	if d == KindAny {
		return CmpCompAny
	}
	if s == KindAny {
		return CmpCheckAny
	}
	// handle base types starting with primitives
	if d == BaseNum || d == BaseChar {
		if s&d == 0 {
			return CmpNone
		}
		return CmpCompBase
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
		return CmpCompSpec
	}
	// handle container base type list and dict
	if d == BaseList {
		if s&BaseList == 0 {
			return CmpNone
		}
		return CmpCompList
	}
	if d == BaseDict {
		if s&BaseDict == 0 {
			return CmpNone
		}
		return CmpCompDict
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
			sub := Compare(src.Next(), dst.Next())
			if sub > LvlConv {
				return CmpConvArr
			}
			if sub > LvlCheck {
				return CmpCheckArr
			}
		}
	case KindMap:
		if s&MaskElem == KindMap {
			sub := Compare(src.Next(), dst.Next())
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
	res := CmpSame
	if src.Key() != dst.Key() {
		res = CmpConvObj
	}
	if len(src.Params) == 0 {
		if len(dst.Params) == 0 {
			return res
		}
		return CmpNone
	}
	if len(dst.Params) == 0 {
		return CmpNone
	}
	for _, df := range dst.Params {
		sf := findField(src.Params, df.Name)
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

func findField(fs []Param, key string) *Param {
	for _, f := range fs {
		if f.Name == key {
			return &f
		}
	}
	return nil
}
