package lit

import (
	"errors"
	"reflect"
	"time"

	"github.com/mb0/xelf/typ"
)

// Reflect returns the xelf type for the interface value v or an error
func Reflect(v interface{}) (typ.Type, error) {
	return ReflectType(reflect.TypeOf(v))
}

// ReflectType returns the xelf type for the reflect type t or an error
func ReflectType(t reflect.Type) (res typ.Type, err error) {
	var ptr bool
	if ptr = t.Kind() == reflect.Ptr; ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool:
		res = typ.Bool
	case reflect.Int64:
		if isRef(t, refSecs) {
			res = typ.Span
			break
		}
		fallthrough
	case reflect.Int, reflect.Int32:
		res = typ.Int
	case reflect.Uint64:
		// TODO check flags
		fallthrough
	case reflect.Uint, reflect.Uint32:
		res = typ.Int
	case reflect.Float32, reflect.Float64:
		res = typ.Real
	case reflect.String:
		// TODO check flags
		res = typ.Str
	case reflect.Struct:
		if isRef(t, refTime) {
			res = typ.Time
			break
		}
		if isRef(t, refDict.Elem()) {
			if !ptr {
				return typ.Void, typ.ErrInvalid
			}
			return typ.Dict, nil
		}
		// TODO check rec
		fs, _, err := reflectFields(t)
		if err != nil {
			return typ.Void, err
		}
		res = typ.Obj(fs)
	case reflect.Array:
		if isRef(t, refUUID) {
			res = typ.UUID
		}
	case reflect.Map:
		if !isRef(t.Key(), refStr) {
			return typ.Void, errors.New("map key must by string type")
		}
		et, err := ReflectType(t.Elem())
		if err != nil {
			return typ.Void, err
		}
		res = typ.Map(et)
	case reflect.Slice:
		if isRef(t, refRaw) {
			res = typ.Raw
			break
		}
		if isRef(t, refList) {
			return typ.List, nil
		}
		et, err := ReflectType(t.Elem())
		if err != nil {
			return typ.Void, err
		}
		res = typ.Arr(et)
	case reflect.Interface:
		return typ.Any, nil
	}
	if res.IsZero() {
		return typ.Void, typ.ErrInvalid
	}
	if ptr {
		return typ.Opt(res), nil
	}
	return res, nil
}

func reflectFields(t reflect.Type) ([]typ.Field, [][]int, error) {
	n := t.NumField()
	fs := make([]typ.Field, 0, n)
	idx := make([][]int, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		ft, err := ReflectType(f.Type)
		if err != nil {
			return nil, nil, err
		}
		fs = append(fs, typ.Field{Name: f.Name, Type: ft})
		idx = append(idx, f.Index)
	}
	return fs, idx, nil
}

var (
	refLit  = reflect.TypeOf((*Lit)(nil)).Elem()
	refStr  = reflect.TypeOf("")
	refRaw  = reflect.TypeOf([]byte(nil))
	refUUID = reflect.TypeOf([16]byte{})
	refSpan = reflect.TypeOf(time.Second)
	refTime = reflect.TypeOf(time.Time{})
	refList = reflect.TypeOf(List(nil))
	refDict = reflect.TypeOf((*Dict)(nil))
	refSecs = reflect.TypeOf((*interface{ Seconds() float64 })(nil))
)

func isRef(t reflect.Type, ref reflect.Type) bool {
	return t == ref || t.ConvertibleTo(ref)
}
