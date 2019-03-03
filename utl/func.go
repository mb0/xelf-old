package utl

import (
	"reflect"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// ReflectBody is a funcition resolver wrapping a reflected go function.
type ReflectBody struct {
	val   reflect.Value
	ptyps []reflect.Type
	err   bool
}

func (f *ReflectBody) ResolveCall(c *exp.Ctx, env exp.Env, fc *exp.Call, h typ.Type) (exp.El, error) {
	args := make([]reflect.Value, len(f.ptyps))
	for i, pt := range f.ptyps {
		v := reflect.New(pt)
		args[i] = v.Elem()
		n := fc.Args[i]
		if len(n.Args) == 0 {
			// reflect already provides a zero value
			continue
		}
		// resolve tag arg
		l, err := c.Resolve(env, n.Args[0], typ.Void)
		if err != nil {
			return fc.Expr, err
		}
		err = lit.AssignToValue(l.(lit.Lit), v)
		if err != nil {
			return nil, cor.Errorf("have %s: %w", v, err)
		}
	}
	// get reflect values from argument
	// call reflect function with value
	res := f.val.Call(args)
	if f.err { // check last result
		last := res[len(res)-1]
		if !last.IsNil() {
			return nil, last.Interface().(error)
		}
		res = res[:len(res)-1]
	}
	if len(res) == 0 { // nothing to return
		return nil, nil
	}
	// create a proxy from the result and return
	return lit.AdaptValue(res[0])
}

var refErr = reflect.TypeOf((*error)(nil)).Elem()

// ReflectFunc reflects val and returns a function literal or an error.
// The names are optionally and associated to the arguments by index.
func ReflectFunc(val interface{}, names ...string) (*exp.Func, error) {
	f := ReflectBody{val: reflect.ValueOf(val)}
	if f.val.Kind() != reflect.Func {
		return nil, cor.Errorf("expect function argument got %T", val)
	}
	t := f.val.Type()
	if t.IsVariadic() {
		return nil, cor.Error("variadic fuctions are not yet supported")
	}
	n := t.NumIn()
	fs := make([]typ.Field, 0, n+1)
	pt := make([]reflect.Type, 0, n)
	for i := 0; i < n; i++ {
		rt := t.In(i)
		xt, err := lit.ReflectType(rt)
		if err != nil {
			return nil, err
		}
		var name string
		if i < len(names) {
			name = names[i]
		}
		pt = append(pt, rt)
		fs = append(fs, typ.Field{Name: name, Type: xt})
	}
	f.ptyps = pt
	n = t.NumOut()
	var res typ.Type
	for i := 0; i < n; i++ {
		rt := t.Out(i)
		if rt.ConvertibleTo(refErr) {
			f.err = true
			if i+1 < n {
				return nil, cor.Error("error can only be last result")
			}
			break
		}
		if i > 0 {
			return nil, cor.Error("expect at most one compatible result and optionally an error")
		}
		xt, err := lit.ReflectType(rt)
		if err != nil {
			return nil, err
		}
		res = xt
	}
	fs = append(fs, typ.Field{Type: res})
	return &exp.Func{exp.AnonSig(fs), &f}, nil
}
