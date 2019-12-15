package typ

var (
	Void = Type{Kind: KindVoid}
	Any  = Type{Kind: KindAny}
	Typ  = Type{Kind: KindTyp}

	Num  = Type{Kind: KindNum}
	Bool = Type{Kind: KindBool}
	Int  = Type{Kind: KindInt}
	Real = Type{Kind: KindReal}

	Char = Type{Kind: KindChar}
	Str  = Type{Kind: KindStr}
	Raw  = Type{Kind: KindRaw}
	UUID = Type{Kind: KindUUID}

	Time = Type{Kind: KindTime}
	Span = Type{Kind: KindSpan}

	Expr = Type{Kind: KindExpr}
	Sym  = Type{Kind: KindSym}
	Dyn  = Type{Kind: KindDyn}
	Call = Type{Kind: KindCall}
	Tag  = Type{Kind: KindTag}
)

func Opt(t Type) Type     { return Type{t.Kind | KindOpt, t.Info} }
func Rec(fs []Param) Type { return Type{KindRec, &Info{Params: fs}} }

func List(t Type) Type { return cont(KindList, t) }
func Dict(t Type) Type { return cont(KindDict, t) }
func Idxr(t Type) Type { return cont(KindIdxr, t) }
func Keyr(t Type) Type { return cont(KindKeyr, t) }
func Cont(t Type) Type { return cont(KindCont, t) }
func cont(k Kind, el Type) Type {
	if el == Void || el == Any {
		return Type{Kind: k}
	}
	return Type{k, &Info{Params: []Param{{Type: el}}}}
}

func Ref(n string) Type  { return Type{KindRef, &Info{Ref: n}} }
func Sch(n string) Type  { return Type{KindSch, &Info{Ref: n}} }
func Bits(n string) Type { return Type{KindBits, &Info{Ref: n}} }
func Enum(n string) Type { return Type{KindEnum, &Info{Ref: n}} }
func Obj(n string) Type  { return Type{KindObj, &Info{Ref: n}} }

func Func(n string, ps []Param) Type { return Type{KindFunc, &Info{Ref: n, Params: ps}} }
func Form(n string, ps []Param) Type { return Type{KindForm, &Info{Ref: n, Params: ps}} }

func VarKind(id uint64) Kind { return KindVar | Kind(id<<SlotSize) }
func Var(id uint64, alts ...Type) Type {
	t := Type{VarKind(id), &Info{}}
	if len(alts) != 0 {
		ps := make([]Param, 0, len(alts))
		for _, a := range alts {
			ps = append(ps, Param{Type: a})
		}
		t.Params = ps
	}
	return t
}

// IsOpt returns whether t is an optional type and not any.
func (t Type) IsOpt() bool {
	return t.Kind&KindOpt != 0 && t.Kind&MaskRef != 0
}

// Deopt returns the non-optional type of t if t is a optional type and not any,
// otherwise t is returned as is.
func (t Type) Deopt() (_ Type, ok bool) {
	if ok = t.IsOpt(); ok {
		t.Kind &^= KindOpt
	}
	return t, ok
}

// Elem returns a generalized element type for container types and void otherwise.
func (t Type) Elem() Type {
	switch t.Kind & MaskElem {
	case KindCont, KindIdxr, KindKeyr, KindList, KindDict, KindExpr:
		if !t.HasParams() {
			return Any
		}
		return t.Params[0].Type
	case KindRec, KindObj:
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
	if t.Kind&KindNum != 0 {
		return true
	}
	switch t.Kind & MaskRef {
	case KindChar, KindStr, KindEnum, KindTime:
		return true
	}
	return false
}

// Resolved returns whether t is fully resolved
func (t Type) Resolved() bool {
	switch t.Kind & MaskRef {
	case KindBits, KindEnum: // check that consts were resolved
		return t.HasConsts()
	case KindList, KindDict: // check elem type
		return t.Elem().Resolved()
	case KindObj, KindRec, KindFunc, KindForm: // check that params were resolved
		if !t.HasParams() {
			return false
		}
		for _, p := range t.Params {
			if !p.Type.Resolved() {
				return false
			}
		}
	case KindSch, KindRef, KindVar, KindAlt:
		return false
	}
	return true
}
