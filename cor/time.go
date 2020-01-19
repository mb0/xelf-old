package cor

import "time"

// UnixMilli returns a integer timestamp since unix epoch in milliseconds.
func UnixMilli(v time.Time) int64 { return v.Unix()*1000 + int64(v.Nanosecond()/1000000) }

// UnixMilliTime constructs and returns a time from the millisecond timestamp since unix epoch.
func UnixMilliTime(n int64) time.Time { return time.Unix(n/1000, (n%1000)*1000000) }

// Time parses s and returns a pointer to a time or nil on error
func Time(s string) *time.Time {
	v, err := ParseTime(s)
	if err != nil {
		return nil
	}
	return &v
}

// FormatTime returns v as string in the RFC3339 format with milliseconds.
func FormatTime(v time.Time) string {
	return v.Format("2006-01-02T15:04:05.999Z07:00")
}

// ParseTime parses s and return a time or error. It accepts variations of the RFC3339 format:
//     2006-01-02([T ]15:04(:05:999999999)?)?(Z|[+-]07([:]?00)?)?
// The returned time will be parsed in the local timezone if none is specified.
func ParseTime(s string) (time.Time, error) {
	fmt, tz := "2006-01-02", ""
	if len(s) > 10 {
		switch s[10] {
		case 'T', ' ':
			fmt += string(s[10]) + "15:04"
			if len(s) > 16 {
				switch s[16] {
				case ':': // parse nano
					fmt += ":05.999999999"
					tz = tzfmt(s[19:])
				default:
					tz = tzfmt(s[16:])
				}
			}
		default:
			tz = tzfmt(s[10:])
		}
	}
	if len(tz) != 0 {
		return time.Parse(fmt+tz, s)
	}
	return ParseTimeFormat(s, fmt)
}

// ParseTimeFormat parses s with go time fmt in the local timezone and returns a time or error.
func ParseTimeFormat(s, fmt string) (time.Time, error) {
	return time.ParseInLocation(fmt, s, time.Local)
}

func tzfmt(s string) string {
	for len(s) > 0 {
		switch s[0] {
		case 'Z', '+', '-':
		default:
			s = s[1:]
			continue
		}
		break
	}
	if len(s) == 0 {
		return ""
	}
	if s[0] == 'Z' {
		return "Z"
	}
	if len(s) > 3 {
		if s[3] == ':' {
			return "-07:00"
		}
		return "-0700"
	}
	return "-07"
}
