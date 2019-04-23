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

	Idxer = Type{Kind: BaseIdxr}
	Keyer = Type{Kind: BaseKeyr}

	Infer = Type{Kind: KindVar}

	Sym  = Type{Kind: ExpSym}
	Dyn  = Type{Kind: ExpDyn}
	Tag  = Type{Kind: ExpTag}
	Decl = Type{Kind: ExpDecl}
)

func Opt(t Type) Type     { return Type{t.Kind | FlagOpt, t.Info} }
func List(t Type) Type    { return Type{KindList, &Info{Params: []Param{{Type: t}}}} }
func Dict(t Type) Type    { return Type{KindDict, &Info{Params: []Param{{Type: t}}}} }
func Rec(fs []Param) Type { return Type{KindRec, &Info{Params: fs}} }

func Ref(n string) Type  { return Type{KindRef, &Info{Ref: n}} }
func Flag(n string) Type { return Type{KindFlag, &Info{Ref: n}} }
func Enum(n string) Type { return Type{KindEnum, &Info{Ref: n}} }
func Obj(n string) Type  { return Type{KindObj, &Info{Ref: n}} }

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
	case KindList, KindDict:
		if t.Info == nil || len(t.Params) == 0 {
			return Any
		}
		return t.Params[0].Type
	case BaseIdxr, BaseKeyr:
		return Any
	case KindRec:
		// TODO consider an attempt to unify field types
		return Any
	}
	return Void
}

// Last returns the last element type if t is a list or dict type otherwise t is returned as is.
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

// Resolved returns whether t is fully resolved
func (t Type) Resolved() bool {
	switch t.Kind & SlotMask {
	case KindFlag, KindEnum: // check that consts were resolved
		return t.Info != nil && len(t.Consts) != 0
	case KindList, KindDict: // check elem type
		return t.Elem().Resolved()
	case KindObj, KindRec: // check that params were resolved
		if t.Info == nil || len(t.Params) == 0 {
			return false
		}
		for _, p := range t.Params {
			if !p.Type.Resolved() {
				return false
			}
		}
	case KindExp, KindVoid, KindRef, KindVar, KindAlt:
		return false
	}
	return true
}
