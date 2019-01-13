package typ

var (
	Void = Type{Kind: KindVoid}
	Any  = Type{Kind: KindAny}

	Num  = Type{Kind: BaseNum}
	Bool = Type{Kind: KindBool}
	Int  = Type{Kind: KindInt}
	Real = Type{Kind: KindReal}

	Char = Type{Kind: BaseChar}
	Str  = Type{Kind: KindStr}
	Raw  = Type{Kind: KindRaw}
	UUID = Type{Kind: KindUUID}

	Time = Type{Kind: KindTime}
	Span = Type{Kind: KindSpan}

	List = Type{Kind: BaseList}
	Dict = Type{Kind: BaseDict}
)

func Opt(t Type) Type     { return Type{t.Kind | FlagOpt, t.Info} }
func Arr(t Type) Type     { return Type{t.Kind<<SlotSize | KindArr, t.Info} }
func Map(t Type) Type     { return Type{t.Kind<<SlotSize | KindMap, t.Info} }
func Obj(fs []Field) Type { return Type{KindObj, &Info{Fields: fs}} }

func Ref(name string) Type          { return Type{KindRef, &Info{Ref: name}} }
func Flag(name string) Type         { return Type{KindFlag, &Info{Ref: name}} }
func Enum(name string) Type         { return Type{KindEnum, &Info{Ref: name}} }
func Rec(n string, fs []Field) Type { return Type{KindRec, &Info{Ref: n, Fields: fs}} }
