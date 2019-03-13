package lit

import (
	"reflect"
	"time"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var (
	ErrRequiresPtr = cor.StrError("requires non-nil pointer argument")
	ErrNotMap      = cor.StrError("proxy not a map")
	ErrNotSlice    = cor.StrError("proxy not a slice")
	ErrNotStruct   = cor.StrError("proxy not a struct")
)

func ErrAssign(src, dst typ.Type) error { return cor.Errorf("%q not assignable to %q", src, dst) }

// Assign assigns the value of l to the interface pointer value or returns an error
func AssignTo(l Lit, ptr interface{}) error {
	if a, ok := ptr.(Assignable); ok {
		return assignTo(l, a)
	}
	return AssignToValue(l, reflect.ValueOf(ptr))
}

// AssignTo assigns the value of l to the reflect pointer value or returns an error
func AssignToValue(l Lit, ptr reflect.Value) (err error) {
	if !ptr.IsValid() || ptr.Kind() != reflect.Ptr {
		return ErrRequiresPtr
	}
	pp, err := ProxyValue(ptr)
	if err != nil {
		return err
	}
	return assignTo(l, pp)
}

func assignTo(l Lit, p Assignable) error {
	l, err := Convert(l, p.Typ(), 0)
	if err != nil {
		return err
	}
	return p.Assign(l)
}

// Proxy returns an assignable literal for the pointer argument ptr or an error
func Proxy(ptr interface{}) (Assignable, error) {
	return ProxyValue(reflect.ValueOf(ptr))
}

// ProxyValue returns an assignable literal for the reflect value v or an error.
// Types convertible to the following types use an assignable adapter type:
//     bool, int64, float64, string, [16]byte, []byte, time.Time, List and *Dict
// The numeric types int, int32, uint, uint32, float32 all arr, map and obj types
// use a proxy variant using reflection.
func ProxyValue(ptr reflect.Value) (Assignable, error) {
	if ptr.Kind() != reflect.Ptr {
		return nil, ErrRequiresPtr
	}
	et := ptr.Type().Elem()
	// check for assignable primitives
	switch et.Kind() {
	case reflect.Bool:
		if v, ok := ptrRef(et, refBool, ptr); ok {
			return (*Bool)(v.Interface().(*bool)), nil
		}
	case reflect.Int64:
		if isRef(et, refSecs) {
			if v, ok := ptrRef(et, refSpan, ptr); ok {
				return (*Span)(v.Interface().(*time.Duration)), nil
			}
		}
		if v, ok := ptrRef(et, refInt, ptr); ok {
			return (*Int)(v.Interface().(*int64)), nil
		}
	case reflect.Float64:
		if v, ok := ptrRef(et, refReal, ptr); ok {
			return (*Real)(v.Interface().(*float64)), nil
		}
	case reflect.String:
		if v, ok := ptrRef(et, refStr, ptr); ok {
			return (*Str)(v.Interface().(*string)), nil
		}
	case reflect.Slice:
		if v, ok := ptrRef(et, refRaw, ptr); ok {
			return (*Raw)(v.Interface().(*[]byte)), nil
		}
		if v, ok := ptrRef(et, refList, ptr); ok {
			return v.Interface().(*List), nil
		}
	case reflect.Array:
		if v, ok := ptrRef(et, refUUID, ptr); ok {
			return (*UUID)(v.Interface().(*[16]byte)), nil
		}
	case reflect.Struct:
		if v, ok := ptrRef(et, refTime, ptr); ok {
			return (*Time)(v.Interface().(*time.Time)), nil
		}
		if v, ok := ptrRef(et, refType, ptr); ok {
			return typProxy{v.Interface().(*typ.Type)}, nil
		}
		if v, ok := toRef(ptr.Type(), refDict, ptr); ok {
			return v.Interface().(*Dict), nil
		}
	}
	// generic proxy fallback
	t, err := ReflectType(et)
	if err != nil {
		return nil, err
	}
	if t.Kind == typ.KindAny {
		return &anyProxy{ptr, Nil}, nil
	}
	p := proxy{t, ptr}
	switch t.Kind & typ.MaskBase {
	case typ.BaseNum:
		return &proxyNum{p}, nil
	case typ.BaseList:
		return &proxyArr{p}, nil
	case typ.BaseDict:
		return &proxyMap{p}, nil
	case typ.MaskCont:
		if et.Kind() == reflect.Ptr {
			et = et.Elem()
		}
		idx, err := fieldIndices(et, p.typ.Params)
		if err != nil {
			return nil, err
		}
		return &proxyObj{p, idx}, nil
	}
	return nil, cor.Errorf("cannot proxy type %s as %s", ptr.Type(), t)
}

type proxy struct {
	typ typ.Type
	val reflect.Value
}

func (p *proxy) Typ() typ.Type { return p.typ }
func (p *proxy) Ptr() interface{} {
	if p.val.IsValid() {
		return p.val.Interface()
	}
	return nil
}
func (p *proxy) el() reflect.Value {
	v := p.val
	if v.IsValid() && v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
func (p *proxy) elem(k reflect.Kind) (reflect.Value, bool) {
	pv := p.val
	if !pv.IsValid() || pv.Kind() != reflect.Ptr {
		return pv, false
	}
	v := pv.Elem()
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v, v.Kind() == k
}

func ptrRef(et reflect.Type, ref reflect.Type, v reflect.Value) (reflect.Value, bool) {
	if et == ref {
		return v, true
	}
	if et.ConvertibleTo(ref) {
		return v.Convert(reflect.PtrTo(ref)), true
	}
	return v, false
}

type typProxy struct {
	*typ.Type
}

func (p typProxy) Ptr() interface{} {
	return p.Type
}
func (p typProxy) Assign(l Lit) error {
	if t, ok := l.(typ.Type); ok {
		*p.Type = t
		return nil
	}
	return ErrAssign(l.Typ(), typ.Typ)
}

type anyProxy struct {
	val reflect.Value
	Lit
}

func (p *anyProxy) Ptr() interface{} {
	return p.val.Interface()
}
func (p *anyProxy) Assign(l Lit) error {
	p.Lit = l
	var v interface{}
	switch x := l.(type) {
	case Charer:
		v = x.Val()
	case Numer:
		v = x.Val()
	case Assignable:
		v = x.Ptr()
		p.val.Elem().Set(reflect.ValueOf(v).Elem())
		return nil
	default:
		v = x
	}
	p.val.Elem().Set(reflect.ValueOf(v))
	return nil
}
