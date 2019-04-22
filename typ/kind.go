package typ

import (
	"fmt"

	"github.com/mb0/xelf/bfr"
)

// Kind is a bit-set describing a type. It represents all type information except reference names
// and record fields. It is a handy implementation detail, but not part of the xelf specification.
type Kind uint64

func (Kind) Flags() map[string]int64 { return kindConsts }

// A Kind consists of one slot that uses the least significant byte and stores the kind bits.
// The rest of the bytes is used by expression types to store extended bits and by type variables
// to store a unique type id.
const (
	SlotSize = 8
	SlotMask = 0xff
)

// Each bit in a slot has a certain meaning. The first four bits specify a base type, next two bits
// further specify the type. The last two bits flag a type as optional or reference version.
const (
	BaseNum  Kind = 1 << iota // 0000 0001
	BaseChar                  // 0000 0010
	BaseIdxr                  // 0000 0100
	BaseKeyr                  // 0000 1000
	Spec1                     // 0001 0000
	Spec2                     // 0010 0000
	FlagRef                   // 0100 0000
	FlagOpt                   // 1000 0000

	Spec3    = Spec1 | Spec2       // 0011 0000
	MaskPrim = BaseNum | BaseChar  // 0000 0011
	MaskCont = BaseIdxr | BaseKeyr // 0000 1100
	MaskBase = MaskPrim | MaskCont // 0000 1111
	MaskElem = MaskBase | Spec3    // 0011 1111
	MaskRef  = MaskElem | FlagRef  // 0111 1111
)

const (
	KindVoid = 0x00
	KindRef  = FlagRef
	KindAny  = FlagOpt
	KindTyp  = Spec1
	KindExp  = Spec2

	KindVar = FlagRef | Spec1
	KindAlt = FlagRef | Spec2

	KindBool = BaseNum | Spec1
	KindInt  = BaseNum | Spec2
	KindReal = BaseNum | Spec3

	KindStr  = BaseChar | Spec1
	KindRaw  = BaseChar | Spec2
	KindUUID = BaseChar | Spec3

	KindTime = BaseChar | BaseNum | Spec1
	KindSpan = BaseChar | BaseNum | Spec2

	KindList = BaseIdxr | Spec1
	KindDict = BaseKeyr | Spec1
	KindRec  = BaseKeyr | BaseIdxr | Spec1

	KindFlag = FlagRef | KindInt
	KindEnum = FlagRef | KindStr
	KindObj  = FlagRef | KindRec
)
const (
	ExpDyn  = KindExp | BaseIdxr<<SlotSize
	ExpForm = KindExp | BaseKeyr<<SlotSize
	ExpFunc = KindExp | KindRec<<SlotSize
	ExpSym  = KindExp | KindAny<<SlotSize
	ExpTag  = KindExp | KindStr<<SlotSize
	ExpDecl = KindExp | KindRaw<<SlotSize
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
	case "typ":
		return KindTyp, nil
	case "list":
		return BaseIdxr, nil
	case "dict":
		return BaseKeyr, nil
	case "sym":
		return ExpSym, nil
	case "dyn":
		return ExpDyn, nil
	case "form":
		return ExpForm, nil
	case "func":
		return ExpFunc, nil
	case "tag":
		return ExpTag, nil
	case "decl":
		return ExpDecl, nil
	case "alt":
		return KindAlt, nil
	}
	var kk Kind
	if str[len(str)-1] == '?' {
		str = str[:len(str)-1]
		kk = FlagOpt
	}
	if len(str) > 4 {
		return KindVoid, ErrInvalid
	}
	switch str {
	case "ref":
		return kk | KindRef, nil
	case "num":
		return kk | BaseNum, nil
	case "char":
		return kk | BaseChar, nil
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
		return kk | KindFlag, nil
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
		if k != KindAny && k&FlagOpt != 0 {
			err = b.WriteByte('?')
		}
		return err
	}
	return nil
}

func (k Kind) String() string {
	str := simpleStr(k)
	if str != "" {
		if k != KindAny && k&FlagOpt != 0 {
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
		return "typ"
	case KindExp:
		switch k {
		case ExpSym:
			return "sym"
		case ExpDyn:
			return "dyn"
		case ExpForm:
			return "form"
		case ExpFunc:
			return "func"
		case ExpTag:
			return "tag"
		case ExpDecl:
			return "decl"
		}
	case KindVar:
		id := k >> SlotSize
		if id == 0 {
			return "@"
		}
		return fmt.Sprintf("@%d", k>>SlotSize)
	case KindAlt:
		return "alt"
	case KindList:
		return "list|"
	case KindDict:
		return "dict|"
	}
	switch k & MaskRef {
	case KindRef:
		return "@"
	case BaseNum:
		return "num"
	case BaseChar:
		return "char"
	case BaseIdxr:
		return "list"
	case BaseKeyr:
		return "dict"
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
	case KindFlag:
		return "flag"
	case KindEnum:
		return "enum"
	case KindObj:
		return "obj"
	}
	return ""
}

var kindConsts = map[string]int64{
	"Ref":  int64(KindRef),
	"Any":  int64(KindAny),
	"Typ":  int64(KindTyp),
	"Exp":  int64(KindExp),
	"Bool": int64(KindBool),
	"Int":  int64(KindInt),
	"Real": int64(KindReal),
	"Str":  int64(KindStr),
	"Raw":  int64(KindRaw),
	"UUID": int64(KindUUID),
	"Time": int64(KindTime),
	"Span": int64(KindSpan),
	"List": int64(KindList),
	"Dict": int64(KindDict),
	"Rec":  int64(KindRec),
	"Flag": int64(KindFlag),
	"Enum": int64(KindEnum),
	"Obj":  int64(KindObj),
}
