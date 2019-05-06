package typ

import (
	"sort"
	"strings"

	"github.com/mb0/xelf/cor"
)

func Constants(m map[string]int64) Consts {
	res := make(Consts, 0, len(m))
	for name, val := range m {
		res = append(res, Const{name, val})
	}
	sort.Sort(res)
	return res
}

// Const represents named integer constant for flags or enums
type Const struct {
	Name string `json:"name"`
	Val  int64  `json:"val"`
}

func (c Const) Key() string   { return cor.LastKey(c.Name) }
func (c Const) Cased() string { return cor.Cased(c.Name) }

type Consts []Const

// ConstByKey finds and returns a constant with key in s. If a const was found, ok is true.
func (cs Consts) ByKey(key string) (c Const, ok bool) {
	for _, e := range cs {
		if key == e.Key() {
			return e, true
		}
	}
	return
}

// ConstByVal finds and returns a constant with value val in s. If a const was found, ok is true.
func (cs Consts) ByVal(val int64) (c Const, ok bool) {
	for _, e := range cs {
		if val == e.Val {
			return e, true
		}
	}
	return
}

// FormatEnum returns the lowercase name of the constant matching val or an empty string.
func (cs Consts) FormatEnum(val int64) string {
	if c, ok := cs.ByVal(val); ok {
		return c.Key()
	}
	return ""
}

// FormatFlag returns a string representing mask. It returns the matched constants'
// lowercase names seperated by a pip '|'.
func (cs Consts) FormatFlag(mask int64) string {
	res := cs.Flags(uint64(mask))
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
func (cs Consts) Flags(mask uint64) Consts {
	if len(cs) == 0 {
		return nil
	}
	res := make(Consts, 0, 4)
	for i := len(cs) - 1; i >= 0 && mask != 0; i-- {
		e := cs[i]
		b := uint64(e.Val)
		if mask&b == b {
			mask &^= b
			res = append(res, e)
		}
	}
	sort.Sort(res)
	return res
}

func (cs Consts) Len() int           { return len(cs) }
func (cs Consts) Swap(i, j int)      { cs[i], cs[j] = cs[j], cs[i] }
func (cs Consts) Less(i, j int) bool { return cs[i].Val < cs[j].Val }
