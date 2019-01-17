package cor

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/mb0/xelf/lex"
)

// Raw parses s and returns a pointer to a byte slice or nil on error
func Raw(s string) *[]byte {
	v, err := ParseRaw(s)
	if err != nil {
		return nil
	}
	return &v
}

// FormatRaw returns v as a string starting with '\x' and folled by the bytes as lower hex
func FormatRaw(v []byte) string {
	if isPlain(v) {
		return string(v)
	}
	var b strings.Builder
	b.Grow(2 + len(v)*2)
	b.WriteString(`\x`)
	hex.NewEncoder(&b).Write(v)
	return b.String()
}

// ParseRaw parses s and returns a byte slice or error
func ParseRaw(s string) ([]byte, error) {
	if strings.HasPrefix(s, `\x`) {
		return hex.DecodeString(s[2:])
	}
	return []byte(s), nil
}

// isPlain returns whether b consist only of basic latin and latin-1 supplement non control runes
func isPlain(b []byte) bool {
	for len(b) > 0 {
		r, n := utf8.DecodeRune(b)
		if !(lex.IsSpace(r) || r > 0x20 && r < 0x7f || r >= 0xad && r <= 0xff) {
			return false
		}
		b = b[n:]
	}
	return true
}
