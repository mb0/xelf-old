package typ

import (
	"fmt"

	"github.com/mb0/xelf/bfr"
)

// Kind is a bit-set describing a type. It represents all type information except reference names
// and type parameters. It is a handy implementation detail, but not part of the xelf specification.
type Kind uint64

func (Kind) Flags() map[string]int64 { return kindConsts }

// A Kind describes a type in a slot that uses the 12 least significant bits. The rest of the bits
// are reserved to be used by specific types. Type variables use it to store a unique type id and
// other types might use it in the future to optimization access the most important type parameter
// details without chasing pointers.
const (
	SlotSize = 12
	SlotMask = 0xfff
)

// Each bit in a slot has a certain meaning. The first four bits specify a base type, next two bits
// further specify the type. The last two bits flag a type as optional or reference version.
const (
	KindNum  Kind = 1 << iota // 0x001
	KindChar                  // 0x002
	KindIdxr                  // 0x004
	KindKeyr                  // 0x008
	KindExpr                  // 0x010
	KindMeta                  // 0x020
	KindCtx                   // 0x040
	KindOpt                   // 0x080
	KindBit1                  // 0x100
	KindBit2                  // 0x200
	KindBit3                  // 0x400
	KindBit4                  // 0x800
)

const (
	MaskPrim = KindNum | KindChar            // 0000 0000 0011
	MaskCont = KindIdxr | KindKeyr           // 0000 0000 1100
	MaskBase = MaskPrim | MaskCont           // 0000 0000 1111
	MaskUber = KindExpr | KindMeta           // 0000 0011 0000
	MaskBits = 0xf00                         // 1111 0000 0000
	MaskElem = MaskBase | MaskBits           // 1111 0000 1111
	MaskRef  = MaskElem | MaskUber | KindCtx // 1111 0111 1111
	MaskLit  = MaskRef | KindOpt             // 1111 1100 1111
)

const (
	KindVoid = 0x00
	KindAny  = MaskBase

	KindBool = KindNum | KindBit1 // 0x101
	KindInt  = KindNum | KindBit2 // 0x201
	KindReal = KindNum | KindBit3 // 0x401
	KindSpan = KindNum | KindBit4 // 0x801

	KindStr  = KindChar | KindBit1 // 0x102
	KindRaw  = KindChar | KindBit2 // 0x202
	KindUUID = KindChar | KindBit3 // 0x402
	KindTime = KindChar | KindBit4 // 0x802

	KindList = KindIdxr | KindBit1 // 0x104
	KindDict = KindKeyr | KindBit2 // 0x208
	KindRec  = MaskCont | KindBit3 // 0x30c

	KindBits = KindCtx | KindInt
	KindEnum = KindCtx | KindStr
	KindObj  = KindCtx | KindRec

	KindTyp   = KindExpr | KindBit1 // 0x110
	KindFunc  = KindExpr | KindBit2 // 0x210
	KindForm  = KindCtx | KindExpr
	KindDyn   = KindExpr | KindBit3 // 0x410
	KindCall  = KindCtx | KindDyn
	KindNamed = KindExpr | KindBit4 // 0x810
	KindSym   = KindCtx | KindNamed

	KindVar = KindMeta | KindBit1 // 0x120
	KindRef = KindMeta | KindBit2 // 0x220
	KindAlt = KindMeta | KindBit3 // 0x420
)

func ParseKind(str string) (Kind, error) {
	if len(str) == 0 {
		return KindVoid, ErrInvalid
	}
	if len(str) > 5 && str[4] == '|' {
		switch str[:4] {
		case "list":
			return KindList, nil
		case "dict":
			return KindDict, nil
		}
	}
	switch str {
	case "void":
		return KindVoid, nil
	case "any":
		return KindAny, nil
	case "~typ":
		return KindTyp, nil
	case "~idxr":
		return KindIdxr, nil
	case "~keyr":
		return KindKeyr, nil
	case "list":
		return KindList, nil
	case "dict":
		return KindDict, nil
	case "~sym":
		return KindSym, nil
	case "~dyn":
		return KindDyn, nil
	case "~call":
		return KindCall, nil
	case "form":
		return KindForm, nil
	case "func":
		return KindFunc, nil
	case "~named":
		return KindNamed, nil
	case "alt":
		return KindAlt, nil
	}
	var kk Kind
	if str[len(str)-1] == '?' {
		str = str[:len(str)-1]
		kk = KindOpt
	}
	if len(str) > 5 {
		return KindVoid, ErrInvalid
	}
	switch str {
	case "~num":
		return kk | KindNum, nil
	case "~char":
		return kk | KindChar, nil
	case "bool":
		return kk | KindBool, nil
	case "int":
		return kk | KindInt, nil
	case "real":
		return kk | KindReal, nil
	case "str":
		return kk | KindStr, nil
	case "raw":
		return kk | KindRaw, nil
	case "uuid":
		return kk | KindUUID, nil
	case "time":
		return kk | KindTime, nil
	case "span":
		return kk | KindSpan, nil
	case "rec":
		return kk | KindRec, nil
	case "flag":
		return kk | KindBits, nil
	case "enum":
		return kk | KindEnum, nil
	case "obj":
		return kk | KindObj, nil
	}
	return KindVoid, ErrInvalid
}

func (k Kind) WriteBfr(b *bfr.Ctx) (err error) {
	str := simpleStr(k)
	if str != "" {
		err = b.Fmt(str)
		if k != KindAny && k&KindOpt != 0 {
			err = b.WriteByte('?')
		}
		return err
	}
	return nil
}

func (k Kind) String() string {
	str := simpleStr(k)
	if str != "" {
		if k != KindAny && k&KindOpt != 0 {
			return str + "?"
		}
		return str
	}
	return "invalid"
}

func (k Kind) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *Kind) UnmarshalText(txt []byte) error {
	kk, err := ParseKind(string(txt))
	*k = kk
	return err
}

func simpleStr(k Kind) string {
	switch k & SlotMask {
	case KindVoid:
		return "void"
	case KindAny:
		return "any"
	case KindTyp:
		return "~typ"
	case KindForm:
		return "form"
	case KindFunc:
		return "func"
	case KindDyn:
		return "~dyn"
	case KindCall:
		return "~call"
	case KindNamed:
		return "~named"
	case KindSym:
		return "~sym"
	case KindVar:
		id := k >> SlotSize
		if id == 0 {
			return "@"
		}
		return fmt.Sprintf("@%d", k>>SlotSize)
	case KindAlt:
		return "alt"
	case KindList:
		return "list"
	case KindDict:
		return "dict"
	}
	switch k & MaskRef {
	case KindRef:
		return "@"
	case KindNum:
		return "~num"
	case KindChar:
		return "~char"
	case KindIdxr:
		return "~idxr"
	case KindKeyr:
		return "~keyr"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindReal:
		return "real"
	case KindStr:
		return "str"
	case KindRaw:
		return "raw"
	case KindUUID:
		return "uuid"
	case KindTime:
		return "time"
	case KindSpan:
		return "span"
	case KindRec:
		return "rec"
	case KindBits:
		return "flag"
	case KindEnum:
		return "enum"
	case KindObj:
		return "obj"
	}
	return ""
}

var kindConsts = map[string]int64{
	"Void":  int64(KindVoid),
	"Any":   int64(KindAny),
	"Bool":  int64(KindBool),
	"Int":   int64(KindInt),
	"Real":  int64(KindReal),
	"Span":  int64(KindSpan),
	"Str":   int64(KindStr),
	"Raw":   int64(KindRaw),
	"UUID":  int64(KindUUID),
	"Time":  int64(KindTime),
	"List":  int64(KindList),
	"Dict":  int64(KindDict),
	"Rec":   int64(KindRec),
	"Bits":  int64(KindBits),
	"Enum":  int64(KindEnum),
	"Obj":   int64(KindObj),
	"Typ":   int64(KindTyp),
	"Func":  int64(KindFunc),
	"Form":  int64(KindForm),
	"Dyn":   int64(KindDyn),
	"Call":  int64(KindDyn),
	"Named": int64(KindNamed),
	"Sym":   int64(KindSym),
	"Var":   int64(KindVar),
	"Ref":   int64(KindRef),
	"Alt":   int64(KindAlt),
}
