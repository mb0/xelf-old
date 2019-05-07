package typ

// Cmp is a bit-set for detailed type comparisons results
type Cmp uint64

// Exclusive level bits define the meaning of the other bits. Levels are check, conv, wrap and equal.
// These level bits are the most significant in the set and allow simple compatibility tests.
//     res < LvlConv
const (
	LvlAbstr Cmp = 1 << (iota + 24)
	LvlCheck
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
	// convert from one to another list
	CmpConvList = LvlConv | (1 << iota)
	// convert from one to another dict
	CmpConvDict
	// convert from rec to another rec
	CmpConvRec
)

const (
	// resolve unresolved type references in either src or dst
	CmpCheckRef = LvlCheck | (1 << iota)
	// compare src any value type to dst
	CmpCheckAny
	// parse base to spec type
	CmpCheckSpec
	// compare list element value types to dst arr or rec
	CmpCheckListAny
	// compare dict element value types to dst map or rec
	CmpCheckDictAny
	// try to convert arr to another arr
	CmpCheckList
	// try to convert map to another map
	CmpCheckDict
	// try to convert rec to another rec
	CmpCheckRec
)
const (
	CmpAbstrPrim = LvlAbstr | (1 << iota)
	CmpAbstrCont
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
		m = CmpCheckListAny
	case CmpCompDict:
		m = CmpCheckDictAny
	case CmpCheckAny:
		m = CmpCompAny
	case CmpCheckSpec:
		m = CmpCompBase
	case CmpCheckListAny:
		m = CmpCompList
	case CmpCheckDictAny:
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
	if s == KindVoid || d == KindVoid {
		return CmpNone
	}
	if s == KindRef || d == KindRef {
		// ref always need to be resolved first @a != @a
		return CmpCheckRef
	}
	if s == d { // same kind
		// type constraints not relevant to var identity
		if s&SlotMask == KindVar {
			return CmpSame
		}
		// otherwise an equal type is the same same
		if src.Info.equal(dst.Info, s&KindCtx != 0, nil) {
			return CmpSame
		}
	}
	if s == KindSch || d == KindSch {
		// schema refs need to be resolved if not the same
		return CmpCheckRef
	}
	s, d = s&SlotMask, d&SlotMask
	// always infer variable destination
	if d == KindVar {
		return CmpInfer
	}
	// we can work with flags and enums as is but rec must be resolved
	if s == KindObj && !dst.HasParams() {
		return CmpCheckRef
	}
	if s == KindObj && !src.HasParams() {
		return CmpCheckRef
	}
	// rule out special types, which have strict equality
	if s&KindAny == 0 || d&KindAny == 0 {
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
	if d == KindNum || d == KindChar {
		if s&d == 0 && s != KindTime && s != KindSpan {
			return CmpNone
		}
		return CmpCompBase
	}
	if s == KindNum || s == KindChar {
		if d&s == 0 && d != KindTime && d != KindSpan {
			return CmpNone
		}
		if s == KindChar {
			switch d & MaskElem {
			case KindRaw, KindUUID, KindTime, KindSpan:
				return CmpCheckSpec
			}
		}
		return CmpCompSpec
	}
	// handle container base type list and dict
	if d == KindList {
		if s&KindIdxr == 0 {
			return CmpNone
		}
		el := dst.Elem()
		if el == Any {
			return CmpCompList
		}
		se := src.Elem()
		if se == Any {
			return CmpCheckListAny
		}
		sub := Compare(src.Elem(), dst.Elem())
		if sub > LvlConv {
			return CmpConvList
		}
		if sub > LvlCheck {
			return CmpCheckList
		}
		return CmpNone
	}
	if d == KindDict {
		if s&(d&KindCont) == 0 {
			return CmpNone
		}
		de := dst.Elem()
		if de == Any {
			return CmpCompDict
		}
		se := src.Elem()
		if se == Any {
			return CmpCheckDictAny
		}
		sub := Compare(src.Elem(), dst.Elem())
		if sub > LvlConv {
			return CmpConvDict
		}
		if sub > LvlCheck {
			return CmpCheckDict
		}
		return CmpNone
	}
	if d&MaskElem == KindRec {
		if s&KindCont == 0 {
			return CmpNone
		}
		if !src.HasParams() {
			if s == KindDict {
				return CmpCheckDictAny
			}
			return CmpCheckListAny
		}
		if s&MaskElem == KindRec {
			sub := compareInfo(src.Info, dst.Info)
			if sub > LvlConv {
				return CmpConvRec
			}
			if sub > LvlCheck {
				return CmpCheckRec
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
		res = CmpConvRec
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
		sf := findField(src.Params, df.Key())
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
		if f.Key() == key {
			return &f
		}
	}
	return nil
}
