package utl

import (
	"sync"

	"github.com/mb0/xelf/exp"
)

// Fmap is a go template.FuncMap compatible map type alias
type Fmap = map[string]interface{}

// MustReflectFmap reflects m and returns a lookup function or panics.
func MustReflectFmap(m Fmap) exp.LookupFunc {
	f, err := ReflectFmap(m)
	if err != nil {
		panic(err)
	}
	return f
}

// ReflectFmap reflects m and returns a lookup function or an error.
func ReflectFmap(m Fmap) (exp.LookupFunc, error) {
	res := make(map[string]*exp.Spec, len(m))
	for key, val := range m {
		f, err := ReflectFunc(key, val)
		if err != nil {
			return nil, err
		}
		res[key] = f
	}
	return func(sym string) *exp.Spec {
		return res[sym]
	}, nil
}

// LazyFmap reflects a fmap on first use and is protected by a mutex
type LazyFmap struct {
	mu sync.Mutex
	m  Fmap
	l  exp.LookupFunc
}

// Lazy returns a LazyFmap for fmap
func Lazy(fmap Fmap) *LazyFmap {
	return &LazyFmap{m: fmap}
}

// Lookup returns the same fmap lookup function, it is created once on the first call
func (m *LazyFmap) Lookup() exp.LookupFunc {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.l == nil {
		m.l = MustReflectFmap(m.m)
	}
	return m.l
}
