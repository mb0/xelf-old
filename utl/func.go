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

func (f *ReflectBody) ResolveFunc(c *exp.Ctx, env exp.Env, x *exp.Expr, h typ.Type) (exp.El, error) {
	lo, err := exp.FuncArgs(x)
	if err != nil {
		return nil, err
	}
	// resolve args
	err = lo.Resolve(c, env)
	if err != nil {
		return x, err
	}
	if !c.Exec {
		return x, exp.ErrExec
	}
	args := make([]reflect.Value, len(f.ptyps))
	for i, pt := range f.ptyps {
		v := reflect.New(pt)
		args[i] = v.Elem()
		n := lo.Args(i)
		if len(n) == 0 {
			// reflect already provides a zero value
			continue
		}
		err = lit.AssignToValue(n[0].(lit.Lit), v)
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
func ReflectFunc(name string, val interface{}, names ...string) (*exp.Func, error) {
	f := ReflectBody{val: reflect.ValueOf(val)}
	if f.val.Kind() != reflect.Func {
		return nil, cor.Errorf("expect function argument got %T", val)
	}
	t := f.val.Type()
	if t.IsVariadic() {
		return nil, cor.Error("variadic fuctions are not yet supported")
	}
	n := t.NumIn()
	fs := make([]typ.Param, 0, n+1)
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
		fs = append(fs, typ.Param{Name: name, Type: xt})
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
	fs = append(fs, typ.Param{Type: res})
	return &exp.Func{exp.FuncSig(name, fs), &f}, nil
}
