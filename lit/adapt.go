package lit

import (
	"reflect"
	"time"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// Adapt returns a literal adapter for the interface value v or an error.
func Adapt(v interface{}) (Lit, error) {
	return AdaptValue(reflect.ValueOf(v))
}

// AdaptValue returns a literal adapter for the reflect value val or an error.
func AdaptValue(val reflect.Value) (Lit, error) {
	if !val.IsValid() {
		return Nil, nil
	}
	var ptr bool
	v, t := val, val.Type()
	if t.Implements(refLit) {
		return v.Interface().(Lit), nil
	}
	if ptr = t.Kind() == reflect.Ptr; ptr {
		t = t.Elem()
		v = v.Elem()
		if !v.IsValid() {
			t, err := ReflectType(t)
			if err != nil {
				return nil, err
			}
			return Null(t), nil
		}
	}
	var l Lit
	switch t.Kind() {
	case reflect.Bool:
		l = Bool(v.Bool())
	case reflect.Int64:
		if isRef(t, refSecs) {
			if v, ok := toRef(t, refSpan, v); ok {
				l = Span(v.Interface().(time.Duration))
				break
			}
		}
		fallthrough
	case reflect.Int, reflect.Int32:
		l = Int(v.Int())
	case reflect.Uint64:
		// TODO check flags
		fallthrough
	case reflect.Uint, reflect.Uint32:
		l = Int(int64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		l = Real(v.Float())
	case reflect.String:
		// TODO check enum
		l = Str(v.String())
	case reflect.Struct:
		if v, ok := toRef(t, refTime, v); ok {
			l = Time(v.Interface().(time.Time))
			break
		}
		// TODO check rec
		res, err := adaptObj(v)
		if err != nil {
			return nil, err
		}
		l = res
	case reflect.Array:
		if v, ok := toRef(t, refUUID, v); ok {
			l = UUID(v.Interface().([16]byte))
		}
	case reflect.Map:
		return adaptMap(v)
	case reflect.Slice:
		if v, ok := toRef(t, refRaw, v); ok {
			l = Raw(v.Bytes())
		}
		return adaptArr(v)
	case reflect.Interface:
		if v.IsNil() {
			return Nil, nil
		}
		l, err := AdaptValue(v.Elem())
		if err != nil {
			return nil, err
		}
		return Any{l}, nil
	}
	if l == nil {
		return nil, cor.Error("not adaptable")
	}
	if ptr {
		l = Some{l}
	}
	return l, nil
}

func adaptArr(v reflect.Value) (Arr, error) {
	et, err := ReflectType(v.Type().Elem())
	if err != nil {
		return nil, err
	}
	n := v.Len()
	res, err := MakeArr(typ.Arr(et), n)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		el, err := AdaptValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		err = res.SetIdx(i, el)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func adaptMap(v reflect.Value) (Map, error) {
	mt, err := ReflectType(v.Type())
	if err != nil {
		return nil, err
	}
	keys := v.MapKeys()
	res, err := MakeMapCap(mt, len(keys))
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		el, err := AdaptValue(v.MapIndex(k))
		if err != nil {
			return nil, err
		}
		err = res.SetKey(k.String(), el)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func adaptObj(v reflect.Value) (Obj, error) {
	nfo, idx, err := reflectFields(v.Type(), make(infoMap))
	if err != nil {
		return nil, err
	}
	res, err := MakeObj(typ.Obj(nfo.Params))
	if err != nil {
		return nil, err
	}
	for i, f := range nfo.Params {
		el, err := AdaptValue(v.FieldByIndex(idx[i]))
		if err != nil {
			return nil, err
		}
		err = res.SetKey(f.Key(), el)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func toRef(t reflect.Type, ref reflect.Type, v reflect.Value) (reflect.Value, bool) {
	if t == ref {
		return v, true
	}
	if t.ConvertibleTo(ref) {
		return v.Convert(ref), true
	}
	return v, false
}
