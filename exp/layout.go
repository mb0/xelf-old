package exp

import (
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
// kinds of accepted elements, is: (form 'full' +args +decls +tail)
//
// These are the recognised parameter names:
//    plain or rest accepts any number of plain elements.
//    tags accepts any number of tag expressions.
//    args or tail  accepts any number of leading plain elements and then tag expression.
//    unis accepts declarations with at most one argument.
//    decls accepts declarations with multiple arguments.
//
// With some caveats:
//    Each special parameter name can occur only once.
//    Plain and rest or args and tail cannot follow each other.
//    Plain or rest can not follow args or tail.
//    Only one of either unis or decls can occur.
//
// Explicit parameters are plain elements and can only occur in front or instead plain arguments.
//
// The parameter types give hints at what types are accepted. The special parameters can only use
// container types. The unis and decls parameters expect a keyer type, while all others accpect an
// idxer type. If the type is omitted, the layout will not resolve or check that parameter.
//
// TODO:
//    allow naked and unnamed fields for form signatures
//    add type validation when hints are provided
type Layout struct {
	ps   []typ.Param
	args [][]El
}

func (l *Layout) Args(idx int) []El {
	if l == nil || idx < 0 || idx >= len(l.args) {
		return nil
	}
	return l.args[idx]
}

func (l *Layout) Arg(idx int) El {
	args := l.Args(idx)
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

func (l *Layout) Tags(idx int) ([]Tag, error) {
	args := l.Args(idx)
	res := make([]Tag, 0, len(args))
	for i, arg := range args {
		switch arg.Typ() {
		case typ.Tag:
			res = append(res, arg.(Tag))
		case typ.Decl:
			return nil, cor.Errorf("unexpected decl %s", arg)
		default:
			res = append(res, Tag{Args: args[i : i+1]})
		}
	}
	return res, nil
}

func (l *Layout) Decls(idx int) ([]Decl, error) {
	args := l.Args(idx)
	res := make([]Decl, 0, len(args))
	for _, arg := range args {
		switch arg.Typ() {
		case typ.Decl:
			res = append(res, arg.(Decl))
		default:
			return nil, cor.Errorf("unexpected element %s", arg)
		}
	}
	return res, nil
}

func (l *Layout) Unis(idx int) ([]Decl, error) {
	args := l.Args(idx)
	var naked int
	res := make([]Decl, 0, len(args))
	for _, arg := range args {
		switch arg.Typ() {
		case typ.Decl:
			d := arg.(Decl)
			switch len(d.Args) {
			case 0:
				naked++
			case 1:
				for naked > 0 {
					res[len(res)-naked].Args = d.Args
					naked--
				}
			default:
				return nil, cor.Errorf("unexpected tail for %s", d)
			}
			res = append(res, d)
		default:
			return nil, cor.Errorf("unexpected element %s", arg)
		}
	}
	return res, nil
}

func (l *Layout) Resolve(c *Ctx, env Env) error {
	if l == nil {
		return nil
	}
	var res error
	for i, p := range l.ps {
		if p.Type == typ.Void {
			continue
		}
		args := l.args[i]
		if len(args) == 0 {
			continue
		}
		switch p.Name {
		case "plain", "rest",
			"tags", "tail", "args",
			"decls",
			"unis":
			args, err := c.ResolveAll(env, args, typ.Void)
			if err != nil {
				if err == ErrUnres {
					res = err
					continue
				}
				return err
			}
			if !c.Part {
				l.args[i] = args
			}
		default: // explicit param
			el, err := c.Resolve(env, args[0], p.Type)
			if err != nil {
				if err == ErrUnres {
					res = err
					continue
				}
				return err
			}
			cmp := typ.Compare(el.Typ(), p.Type)
			if cmp < typ.LvlCheck {
				return cor.Errorf("cannot convert %s to %s", el.Typ(), p.Type)
			}
			if c.Part {
				args[0] = el
			} else {
				l.args[i] = []El{el}
			}
		}
	}
	return res
}

func LayoutArgs(params []typ.Param, args []El) (*Layout, error) {
	if len(params) == 0 {
		if len(args) > 0 {
			return nil, cor.Error("unexpected argument count")
		}
		return nil, nil
	}
	res := make([][]El, 0, len(params))
	var tmp []El
Loop:
	for _, p := range params {
		// check kind of parameter and consume matching args
		tmp = nil
		switch p.Name {
		case "plain", "rest":
			tmp, args = consumePlain(args)
		case "tags", "tail", "args":
			tmp, args = consumeTags(args, p.Name != "tags")
		case "decls":
			tmp, args = consumeDecls(args)
		case "unis":
			tmp, args = consumeUnis(args)
		default: // explicit param
			if len(args) == 0 {
				if !p.Opt() {
					return nil, cor.Errorf("missing argument for %s", p)
				}
				continue
			}
			switch args[0].Typ() {
			case typ.Tag:
				if !p.Opt() {
					break Loop
				}
			case typ.Decl:
				if !p.Opt() {
					break Loop
				}
			default:
				tmp, args = args[:1], args[1:]
			}
		}
		res = append(res, tmp)
	}
	// at this point all arguments should have been consumed
	if len(args) > 0 {
		return nil, cor.Errorf("unexpected element %s", args[0])
	}
	return &Layout{params, res}, nil
}

func consumePlain(es []El) ([]El, []El) {
	for i, e := range es {
		switch e.Typ() {
		case typ.Tag:
		case typ.Decl:
		default:
			continue
		}
		return es[:i], es[i:]
	}
	return es, nil
}

func consumeTags(es []El, plain bool) ([]El, []El) {
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
		case typ.Tag:
			v := e.(Tag)
			if v.Name == "::" {
				rest := es[i+1:]
				if len(v.Args) > 0 {
					rest = make([]El, 0, len(v.Args)+len(rest))
					rest = append(rest, v.Args...)
					rest = append(rest, es[i+1:]...)
				}
				return es[:i], rest
			}
			plain = false
			continue
		default:
			if !plain {
				break
			}
			continue
		}
		return es[:i], es[i:]
	}
	return es, nil
}
func consumeDecls(es []El) ([]El, []El) {
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			continue
		}
		return es[:i], es[i:]
	}
	return es, nil
}
func consumeUnis(es []El) ([]El, []El) {
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			v := e.(Decl)
			if len(v.Args) > 1 {
				rest := v.Args[1:]
				v.Args = v.Args[:1]
				es[i] = v
				if len(es) > i+1 {
					if l := len(rest); l == 0 {
						rest = es[i+1:]
					} else {
						rest = append(rest[:l:l], es[i+1:]...)
					}
				}
				return es[:i+1], append(rest, es[i+1:]...)
			}
			continue
		}
		return es[:i], es[i:]
	}
	return es, nil
}
