package typ

var (
	Void = Type{Kind: KindVoid}
	Any  = Type{Kind: KindAny}
	Typ  = Type{Kind: KindTyp}

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

func Ref(name string) Type  { return Type{KindRef, &Info{Ref: name}} }
func Flag(name string) Type { return Type{KindFlag, &Info{Ref: name}} }
func Enum(name string) Type { return Type{KindEnum, &Info{Ref: name}} }
func Rec(n string) Type     { return Type{KindRec, &Info{Ref: n}} }

// IsOpt returns whether t is an optional type and not any.
func (t Type) IsOpt() bool {
	return t.Kind&FlagOpt != 0 && t.Kind&MaskRef != 0
}

// Deopt returns the non-optional type of t if t is a optional type and not any,
// otherwise t is returned as is.
func (t Type) Deopt() (_ Type, ok bool) {
	if ok = t.IsOpt(); ok {
		t.Kind &^= FlagOpt
	}
	return t, ok
}

// Next returns the next sub type of t if t is a arr or map type,
// otherwise t is returned as is.
func (t Type) Next() Type {
	switch t.Kind & MaskElem {
	case KindArr, KindMap:
		t.Kind = t.Kind >> SlotSize
	}
	return t
}

// Last returns the last sub type of t if t is a arr or map type,
// otherwise t is returned as is.
func (t Type) Last() Type {
	for k := t.Kind; ; k = k >> SlotSize {
		switch k & MaskElem {
		case KindArr, KindMap:
			continue
		}
		return Type{k, t.Info}
	}
	return t
}

// Ordered returns whether type t supports ordering.
func (t Type) Ordered() bool {
	if t.Kind&BaseNum != 0 {
		return true
	}
	switch t.Kind & MaskRef {
	case BaseChar, KindStr, KindEnum:
		return true
	}
	return false
}
