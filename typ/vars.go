package typ

import (
	"fmt"
	"sort"
)

// Vars is a sorted set of type variable kinds.
type Vars []Kind

func (vs Vars) Len() int           { return len(vs) }
func (vs Vars) Less(i, j int) bool { return vs[i] < vs[j] }
func (vs Vars) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }

// Copy returns a copy of vs
func (vs Vars) Copy() Vars { return append(make(Vars, 0, len(vs)+8), vs...) }

func (vs Vars) idx(v Kind) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

// Has returns whether vs contains type variable v.
func (vs Vars) Has(v Kind) bool {
	i := vs.idx(v)
	return i < len(vs) && vs[i] == v
}

// Add inserts v into vs and returns the resulting set.
func (vs Vars) Add(v Kind) Vars {
	i := vs.idx(v)
	if i >= len(vs) {
		vs = append(vs, v)
	} else if vs[i] != v {
		vs = append(vs[:i+1], vs[i:]...)
		vs[i] = v
	}
	return vs
}

// Del removes v from vs and returns the resulting set.
func (vs Vars) Del(v Kind) Vars {
	i := vs.idx(v)
	if i < len(vs) && vs[i] == v {
		vs = append(vs[:i], vs[i+1:]...)
	}
	return vs
}

// Bind represents a type variable binding to another type
type Bind struct {
	Var Kind
	Type
}

func (b Bind) String() string { return fmt.Sprintf("%s = %s", b.Var, b.Type) }

// Binds is a sorted set of type variable bindings.
type Binds []Bind

func (bs Binds) Len() int           { return len(bs) }
func (bs Binds) Less(i, j int) bool { return bs[i].Var < bs[j].Var }
func (bs Binds) Swap(i, j int)      { bs[i], bs[j] = bs[j], bs[i] }

// Copy returns a copy of bs.
func (bs Binds) Copy() Binds { return append(make(Binds, 0, len(bs)+8), bs...) }

func (bs Binds) idx(v Kind) int {
	return sort.Search(len(bs), func(i int) bool { return bs[i].Var >= v })
}

// Get returns the type bound to v and a boolean indicating whether a binding was found.
func (bs Binds) Get(v Kind) (Type, bool) {
	i := bs.idx(v)
	if i >= len(bs) || bs[i].Var != v {
		return Void, false
	}
	return bs[i].Type, true
}

// Set inserts a binding with v and t to bs and returns the resulting set.
func (bs Binds) Set(v Kind, t Type) Binds {
	i := bs.idx(v)
	if i >= len(bs) {
		return append(bs, Bind{v, t})
	}
	if bs[i].Var != v {
		bs = append(bs[:i+1], bs[i:]...)
	}
	bs[i] = Bind{v, t}
	return bs
}

// Del removes a binding with v from bs and returns the resulting set.
func (bs Binds) Del(v Kind) Binds {
	i := bs.idx(v)
	if i < len(bs) && bs[i].Var == v {
		return append(bs[:i], bs[i+1:]...)
	}
	return bs
}
