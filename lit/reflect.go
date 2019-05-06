package lit

import (
	"reflect"
	"strings"
	"time"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var (
	refLit  = reflect.TypeOf((*Lit)(nil)).Elem()
	refBool = reflect.TypeOf(false)
	refInt  = reflect.TypeOf(int64(0))
	refReal = reflect.TypeOf(float64(0))
	refStr  = reflect.TypeOf("")
	refRaw  = reflect.TypeOf([]byte(nil))
	refUUID = reflect.TypeOf([16]byte{})
	refSpan = reflect.TypeOf(time.Second)
	refTime = reflect.TypeOf(time.Time{})
	refList = reflect.TypeOf(Idxr(nil))
	refDict = reflect.TypeOf((*Keyr)(nil))
	refSecs = reflect.TypeOf((*MarkSpan)(nil))
	refFlag = reflect.TypeOf((*MarkFlag)(nil))
	refEnum = reflect.TypeOf((*MarkEnum)(nil))
	refType = reflect.TypeOf(typ.Void)
	refEl   = reflect.TypeOf((*interface {
		WriteBfr(*bfr.Ctx) error
		String() string
		Typ() typ.Type
	})(nil)).Elem()
)

// Reflect returns the xelf type for the interface value v or an error.
func Reflect(v interface{}) (typ.Type, error) {
	return ReflectType(reflect.TypeOf(v))
}

// ReflectType returns the xelf type for the reflect type t or an error.
func ReflectType(t reflect.Type) (res typ.Type, err error) {
	nfos := make(infoMap)
	return reflectType(t, nfos)
}

type fields = struct {
	*typ.Info
	Idx [][]int
}
type infoMap = map[reflect.Type]*fields

func getConstInfo(t reflect.Type, cs []typ.Const) *typ.Info {
	return &typ.Info{
		Ref:    t.String(),
		Consts: cs,
	}
}

func reflectType(t reflect.Type, nfos infoMap) (res typ.Type, err error) {
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
		if isRef(t, refEnum) {
			cs := reflect.Zero(t).Interface().(MarkEnum).Enums()
			res = typ.Type{typ.KindBits, getConstInfo(t, typ.Consts(cs))}
			break
		}
		fallthrough
	case reflect.Int, reflect.Int32:
		res = typ.Int
	case reflect.Uint64:
		if isRef(t, refFlag) {
			cs := reflect.Zero(t).Interface().(MarkFlag).Flags()
			res = typ.Type{typ.KindBits, getConstInfo(t, typ.Consts(cs))}
			break
		}
		fallthrough
	case reflect.Uint, reflect.Uint32:
		res = typ.Int
	case reflect.Float32, reflect.Float64:
		res = typ.Real
	case reflect.String:
		if isRef(t, refEnum) {
			cs := reflect.Zero(t).Interface().(MarkEnum).Enums()
			res = typ.Type{typ.KindBits, getConstInfo(t, typ.Consts(cs))}
			break
		}
		res = typ.Str
	case reflect.Struct:
		if isRef(t, refTime) {
			res = typ.Time
			break
		}
		if isRef(t, refType) {
			res = typ.Typ
			break
		}
		if isRef(t, refDict.Elem()) {
			if !ptr {
				return typ.Void, typ.ErrInvalid
			}
			return typ.Keyer, nil
		}
		nfo, _, err := reflectFields(t, nfos)
		if err != nil {
			return typ.Void, err
		}
		k := typ.KindRec
		if tn := t.Name(); tn != "" {
			if c := tn[0]; c >= 'A' && c <= 'Z' {
				k = typ.KindObj
				nfo.Ref = t.String()
			}
		}
		res = typ.Type{Kind: k, Info: nfo}
	case reflect.Array:
		if isRef(t, refUUID) {
			res = typ.UUID
		}
	case reflect.Map:
		if !isRef(t.Key(), refStr) {
			return typ.Void, cor.Error("dict key must be string type")
		}
		et, err := reflectType(t.Elem(), nfos)
		if err != nil {
			return typ.Void, err
		}
		res = typ.Dict(et)
	case reflect.Slice:
		if isRef(t, refRaw) {
			res = typ.Raw
			break
		}
		if isRef(t, refList) {
			return typ.Idxer, nil
		}
		if t.Name() == "Dyn" && isRef(t, refEl) {
			res = typ.Dyn
			break
		}
		et, err := reflectType(t.Elem(), nfos)
		if err != nil {
			return typ.Void, err
		}
		res = typ.List(et)
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

func isRef(t reflect.Type, ref reflect.Type) bool {
	return t == ref || t.ConvertibleTo(ref)
}

func reflectFields(t reflect.Type, nfos infoMap) (*typ.Info, [][]int, error) {
	nfo := nfos[t]
	if nfo != nil {
		return nfo.Info, nfo.Idx, nil
	}
	nfo = &fields{Info: new(typ.Info)}
	nfos[t] = nfo
	fs := make([]typ.Param, 0, 16)
	idx := make([][]int, 0, 16)
	err := collectFields(t, nil, func(name, _ string, et reflect.Type, i []int) error {
		ft, err := reflectType(et, nfos)
		if err != nil {
			return err
		}
		fs = append(fs, typ.Param{name, ft})
		idx = append(idx, i)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	nfo.Params = fs
	nfo.Idx = idx
	return nfo.Info, idx, nil
}

type fidx struct {
	name string
	idx  []int
}

func fieldIndices(t reflect.Type, fs []typ.Param) ([][]int, error) {
	m := make(map[string]fidx, len(fs)+8)
	err := collectFields(t, nil, func(name, key string, _ reflect.Type, idx []int) error {
		m[key] = fidx{name, idx}
		return nil
	})
	if err != nil {
		return nil, err
	}
	res := make([][]int, 0, len(fs))
	for _, f := range fs {
		fi, ok := m[f.Key()]
		if !ok {
			return nil, cor.Errorf("field %s not found", f.Key())
		}
		res = append(res, fi.idx)
	}
	return res, nil
}

type fieldCollector = func(name, key string, t reflect.Type, idx []int) error

func collectFields(t reflect.Type, idx []int, col fieldCollector) error {
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Name != "" && !f.Anonymous {
			if c := f.Name[0]; c < 'A' || c > 'Z' {
				continue
			}
		}
		var key string
		var opt bool
		// check for a json struct tag first
		tag := strings.Split(f.Tag.Get("json"), ",")
		if len(tag) > 0 && tag[0] != "" {
			key = tag[0]
			if key == "-" { // skip ignored fields
				continue
			}
			// we found a key check if optional field
			for _, t := range tag[1:] {
				if opt = t == "omitempty"; opt {
					break
				}
			}
		}
		// collect embedded only if we have no key set by json tag explicitly
		if key == "" && f.Anonymous {
			et := f.Type
			if et.Kind() == reflect.Ptr {
				et = et.Elem()
			}
			err := collectFields(et, append(idx, i), col)
			if err != nil {
				return err
			}
			continue
		}
		name := f.Name
		// use simple capitalization if key does not match the lowercase name
		if key != "" && key != cor.LastKey(name) {
			name = cor.Cased(key)
		}
		if key == "" {
			key = cor.LastKey(name)
		}
		if opt { // append a question mark to optional fields
			name += "?"
		}
		err := col(name, key, f.Type, append(idx, i))
		if err != nil {
			return err
		}
	}
	return nil
}
