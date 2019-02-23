package cor

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

// Raw parses s and returns a pointer to a byte slice or nil on error.
func Raw(s string) *[]byte {
	v, err := ParseRaw(s)
	if err != nil {
		return nil
	}
	return &v
}

// FormatRaw returns v as string, plain texts are returned as is, otherwise FormatHex is used.
func FormatRaw(v []byte) string {
	if isPlain(v) {
		return string(v)
	}
	return FormatHex(v)
}

// FormatHex returns v as string starting with '\x' and followed by the bytes as lower hex.
func FormatHex(v []byte) string {
	var b strings.Builder
	b.Grow(2 + len(v)*2)
	b.WriteString(`\x`)
	hex.NewEncoder(&b).Write(v)
	return b.String()
}

// ParseRaw parses s and returns a byte slice or an error.
func ParseRaw(s string) ([]byte, error) {
	if strings.HasPrefix(s, `\x`) {
		return hex.DecodeString(s[2:])
	}
	return []byte(s), nil
}

// isPlain returns whether b consist only of basic latin and latin-1 supplement none-control runes.
func isPlain(b []byte) bool {
	for len(b) > 0 {
		r, n := utf8.DecodeRune(b)
		if !isPlainRune(r) {
			return false
		}
		b = b[n:]
	}
	return true
}

func isPlainRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' ||
		r > 0x20 && r < 0x7f || r >= 0xad && r <= 0xff
}
