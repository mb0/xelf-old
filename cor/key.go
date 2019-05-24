package cor

import "strings"

// KeyStart tests whether r is ascii lowercase letter or underscore.
func KeyStart(r rune) bool { return r >= 'a' && r <= 'z' || r == '_' }

// KeyPart tests whether r is ascii lowercase letter, digit, dot or underscore.
func KeyPart(r rune) bool { return KeyStart(r) || Digit(r) || r == '.' }

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
			s = s[i:]
			break
		}
	}
	for i, c := range s {
		if !NamePart(c) && c != '.' {
			s = s[:i]
			break
		}
	}
	return strings.ToLower(s)
}

const trans = "aáaàaâaãaåaāaăaąeéeèeéeêeëeēeĕeėeęeěiìiíiîiïiìiĩiīiĭoóoôoõoōoŏoőuùuúuûuũuūuŭuůcçyÿnñ"

func Keyify(s string) string {
	sep := true
	var buf strings.Builder
	for _, r := range strings.ToLower(s) {
		if KeyPart(r) {
			sep = false
			buf.WriteRune(r)
		} else if idx := strings.IndexRune(trans, r); idx > 0 {
			sep = false
			buf.WriteByte(trans[idx-1])
		} else {
			switch r {
			case 'ä':
				buf.WriteString("ae")
			case 'ö':
				buf.WriteString("oe")
			case 'ü':
				buf.WriteString("ue")
			case 'ß':
				buf.WriteString("ss")
			case 'æ':
				buf.WriteString("ae")
			case 'œ':
				buf.WriteString("oe")
			case '€':
				buf.WriteString("euro")
			case '$':
				buf.WriteString("dollar")
			case '£':
				buf.WriteString("pound")
			case '¥':
				buf.WriteString("yen")
			default:
				if !sep {
					sep = true
					buf.WriteByte('_')
				}
				continue
			}
			sep = false
		}
	}
	return strings.Trim(buf.String(), "_")
}
