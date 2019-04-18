package cor

import (
	"sort"
	"strings"
)

// Const represents named integer constant for flags or enums
type Const = struct {
	Name string `json:"name"`
	Val  int64  `json:"val"`
}

// ConstByKey finds and returns a constant with key in s. If a const was found, ok is true.
func ConstByKey(s []Const, key string) (c Const, ok bool) {
	for _, e := range s {
		if strings.EqualFold(key, e.Name) {
			return e, true
		}
	}
	return
}

// ConstByVal finds and returns a constant with value val in s. If a const was found, ok is true.
func ConstByVal(s []Const, val int64) (c Const, ok bool) {
	for _, e := range s {
		if val == e.Val {
			return e, true
		}
	}
	return
}

// FormatEnum returns the lowercase name of the constant matching val or an empty string.
func FormatEnum(s []Const, val int64) string {
	if c, ok := ConstByVal(s, val); ok {
		return LastKey(c.Name)
	}
	return ""
}

// FormatFlag returns a string representing mask. It returns the matched constants'
// lowercase names seperated by a pip '|'.
func FormatFlag(s []Const, mask int64) string {
	res := GetFlags(s, uint64(mask))
	switch len(res) {
	case 0:
		return ""
	case 1:
		return LastKey(res[0].Name)
	}
	var b strings.Builder
	for i, r := range res {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(LastKey(r.Name))
	}
	return b.String()
}

// GetFlags returns the matching constants s contained in mask. The given constants are checkt in
// reverse and thus should match combined, more specific constants first.
func GetFlags(s []Const, mask uint64) []Const {
	if len(s) == 0 {
		return nil
	}
	res := make([]Const, 0, 4)
	for i := len(s) - 1; i >= 0 && mask != 0; i-- {
		e := s[i]
		b := uint64(e.Val)
		if mask&b == b {
			mask &^= b
			res = append(res, e)
		}
	}
	sort.Sort(byVal(res))
	return res
}

type byVal []Const

func (a byVal) Len() int           { return len(a) }
func (a byVal) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byVal) Less(i, j int) bool { return a[i].Val < a[j].Val }
