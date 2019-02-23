package cor

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrSpan indicates an invalid input format when parsing a span.
var ErrSpan = StrError("invalid span format")

// Milli returns the milliseconds in v as integer.
func Milli(v time.Duration) int64 { return int64(v / time.Millisecond) }

// MilliSpan returns milliseconds n as time duration.
func MilliSpan(n int64) time.Duration { return time.Duration(n) * time.Millisecond }

// Span parses s and return a pointer to a time duration or nil on error.
func Span(s string) *time.Duration {
	v, err := ParseSpan(s)
	if err != nil {
		return nil
	}
	return &v
}

// FormatSpan returns v in the string format '-123:04:05.678'.
func FormatSpan(v time.Duration) string {
	var b strings.Builder
	var d = int64(v)
	if d < 0 {
		d = -d
		b.WriteByte('-')
	}
	var pr bool
	h := d / int64(time.Hour)
	d = d % int64(time.Hour)
	if pr = h != 0; pr {
		fmt.Fprintf(&b, "%d:", h)
	}
	m := d / int64(time.Minute)
	d = d % int64(time.Minute)
	if pr = pr || m != 0; pr {
		fmt.Fprintf(&b, "%02d:", m)
	}
	s := d / int64(time.Second)
	d = d % int64(time.Second)
	if pr = pr || s != 0; pr {
		fmt.Fprintf(&b, "%02d", s)
	}
	ms := d / int64(time.Millisecond)
	if ms != 0 {
		if pr {
			fmt.Fprintf(&b, ".%03d", ms)
		} else {
			fmt.Fprintf(&b, "%d", ms)
		}
	} else if !pr {
		b.WriteByte('0')
	}
	return b.String()
}

// ParseSpan parses s and returns a time duration or an error.
// It accepts two formats '-123h4m5s678ms' and '-123:04:05.678'.
func ParseSpan(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	switch s[len(s)-1] { // time.Duration format must end in h, m or s
	case 'h', 'm', 's':
		return time.ParseDuration(s)
	}
	neg := s[0] == '-'
	if neg {
		s = s[1:]
	}
	dot := strings.Split(s, ".")
	if len(dot) == 0 || len(dot) > 2 {
		return 0, ErrSpan
	}
	col := strings.Split(dot[0], ":")
	if len(col) == 0 || len(col) > 3 {
		return 0, ErrSpan
	}
	var res time.Duration
	for len(col) > 0 {
		d, err := strconv.ParseInt(col[0], 10, 64)
		if err != nil {
			return 0, err
		}
		switch len(col) {
		case 3:
			res += time.Duration(d) * time.Hour
		case 2:
			res += time.Duration(d) * time.Minute
		case 1:
			res += time.Duration(d) * time.Second
		}
		col = col[1:]
	}
	if len(dot) > 1 {
		rest := dot[1]
		if len(rest) != 0 {
			d, err := strconv.ParseInt(rest, 10, 64)
			if err != nil {
				return 0, err
			}
			for u := 9 - len(rest); u > 0; u-- {
				d *= 10
			}
			res += time.Duration(d)
		}
	}
	if neg {
		return -res, nil
	}
	return res, nil
}
