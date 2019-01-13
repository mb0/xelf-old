package typ

// Kind is a bit-set describing a type. It represents all type information except reference names
// and object fields. It is a handy implementation detail, but not part of the xelf specification.
type Kind uint64

// A Kind consists of up to seven slots each eight bits wide. The first slot uses the least
// significant byte. The following slots are only used for arr and map type slots.
const (
	SlotCount = 7
	SlotSize  = 8
)

// Each bit in a slot has a certain meaning. The first four bits specify a base type, next two bits
// further specify the type. The last two bits flag a type as optional or reference version.
const (
	BaseNum  = 1 << iota // 0000 0001
	BaseChar             // 0000 0010
	BaseList             // 0000 0100
	BaseDict             // 0000 1000
	Spec1                // 0001 0000
	Spec2                // 0010 0000
	FlagRef              // 0100 0000
	FlagOpt              // 1000 0000

	Spec3    = Spec1 | Spec2       // 0011 0000
	MaskPrim = BaseNum | BaseChar  // 0000 0011
	MaskCont = BaseList | BaseDict // 0000 1100
	MaskBase = MaskPrim | MaskCont // 0000 1111
	MaskElem = MaskBase | Spec3    // 0011 1111
)

const (
	KindVoid = 0x00
	KindRef  = FlagRef
	KindAny  = FlagOpt

	KindBool = BaseNum | Spec1
	KindInt  = BaseNum | Spec2
	KindReal = BaseNum | Spec3

	KindStr  = BaseChar | Spec1
	KindRaw  = BaseChar | Spec2
	KindUUID = BaseChar | Spec3

	KindTime = BaseChar | BaseNum | Spec1
	KindSpan = BaseChar | BaseNum | Spec2

	KindArr = BaseList | Spec1
	KindMap = BaseDict | Spec1
	KindObj = BaseDict | BaseList | Spec1

	KindFlag = FlagRef | KindInt
	KindEnum = FlagRef | KindStr
	KindRec  = FlagRef | KindObj
)
