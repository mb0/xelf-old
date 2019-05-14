// Package prx provides literal proxies that write though to the underlying go data structure.
package prx

import (
	"reflect"
	"time"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

var (
	ErrRequiresPtr = cor.StrError("requires non-nil pointer argument")
	ErrNotMap      = cor.StrError("proxy not a map")
	ErrNotSlice    = cor.StrError("proxy not a slice")
	ErrNotStruct   = cor.StrError("proxy not a struct")
)

// Assign assigns the value of l to the interface pointer value or returns an error
func AssignTo(l lit.Lit, ptr interface{}) error {
	if a, ok := ptr.(lit.Assignable); ok {
		return assignTo(l, a)
	}
	return AssignToValue(l, reflect.ValueOf(ptr))
}

// AssignTo assigns the value of l to the reflect pointer value or returns an error
func AssignToValue(l lit.Lit, ptr reflect.Value) (err error) {
	if !ptr.IsValid() || ptr.Kind() != reflect.Ptr {
		return ErrRequiresPtr
	}
	pp, err := ProxyValue(ptr)
	if err != nil {
		return err
	}
	return assignTo(l, pp)
}

func assignTo(l lit.Lit, p lit.Assignable) error {
	l, err := lit.Convert(l, p.Typ(), 0)
	if err != nil {
		return err
	}
	return p.Assign(l)
}

// Proxy returns an assignable literal for the pointer argument ptr or an error
func Proxy(ptr interface{}) (lit.Assignable, error) {
	return ProxyValue(reflect.ValueOf(ptr))
}

// ProxyValue returns an assignable literal for the reflect value v or an error.
// Types convertible to the following types use an assignable adapter type:
//     bool, int64, float64, string, [16]byte, []byte, time.Time, List and *Dict
// The numeric types int, int32, uint, uint32, float32 all list, dict and record types
// use a proxy variant using reflection.
func ProxyValue(ptr reflect.Value) (lit.Assignable, error) {
	if ptr.Kind() != reflect.Ptr {
		return nil, ErrRequiresPtr
	}
	et := ptr.Type().Elem()
	// check for assignable primitives
	switch et.Kind() {
	case reflect.Bool:
		if v, ok := ptrRef(et, refBool, ptr); ok {
			return (*lit.Bool)(v.Interface().(*bool)), nil
		}
	case reflect.Int64:
		if isRef(et, refSecs) {
			if v, ok := ptrRef(et, refSpan, ptr); ok {
				return (*lit.Span)(v.Interface().(*time.Duration)), nil
			}
		}
		if v, ok := ptrRef(et, refInt, ptr); ok {
			return (*lit.Int)(v.Interface().(*int64)), nil
		}
	case reflect.Float64:
		if v, ok := ptrRef(et, refReal, ptr); ok {
			return (*lit.Real)(v.Interface().(*float64)), nil
		}
	case reflect.String:
		if v, ok := ptrRef(et, refStr, ptr); ok {
			return (*lit.Str)(v.Interface().(*string)), nil
		}
	case reflect.Slice:
		if v, ok := ptrRef(et, refRaw, ptr); ok {
			return (*lit.Raw)(v.Interface().(*[]byte)), nil
		}
		if v, ok := ptrRef(et, refList, ptr); ok {
			return v.Interface().(*lit.List), nil
		}
	case reflect.Array:
		if v, ok := ptrRef(et, refUUID, ptr); ok {
			return (*lit.UUID)(v.Interface().(*[16]byte)), nil
		}
	case reflect.Struct:
		if v, ok := ptrRef(et, refTime, ptr); ok {
			return (*lit.Time)(v.Interface().(*time.Time)), nil
		}
		if v, ok := ptrRef(et, refType, ptr); ok {
			return lit.TypProxy{v.Interface().(*typ.Type)}, nil
		}
		if v, ok := toRef(ptr.Type(), refDict, ptr); ok {
			return v.Interface().(*lit.Dict), nil
		}
	case reflect.Ptr:
		if v, ok := ptrRef(et, refDict, ptr); ok {
			dptr := v.Interface().(**lit.Dict)
			if *dptr == nil {
				*dptr = &lit.Dict{}
			}
			return *dptr, nil
		}
	}
	// generic proxy fallback
	t, err := ReflectType(et)
	if err != nil {
		return nil, err
	}
	if t.Kind == typ.KindAny {
		return &lit.AnyProxy{ptr, lit.Nil}, nil
	}
	p := proxy{t, ptr}
	switch t.Kind & typ.KindAny {
	case typ.KindNum:
		return &proxyNum{p}, nil
	case typ.KindIdxr:
		return &proxyList{p}, nil
	case typ.KindKeyr:
		return &proxyDict{p}, nil
	case typ.KindCont:
		if et.Kind() == reflect.Ptr {
			et = et.Elem()
		}
		idx, err := fieldIndices(et, p.typ.Params)
		if err != nil {
			return nil, err
		}
		return &proxyRec{p, idx}, nil
	}
	return nil, cor.Errorf("cannot proxy type %s as %s", ptr.Type(), t)
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
