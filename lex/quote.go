package lex

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Quote is a json compatible string quoter that works for single and double quotes and back ticks.
func Quote(s string, q byte) (string, error) {
	if q == '`' {
		if strings.ContainsRune(s, '`') {
			return "", strconv.ErrSyntax
		}
		var b strings.Builder
		b.Grow(2 + len(s)) // try to avoid more allocations
		b.WriteByte(q)
		b.WriteString(s)
		b.WriteByte(q)
		return b.String(), nil
	}
	if q != '"' && q != '\'' {
		return "", strconv.ErrSyntax
	}
	var b strings.Builder
	b.Grow(2 + 3*len(s)/2) // try to avoid more allocations.
	b.WriteByte(q)
	var runeTmp [utf8.UTFMax]byte
	for w := 0; len(s) > 0; s = s[w:] {
		r := rune(s[0])
		w = 1
		if r >= utf8.RuneSelf {
			r, w = utf8.DecodeRuneInString(s)
		}
		if r == rune(q) || r == '\\' { // always backslashed
			b.WriteByte('\\')
			b.WriteByte(byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			b.Write(runeTmp[:n])
			continue
		}
		switch r {
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			switch {
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				b.WriteString(`\u`)
				for s := 12; s >= 0; s -= 4 {
					hex := "0123456789abcdef"[r>>uint(s)&0xF]
					b.WriteByte(hex)
				}
			default:
				return "", strconv.ErrSyntax
			}
		}
	}
	b.WriteByte(q)
	return b.String(), nil
}

// Unquote is a stripped version of strconv.Unquote.
// It does not complain about single-quoted string having length greater than 1.
func Unquote(s string) (string, error) {
	n := len(s)
	if n < 2 {
		return "", strconv.ErrSyntax
	}
	q := s[0]
	if q != '"' && q != '\'' && q != '`' {
		return "", strconv.ErrSyntax
	}
	if q != s[n-1] {
		return "", strconv.ErrSyntax
	}
	s = s[1 : n-1]
	if q == '`' {
		return s, nil
	}
	if contains(s, '\n') {
		return "", strconv.ErrSyntax
	}
	// Is it trivial?  avoid allocation
	if !contains(s, '\\') && !contains(s, q) {
		return s, nil
	}
	var b strings.Builder
	b.Grow(3 * len(s) / 2) // try to avoid more allocations.
	var runeTmp [utf8.UTFMax]byte
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, q)
		if err != nil {
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			b.WriteByte(byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			b.Write(runeTmp[:n])
		}
	}
	return b.String(), nil
}

func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}
