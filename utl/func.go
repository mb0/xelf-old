package utl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

// FuncResolver is a resolver wrapping a reflected go function.
type FuncResolver struct {
	val   reflect.Value
	param typ.Type
	ptyps []reflect.Type
	res   typ.Type
	err   bool
}

func (r *FuncResolver) Resolve(c *exp.Ctx, env exp.Env, e exp.El) (exp.El, error) {
	// we only resolve function when used as expression
	x, ok := e.(*exp.Expr)
	if !ok {
		return e, exp.ErrUnres
	}
	// valided tags form
	tags, err := exp.TagsForm(x.Args)
	if err != nil {
		return nil, err
	}
	// prime args with parameter reflect types
	args := make([]reflect.Value, len(r.ptyps))
	for i, pt := range r.ptyps {
		args[i] = reflect.New(pt).Elem()
	}
	// collect args from tags
	for i, tag := range tags {
		// find index
		idx := i
		if tag.Name != "" { // not positional arg
			key := strings.ToLower(tag.Name[1:])
			_, idx, err = r.param.FieldByKey(key)
			if err != nil {
				return nil, err
			}
		}
		if idx >= len(args) {
			return nil, cor.Errorf("unexpected argument index for %v", x.Name)
		}
		// resolve tag arg
		el, err := c.Resolve(env, tag.Args[0])
		if err != nil {
			return nil, err
		}
		l, ok := el.(lit.Lit)
		if !ok {
			return nil, cor.Errorf("expect literal got %T", el)
		}
		// assign into arg list
		val := args[idx].Addr()
		err = lit.AssignToValue(l, val)
		if err != nil {
			return nil, cor.Errorf("%v have %s", err, val)
		}
	}
	// get reflect values from argument
	// call reflect function with value
	res := r.val.Call(args)
	if r.err { // check last result
		last := res[len(res)-1]
		if !last.IsNil() {
			return nil, last.Interface().(error)
		}
	}
	if r.res == typ.Void { // nothing to return
		return nil, nil
	}
	// create a proxy from the result and return
	return lit.AdaptValue(res[0])
}

var refErr = reflect.TypeOf((*error)(nil)).Elem()

// ReflectFunc will reflect val and return a function resolver or an error.
// The names are optionally and associated to the arguments by index.
// If no name is defined, it uses a uppercase p and the index as name: 'P0'.
func ReflectFunc(val interface{}, names ...string) (*FuncResolver, error) {
	f := FuncResolver{val: reflect.ValueOf(val)}
	if f.val.Kind() != reflect.Func {
		return nil, cor.Errorf("expect function argument got %T", val)
	}
	t := f.val.Type()
	if t.IsVariadic() {
		return nil, cor.Error("variadic fuctions are not yet supported")
	}
	n := t.NumIn()
	fs := make([]typ.Field, 0, n)
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
		if name == "" {
			name = fmt.Sprintf("P%d", i)
		}
		pt = append(pt, rt)
		fs = append(fs, typ.Field{Name: name, Type: xt})
	}
	f.ptyps = pt
	f.param = typ.Obj(fs)
	n = t.NumOut()
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
			return nil, cor.Error("expect at most one compatible result and an optional error")
		}
		xt, err := lit.ReflectType(rt)
		if err != nil {
			return nil, err
		}
		f.res = xt
	}
	return &f, nil
}
