package cor

import (
	"strconv"
	"strings"
	"unicode/utf16"
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
	var last rune
	const hex = "0123456789abcdef"
	for w := 0; len(s) > 0; s = s[w:] {
		r := rune(s[0])
		w = 1
		if r < utf8.RuneSelf {
			switch r {
			case rune(q):
				b.WriteByte('\\')
				b.WriteByte(q)
			case '\\':
				b.WriteString(`\\`)
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			case '/':
				if last == '<' {
					b.WriteString(`\/`)
				} else {
					b.WriteRune(r)
				}
			default:
				if r >= 0x20 && r <= 0x7e {
					b.WriteByte(byte(r))
				} else {
					b.WriteString(`\u00`)
					b.WriteByte(hex[r>>4])
					b.WriteByte(hex[r&0xf])
				}
			}
		} else {
			r, w = utf8.DecodeRuneInString(s)
			if r == utf8.RuneError && w == 1 {
				b.WriteString(`\ufffd`)
			} else if r == '\u2028' || r == '\u2029' {
				b.WriteString(`\u202`)
				b.WriteByte(hex[r&0xf])
			} else {
				b.WriteRune(r)
			}
		}
		last = r
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
	idx := -1
	for i := 0; i < len(s); i++ {
		if c := s[i]; c == '\\' || c == q {
			idx = i
			break
		} else if c == '\n' {
			return "", strconv.ErrSyntax
		}
	}
	if idx == -1 { // is trivial
		return s, nil
	}
	var b strings.Builder
	b.Grow(len(s)) // try to avoid more allocations.
	b.WriteString(s[:idx])
	s = s[idx:]
	for len(s) > 0 {
		c := s[0]
		if c == q {
			return "", strconv.ErrSyntax
		}
		if c >= utf8.RuneSelf {
			r, w := utf8.DecodeRuneInString(s)
			b.WriteRune(r)
			s = s[w:]
			continue
		}
		if c != '\\' {
			b.WriteByte(c)
			s = s[1:]
			continue
		}
		if len(s) < 2 {
			return "", strconv.ErrSyntax
		}
		x := s[1]
		s = s[2:]
		switch x {
		case q, '\\', '/':
			b.WriteByte(x)
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'u':
			r, err := u4(s)
			if err != nil {
				return "", err
			}
			s = s[4:]
			if utf16.IsSurrogate(r) && len(s) > 5 && s[0] == '\\' && s[1] == 'u' {
				r1, err := u4(s[2:])
				if err != nil {
					return "", err
				}
				r = utf16.DecodeRune(r, r1)
				s = s[6:]
			}
			b.WriteRune(r)
			continue
		default:
			return "", strconv.ErrSyntax
		}
	}
	return b.String(), nil
}

func u4(s string) (r rune, _ error) {
	if len(s) < 4 {
		return 0, strconv.ErrSyntax
	}
	for i := 0; i < 4; i++ {
		h, ok := fromHex(s[i])
		if !ok {
			return 0, strconv.ErrSyntax
		}
		r = r<<4 | rune(h)
	}
	return r, nil
}
