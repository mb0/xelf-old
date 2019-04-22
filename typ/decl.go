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

	Infer = Type{Kind: KindVar}

	Sym  = Type{Kind: ExpSym}
	Dyn  = Type{Kind: ExpDyn}
	Tag  = Type{Kind: ExpTag}
	Decl = Type{Kind: ExpDecl}
)

func Opt(t Type) Type     { return Type{t.Kind | FlagOpt, t.Info} }
func Arr(t Type) Type     { return Type{KindArr, &Info{Params: []Param{{Type: t}}}} }
func Map(t Type) Type     { return Type{KindMap, &Info{Params: []Param{{Type: t}}}} }
func Obj(fs []Param) Type { return Type{KindObj, &Info{Params: fs}} }

func Ref(n string) Type  { return Type{KindRef, &Info{Ref: n}} }
func Flag(n string) Type { return Type{KindFlag, &Info{Ref: n}} }
func Enum(n string) Type { return Type{KindEnum, &Info{Ref: n}} }
func Rec(n string) Type  { return Type{KindRec, &Info{Ref: n}} }

func Func(n string, ps []Param) Type { return Type{ExpFunc, &Info{Ref: n, Params: ps}} }
func Form(n string, ps []Param) Type { return Type{ExpForm, &Info{Ref: n, Params: ps}} }

func VarKind(id uint64) Kind { return KindVar | Kind(id<<SlotSize) }
func Var(id uint64) Type     { return Type{Kind: VarKind(id)} }

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

// Elem returns a generalized element type for container types and void otherwise.
func (t Type) Elem() Type {
	switch t.Kind & MaskElem {
	case KindArr, KindMap:
		if t.Info == nil || len(t.Params) == 0 {
			return Any
		}
		return t.Params[0].Type
	case BaseList, BaseDict:
		return Any
	case KindObj:
		// TODO consider an attempt to unify field types
		return Any
	}
	return Void
}

// Last returns the last element type if t is a arr or map type otherwise t is returned as is.
func (t Type) Last() Type {
	el := t.Elem()
	for el != Void && el != Any {
		t, el = el, el.Elem()
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
