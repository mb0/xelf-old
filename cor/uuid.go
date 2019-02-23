package cor

import (
	"encoding/hex"
	"strings"
)

// ErrUUID indicates an invalid input format when parsing an uuid.
var ErrUUID = StrError("invalid uuid format")

// UUID parses s and returns a pointer to the uuid bytes or nil on error.
func UUID(s string) *[16]byte {
	v, err := ParseUUID(s)
	if err != nil {
		return nil
	}
	return &v
}

// FormatUUID returns v as string in the canonical uuid format.
func FormatUUID(v [16]byte) string {
	var b strings.Builder
	w := hex.NewEncoder(&b)
	var nn int
	for i, n := range [5]int{4, 2, 2, 2, 6} {
		if i > 0 {
			b.WriteByte('-')
		}
		w.Write(v[nn : nn+n])
		nn += n
	}
	return b.String()
}

// ParseUUID parses s and return the uuid bytes or an error.
// It accepts 16 hex encoded bytes with up to four dashes in between.
func ParseUUID(s string) ([16]byte, error) {
	var res [16]byte
	if len(s) < 32 || len(s) > 36 {
		return res, ErrUUID
	}
	if len(s) > 36 {
		return res, ErrUUID
	}
	for i, o := 0, 0; i+1 < len(s) && o < 16; {
		a := s[i]
		if a == '-' {
			i++
			continue
		}
		b := s[i+1]
		a, aok := fromHex(a)
		b, bok := fromHex(b)
		if !aok || !bok {
			return res, ErrUUID
		}
		res[o] = a<<4 | b
		i += 2
		o += 1
	}
	return res, nil
}

func fromHex(b byte) (byte, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	}
	return 0, false
}
