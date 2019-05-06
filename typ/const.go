package typ

import (
	"sort"
	"strings"

	"github.com/mb0/xelf/cor"
)

func Consts(m map[string]int64) []Const {
	res := make([]Const, 0, len(m))
	for name, val := range m {
		res = append(res, Const{name, val})
	}
	sort.Sort(byVal(res))
	return res
}

// Const represents named integer constant for flags or enums
type Const struct {
	Name string `json:"name"`
	Val  int64  `json:"val"`
}

func (c Const) Key() string   { return cor.LastKey(c.Name) }
func (c Const) Cased() string { return cor.Cased(c.Name) }

// ConstByKey finds and returns a constant with key in s. If a const was found, ok is true.
func ConstByKey(s []Const, key string) (c Const, ok bool) {
	for _, e := range s {
		if key == e.Key() {
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
		return c.Key()
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
		return res[0].Key()
	}
	var b strings.Builder
	for i, r := range res {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(r.Key())
	}
	return b.String()
}

// GetFlags returns the matching constants s contained in mask. The given constants are checked in
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
