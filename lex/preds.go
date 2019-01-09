package lex

// IsSpace tests whether r is a space, tab or newline
func IsSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// IsDigit tests whether r is an ascii digit
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsLetter tests whether r is an ascii letter
func IsLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// IsPunct tests whether r is one of the ascii punctuations allowed in symbols
func IsPunct(r rune) bool {
	switch r {
	case ':', '!', '#', '$', '%', '&', '*', '+', '-', '.', '/', '@', '=', '?', '|', '~', '_':
		return true
	}
	return false
}

// IsNamePart tests whether r is ascii letter, digit or underscore
func IsNamePart(r rune) bool {
	return IsLetter(r) || IsDigit(r) || r == '_'
}

// IsName tests whether s is a valid name
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
