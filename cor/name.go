package cor

import "strings"

// Space tests whether r is a space, tab or newline.
func Space(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }

// Digit tests whether r is an ascii digit.
func Digit(r rune) bool { return r >= '0' && r <= '9' }

// Letter tests whether r is an ascii letter.
func Letter(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }

// NameStart tests whether r is ascii letter or underscore.
func NameStart(r rune) bool { return Letter(r) || r == '_' }

// NamePart tests whether r is ascii letter, digit or underscore.
func NamePart(r rune) bool { return NameStart(r) || Digit(r) }

// KeyStart tests whether r is ascii lowercase letter or underscore.
func KeyStart(r rune) bool { return r >= 'a' && r <= 'z' || r == '_' }

// KeyPart tests whether r is ascii lowercase letter, digit, dot or underscore.
func KeyPart(r rune) bool { return KeyStart(r) || Digit(r) || r == '.' }

// Punct tests whether r is one of the ascii punctuations allowed in symbols.
func Punct(r rune) bool {
	switch r {
	case '!', '#', '$', '%', '&', '*', '+', '-', '.', '/', ':',
		'<', '=', '>', '?', '@', '^', '|', '~':
		return true
	}
	return false
}

// IsName tests whether s is a valid name.
func IsName(s string) bool {
	if s == "" || !NameStart(rune(s[0])) {
		return false
	}
	for _, r := range s[1:] {
		if !NamePart(r) {
			return false
		}
	}
	return true
}

// IsKey tests whether s is a valid key.
func IsKey(s string) bool {
	if s == "" || !KeyStart(rune(s[0])) {
		return false
	}
	for _, r := range s[1:] {
		if !KeyPart(r) {
			return false
		}
	}
	return true
}

// LastName returns the last name segment of s.
func LastName(s string) string {
	start, end := -1, 0
	for i, c := range s {
		if start > -1 && end == 0 {
			if !NamePart(c) {
				end = i
			}
		} else if NameStart(c) {
			start = i
			end = 0
		}
	}
	if start < 0 {
		return ""
	}
	if end == 0 {
		return s[start:]
	}
	return s[start:end]
}

// LastKey returns the last name segment of s as key.
func LastKey(s string) string {
	start, end := -1, 0
	for i, c := range s {
		if start > -1 && end == 0 {
			if !NamePart(c) && c != '.' {
				end = i
			}
		} else if NameStart(c) {
			start = i
			end = 0
		}
	}
	if start < 0 {
		return ""
	}
	if end > 0 {
		s = s[:end]
	}
	return strings.ToLower(s[start:])
}

// Keyed returns the s starting with the first name segment as key.
func Keyed(s string) string {
	for i, c := range s {
		if NameStart(c) {
			return strings.ToLower(s[i:])
		}
	}
	return ""
}
