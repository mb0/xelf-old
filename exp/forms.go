package exp

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/typ"
)

var (
	ErrArgCount  = cor.StrError("unexpected argument count")
	ErrRogueEl   = cor.StrError("unexpected element")
	ErrRogueTag  = cor.StrError("unexpected tag")
	ErrRogueDecl = cor.StrError("unexpected declaration")
	ErrRogueHead = cor.StrError("unexpected head element")
	ErrRogueTail = cor.StrError("unexpected tail element")
	ErrTailTag   = cor.StrError("unexpected tail tag")
	ErrTailDecl  = cor.StrError("unexpected tail declaration")
	ErrNakedDecl = cor.StrError("unexpected naked declaration")
)

// Forms are helper functions to check expression arguments
// We distinguish between tag, declaration expressions and all other elements, that we call the
// plain elements in the context of forms.

// ArgsForm accepts only plain elements.
func ArgsForm(es []El) error {
	for _, e := range es {
		switch e.Typ() {
		case typ.Tag:
			return ErrRogueTag
		case typ.Decl:
			return ErrRogueDecl
		}
	}
	return nil
}

// ArgsMin accepts at least min plain elemts.
func ArgsMin(es []El, min int) error {
	if len(es) < min {
		return ErrArgCount
	}
	return ArgsForm(es)
}

// ArgsExact accepts exactly n plain elements.
func ArgsExact(es []El, n int) error {
	if len(es) != n {
		return ErrArgCount
	}
	return ArgsForm(es)
}

// TagsForm accepts leading plain elements and then tag expressions after that.
// This is similar to how most languages use named function arguments.
func TagsForm(es []El) ([]Tag, error) {
	var tag bool
	tags := make([]Tag, 0, len(es))
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			return nil, ErrRogueDecl
		case typ.Tag:
			v := e.(Tag)
			if v.Name == "::" && len(v.Args) > 0 {
				return nil, ErrRogueTail
			}
			tag = true
			tags = append(tags, v)
		default:
			if tag {
				return nil, ErrRogueTail
			}
			tags = append(tags, Tag{Args: es[i : i+1]})
		}
	}
	return tags, nil
}

// NodeForm accepts optional starting tag expressions and then plain elements.
// This is similar to xml node syntax with tags and arguments and the elements as children.
func NodeForm(es []El) (tags []Tag, list []El, _ error) {
	var tail bool
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			return nil, nil, ErrRogueDecl
		case typ.Tag:
			if tail {
				return nil, nil, ErrTailTag
			}
			v := e.(Tag)
			if v.Name == "::" {
				if i != len(es)-1 {
					return nil, nil, ErrRogueTail
				}
				return tags, v.Args, nil
			}
			tags = append(tags, v)
		default:
			tail = true
			if list == nil {
				list = es[i:]
			}
		}
	}
	return tags, list, nil
}

// UniDeclForm accepts only declarations with one argument.
// Naked declaration use the next declaration's argument.
// This is used for the let expression: (let +x +y 0)
// This is also the same syntax as xelf object type definitions.
func UniDeclForm(es []El) ([]Decl, error) {
	var naked int
	decls := make([]Decl, 0, len(es))
	for _, e := range es {
		switch e.Typ() {
		case typ.Decl:
			v := e.(Decl)
			switch len(v.Args) {
			case 0:
				naked++
				decls = append(decls, v)
			case 1:
				for naked > 0 {
					decls[len(decls)-naked].Args = v.Args
					naked--
				}
				decls = append(decls, v)
			default:
				return nil, ErrRogueTail
			}
		case typ.Tag:
			return nil, ErrRogueTag
		default:
			return nil, ErrRogueEl
		}
	}
	if naked > 0 {
		return nil, ErrNakedDecl
	}
	return decls, nil
}

// UniDeclRest accepts leading declarations with one argument and then plain elements.
// Naked declaration use the next declaration's argument.
// This is used for the with expression: (with +say 'Hello' +to 'World' (log (say ' ' to '!'))
func UniDeclRest(es []El) (decls []Decl, list []El, _ error) {
	var naked int
	decls = make([]Decl, 0, len(es))
	var tail bool
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			if tail {
				return nil, nil, ErrTailDecl
			}
			v := e.(Decl)
			switch len(v.Args) {
			case 0:
				naked++
				decls = append(decls, v)
			case 1:
				for naked > 0 {
					decls[len(decls)-naked].Args = v.Args
					naked--
				}
				decls = append(decls, v)
			default:
				if i != len(es)-1 {
					return nil, nil, ErrRogueTail
				}
				list = v.Args[1:]
				v.Args = v.Args[:1]
				for naked > 0 {
					decls[len(decls)-naked].Args = v.Args
					naked--
				}
				decls = append(decls, v)
			}
		case typ.Tag:
			return nil, nil, ErrRogueTag
		default:
			tail = true
			if list == nil {
				list = es[i:]
			}
		}
	}
	if naked > 0 {
		return nil, nil, ErrNakedDecl
	}
	return decls, list, nil
}

// ArgsDeclForm accepts leading plain elements and then declaration expressions.
// This is the same syntax as xelf flag, enum and record type definitions.
func ArgsDeclForm(es []El) ([]El, []Decl, error) {
	decls := make([]Decl, 0, len(es))
	var decl bool
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			if !decl {
				es = es[:i]
				decl = true
			}
			v := e.(Decl)
			decls = append(decls, v)
		case typ.Tag:
			return nil, nil, ErrRogueTag
		default:
			if decl {
				return nil, nil, ErrRogueTail
			}
		}
	}
	return es, decls, nil
}

// TagsDeclForm accepts leading plain elements, then tags and finally declarations.
func TagsDeclForm(es []El) ([]Tag, []Decl, error) {
	tags := make([]Tag, 0, len(es))
	decls := make([]Decl, 0, len(es))
	var tag, decl bool
	for i, e := range es {
		switch e.Typ() {
		case typ.Decl:
			decl = true
			decls = append(decls, e.(Decl))
		case typ.Tag:
			if decl {
				return nil, nil, ErrTailTag
			}
			v := e.(Tag)
			if v.Name == "::" && len(v.Args) > 0 {
				return nil, nil, ErrRogueTail
			}
			tag = true
			tags = append(tags, v)
		default:
			if tag || decl {
				return nil, nil, ErrRogueTail
			}
			tags = append(tags, Tag{Args: es[i : i+1]})
		}
	}
	return nil, nil, nil
}

// FullForm accepts leading plain elements, then tag, then declarations and then more elements.
func FullForm(es []El) ([]Tag, []Decl, []El, error) {
	tags := make([]Tag, 0, len(es))
	decls := make([]Decl, 0, len(es))
	var tag, decl, end bool
	for i, e := range es {
		if end {
			return nil, nil, nil, ErrRogueTail
		}
		switch e.Typ() {
		case typ.Decl:
			decl = true
			decls = append(decls, e.(Decl))
		case typ.Tag:
			if decl {
				return tags, decls, es[i:], nil
			}
			v := e.(Tag)
			if v.Name == "::" {
				if i != len(es)-1 {
					return nil, nil, nil, ErrRogueTail
				}
				return tags, decls, v.Args, nil
			}
			tag = true
			tags = append(tags, v)
		default:
			if tag || decl {
				return tags, decls, es[i:], nil
			}
			tags = append(tags, Tag{Args: es[i : i+1]})
		}
	}
	return tags, decls, nil, nil
}
