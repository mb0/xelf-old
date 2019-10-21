package exp

import (
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

// Layout is a helper to validate form expression arguments.
//
// We distinguish between tag and declaration expressions, and all other elements, that we call the
// plain elements in the context of layouts.
//
// The layout is formalizes by the parameter signature. It uses a number of special parameter names,
// that indicate to the layout how arguments must be parsed. A full form consisting of all possible
// kinds of accepted elements, is: (form 'full' :args :decls :tail)
//
// These are the recognised parameter names:
//    plain accepts any number of plain elements and can be followed by tags or decls
//    tags accepts any number of tag expressions and can be followed by decls or tail
//    decls accepts declarations with multiple arguments and can only be followed by tail.
//    args or tail  accepts any number of leading plain elements and then tag expression
//    args is equavilent to plain follows by tags, tail must be the last element
//
// Explicit parameter arguments must be at the start befor any special parameter.
//
// The parameter types give hints at what types are accepted. The special parameters can only use
// container types. The decls parameters expect a keyer type, while all others accept an
// idxer type. If the type is omitted, the layout will not resolve or check that parameter.
//
type Layout struct {
	Sig    typ.Type
	Groups [][]El
}

func (l *Layout) All() (res []El) {
	for _, g := range l.Groups {
		res = append(res, g...)
	}
	return res
}

func (l *Layout) Count() (n int) {
	for _, g := range l.Groups {
		n += len(g)
	}
	return n
}

func (l *Layout) Args(idx int) []El {
	if l == nil || idx < 0 || idx >= len(l.Groups) {
		return nil
	}
	return l.Groups[idx]
}

func (l *Layout) Arg(idx int) El {
	args := l.Args(idx)
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

func (l *Layout) Tags(idx int) []*Named {
	args := l.Args(idx)
	res := make([]*Named, 0, len(args))
	for _, arg := range args {
		switch arg.Typ() {
		case typ.Named:
			res = append(res, arg.(*Named))
		default:
			res = append(res, &Named{El: arg})
		}
	}
	return res
}

func (l *Layout) Decls(idx int) ([]*Named, error) {
	args := l.Args(idx)
	res := make([]*Named, 0, len(args))
	for _, arg := range args {
		switch arg.Typ() {
		case typ.Named:
			res = append(res, arg.(*Named))
		default:
			return nil, cor.Errorf("unexpected decl element %s", arg)
		}
	}
	return res, nil
}

func (l *Layout) Unis(idx int) ([]*Named, error) {
	args := l.Args(idx)
	res := make([]*Named, 0, len(args))
	var naked int
	for _, arg := range args {
		switch arg.Typ() {
		case typ.Named:
			n := arg.(*Named)
			res = append(res, n)
			if n.El == nil {
				naked++
				continue
			}
			for naked > 0 {
				res[len(res)-naked-1].El = n.El
				naked--
			}
		default:
			if naked == 0 {
				return nil, cor.Errorf("unexpected uni element %T %s", arg, arg)
			}
			for naked > 0 {
				res[len(res)-naked].El = arg
				naked--
			}
		}
	}
	return res, nil
}

func (l *Layout) Resl(p *Prog, env Env, h typ.Type) error {
	if l == nil {
		return nil
	}
	var res error
	inst := typ.Type{l.Sig.Kind, &typ.Info{Params: make([]typ.Param, 0, len(l.Sig.Params))}}
	if l.Sig.HasRef() {
		inst.Ref = l.Sig.Ref
	}
	for i, param := range l.Sig.Params[:len(l.Sig.Params)-1] {
		if i >= len(l.Groups) {
			break
		}
		args := l.Groups[i]
		if len(args) == 0 {
			inst.Params = append(inst.Params, param)
			continue
		}
		switch key := param.Key(); key {
		case "plain", "tags", "tail", "args", "decls":
			v := p.New()
			p.Bind(v.Kind, typ.NewAlt(param.Type.Elem()))
			var err error
			args, err = p.ReslAll(env, args, v)
			if err != nil {
				if err != ErrUnres {
					return err
				}
				res = err
			}
			v = p.Apply(v)
			switch key {
			case "tags", "decls":
				param.Type = typ.Dict(v)
			default:
				param.Type = typ.List(v)
			}
			inst.Params = append(inst.Params, param)
			l.Groups[i] = args
		default: // explicit param
			v := p.New()
			el, err := p.Resl(env, args[0], v)
			if err != nil {
				if err != ErrUnres {
					return err
				}
				res = err
			}
			param.Type = v
			inst.Params = append(inst.Params, param)
			l.Groups[i] = []El{el}
		}
	}
	if h == typ.Void {
		h = p.New()
	}
	inst.Params = append(inst.Params, typ.Param{Type: h})
	r, err := typ.Unify(p.Ctx, l.Sig, inst)
	if err != nil {
		return cor.Errorf("unify sig %s with %s: %v",
			p.Apply(l.Sig), p.Apply(inst), err)
	}
	l.Sig = r
	return res
}

func (l *Layout) Eval(p *Prog, env Env, h typ.Type) error {
	if l == nil {
		return nil
	}
	for i, param := range l.Sig.Params[:len(l.Sig.Params)-1] {
		if i >= len(l.Groups) {
			break
		}
		args := l.Groups[i]
		if len(args) == 0 {
			continue
		}
		switch key := param.Key(); key {
		case "plain", "tags", "tail", "args", "decls":
			args, err := p.EvalAll(env, args, typ.Void)
			if err != nil {
				return err
			}
			l.Groups[i] = args
		default: // explicit param
			el, err := p.Eval(env, args[0], typ.Void)
			if err != nil {
				return err
			}
			l.Groups[i] = []El{el}
		}
	}
	return nil
}

func SigLayout(sig typ.Type, args []El) (*Layout, error) {
	switch sig.Kind {
	case typ.KindFunc:
		return FuncLayout(sig, args)
	case typ.KindForm:
		return FormLayout(sig, args)
	}
	return nil, cor.Errorf("unexpected layout sig %s", sig)
}

func FormLayout(sig typ.Type, args []El) (*Layout, error) {
	if !isSig(sig) {
		return nil, cor.Errorf("invalid signature %s", sig)
	}
	params := sig.Params[:len(sig.Params)-1]
	res := make([][]El, 0, len(params))
	var tmp []El
Loop:
	for _, p := range params {
		// check kind of parameter and consume matching args
		tmp = nil
		switch p.Key() {
		case "plain":
			tmp, args = consumePlain(args, tmp)
		case "tags":
			tmp, args = consumeTags(args, tmp)
		case "args", "tail":
			tmp, args = consumePlain(args, tmp)
			tmp, args = consumeTags(args, tmp)
		case "decls":
			tmp, args = consumeDecls(args, tmp)
		default: // explicit param
			if len(args) > 0 {
				if args[0] == nil {
					return nil, cor.Errorf("arg is null for %s", sig)
				}
				if _, _, _, ok := isSpecial(args[0], ":+-"); ok {
					if !p.Opt() {
						break Loop
					}
				} else {
					tmp, args = args[:1], args[1:]
				}
			}
		}
		if len(tmp) == 0 && !p.Opt() {
			return nil, cor.Errorf("missing argument for %s %s", p.Name, p)
		}
		res = append(res, tmp)
	}
	// at this point all arguments should have been consumed
	if len(args) > 0 {
		return nil, cor.Errorf("unexpected tail element %s", args[0])
	}
	return &Layout{Sig: sig, Groups: res}, nil
}

func isSig(t typ.Type) bool {
	return (t.Kind == typ.KindForm || t.Kind == typ.KindFunc) && t.HasParams()
}

func consumeArg(es []El) (El, []El) {
	if len(es) != 0 {
		e := es[0]
		if _, _, _, ok := isSpecial(e, ":+-"); !ok {
			return e, es[1:]
		}
	}
	return nil, es
}
func consumeTag(es []El) (El, []El) {
	if len(es) != 0 {
		if v, ok := es[0].(*Named); ok && v.IsTag() {
			return v, es[1:]
		}
	}
	return nil, es
}

func consumeDecl(es []El) (El, []El) {
	if len(es) == 0 {
		return nil, nil
	}
	e := es[0]
	if t, s, a, ok := isSpecial(e, "+-"); ok {
		d := &Named{Name: s}
		es = es[1:]
		var els []El
		if t == typ.Dyn || len(a) > 0 {
			els, a = consumePlain(a, els)
			els, a = consumeTags(a, els)
			els, _ = consumeDecls(a, els)
			if t == typ.Dyn {
				d.El = &Dyn{Els: els}
				return d, es
			}
		} else {
			els, es = consumePlain(es, els)
			els, es = consumeTags(es, els)
		}
		switch len(els) {
		case 0:
		case 1:
			d.El = els[0]
		default:
			d.El = &Dyn{Els: els}
		}
		return d, es
	}
	return nil, es
}

func consumePlain(es []El, res []El) ([]El, []El) {
	var e El
	for len(es) > 0 {
		e, es = consumeArg(es)
		if e != nil {
			res = append(res, e)
			continue
		}
		break
	}
	return res, es
}
func consumeTags(es []El, res []El) ([]El, []El) {
	var e El
	for len(es) > 0 {
		e, es = consumeTag(es)
		if e != nil {
			res = append(res, e)
			continue
		}
		break
	}
	return res, es
}
func consumeDecls(es []El, res []El) ([]El, []El) {
	var e El
	for len(es) > 0 {
		e, es = consumeDecl(es)
		if e != nil {
			res = append(res, e)
			continue
		}
		break
	}
	return res, es
}

func isSpecial(e El, pre string) (t typ.Type, s string, a []El, ok bool) {
	switch t = e.Typ(); t {
	case typ.Dyn:
		switch d := e.(type) {
		case *Dyn:
			if len(d.Els) != 0 {
				s, ok = isSpecialSym(d.Els[0], pre)
				a = d.Els[1:]
			}
		}
	case typ.Named:
		v := e.(*Named)
		d := v.Dyn()
		s, ok = v.Name, true
		if d != nil {
			t, a = typ.Dyn, d.Els
		} else {
			a = v.Args()
		}
	}
	return
}
func isSpecialSym(e El, pre string) (string, bool) {
	switch v := e.(type) {
	case *Sym:
		return isSpecialName(v.Name, pre)
	}
	return "", false
}
func isSpecialName(name string, pre string) (string, bool) {
	return name, name != "" && strings.IndexByte(pre, name[0]) != -1
}
