package utl

import "github.com/mb0/xelf/exp"

// Fmap is a go template.FuncMap compatible map type alias
type Fmap = map[string]interface{}

// MustReflectFmap reflects m and returns a lookup function or panics.
func MustReflectFmap(m Fmap) func(string) exp.Resolver {
	f, err := ReflectFmap(m)
	if err != nil {
		panic(err)
	}
	return f
}

// ReflectFmap reflects m and returns a lookup function or an error.
func ReflectFmap(m Fmap) (func(string) exp.Resolver, error) {
	res := make(map[string]exp.Resolver, len(m))
	for key, val := range m {
		f, err := ReflectFunc(val)
		if err != nil {
			return nil, err
		}
		res[key] = f
	}
	return func(sym string) exp.Resolver {
		return res[sym]
	}, nil
}
