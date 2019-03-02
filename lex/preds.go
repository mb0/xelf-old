package lex

// IsSpace tests whether r is a space, tab or newline.
func IsSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// IsDigit tests whether r is an ascii digit.
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsLetter tests whether r is an ascii letter.
func IsLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// IsPunct tests whether r is one of the ascii punctuations allowed in symbols.
func IsPunct(r rune) bool {
	switch r {
	case ':', '!', '#', '$', '%', '&', '*', '+', '-', '.', '/', '@', '=', '?', '|', '~', '_', '^':
		return true
	}
	return false
}

// IsNamePart tests whether r is ascii letter, digit or underscore.
func IsNamePart(r rune) bool {
	return IsLetter(r) || IsDigit(r) || r == '_'
}

// IsSign tests whether r is a plus or minus sign.
func IsSign(r byte) bool {
	return r == '+' || r == '-'
}

// IsName tests whether s is a valid name.
func IsName(s string) bool {
	if len(s) == 0 || !IsLetter(rune(s[0])) {
		return false
	}
	for _, r := range s[1:] {
		if !IsNamePart(r) {
			return false
		}
	}
	return true
}

// IsTag tests whether s starts with a colon.
func IsTag(s string) bool {
	return s != "" && s[0] == ':'
}

// IsTag tests whether s starts with a plus or minus sign.
func IsDecl(s string) bool {
	return s != "" && IsSign(s[0])
}

// IsExp tests whether t is a non-empty expression tree.
func IsExp(t *Tree) bool {
	return t.Tok == '(' && len(t.Seq) > 0
}

// IsSym tests whether t is a non-empty sym token tree.
func IsSym(t *Tree) bool {
	return t.Tok == Sym && len(t.Val) > 0
}

// SymPred returns a tree predicated that checks for symbols nested up to depth.
func SymPred(depth int, pred func(string) bool) func(*Tree) bool {
	return func(t *Tree) bool { _, ok := CheckSym(t, depth, pred); return ok }
}

// CheckSym checks t or the nested first trees up to depth whether it is a symbol that matches pred.
func CheckSym(t *Tree, depth int, pred func(string) bool) (string, bool) {
	if IsSym(t) && pred(t.Val) {
		return t.Val, true
	} else if depth > 0 && IsExp(t) {
		return CheckSym(t.Seq[0], depth-1, pred)
	}
	return "", false
}
