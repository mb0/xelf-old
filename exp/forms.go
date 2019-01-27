package exp

import "errors"

var (
	ErrArgCount  = errors.New("unexpected argument count")
	ErrRogueEl   = errors.New("unexpected element")
	ErrRogueTag  = errors.New("unexpected tag")
	ErrRogueDecl = errors.New("unexpected declaration")
	ErrRogueHead = errors.New("unexpected head element")
	ErrRogueTail = errors.New("unexpected tail element")
	ErrTailTag   = errors.New("unexpected tail tag")
	ErrTailDecl  = errors.New("unexpected tail declaration")
)

// Forms are helper functions to check expression arguments
// We distinguish between tag, declaration expressions and all other elements, that we call the
// plain elements in the context of forms.

// ArgsForm accepts only plain elements.
func ArgsForm(es []El) error {
	for _, e := range es {
		switch e.(type) {
		case Tag:
			return ErrRogueTag
		case Decl:
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

// ArgsMinMax accepts at least min and at most max plain elements.
func ArgsMinMax(es []El, min, max int) error {
	if len(es) < min || len(es) > max {
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
		switch v := e.(type) {
		case Decl:
			return nil, ErrRogueDecl
		case Tag:
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
		switch v := e.(type) {
		case Decl:
			return nil, nil, ErrRogueDecl
		case Tag:
			if tail {
				return nil, nil, ErrTailTag
			}
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

// DeclForm accepts only declarations.
// This is used for the let expression: (let +nice true)
// This is also the same syntax as xelf object type definitions.
func DeclForm(es []El) ([]Decl, error) {
	decls := make([]Decl, 0, len(es))
	for _, e := range es {
		switch v := e.(type) {
		case Decl:
			decls = append(decls, v)
		case Tag:
			return nil, ErrRogueTag
		default:
			return nil, ErrRogueEl
		}
	}
	return decls, nil
}

// DeclRestForm accepts leading declarations and then plain elements.
// This is used for the with expression: (with +say 'Hello' +to 'World' (log (say ' ' to '!'))
func DeclRestForm(es []El) (decl []Decl, list []El, _ error) {
	decl = make([]Decl, 0, len(es))
	var tail bool
	for i, e := range es {
		switch v := e.(type) {
		case Decl:
			if tail {
				return nil, nil, ErrTailDecl
			}
			decl = append(decl, v)
		case Tag:
			return nil, nil, ErrRogueTag
		default:
			tail = true
			if list == nil {
				list = es[i:]
			}
		}
	}
	return decl, list, nil
}

// ArgsDeclForm accepts leading plain elements and then declaration expressions.
// This is the same syntax as xelf flag, enum and record type definitions.
func ArgsDeclForm(es []El) ([]El, []Decl, error) {
	decls := make([]Decl, 0, len(es))
	var decl bool
	for i, e := range es {
		switch v := e.(type) {
		case Decl:
			if !decl {
				es = es[:i]
				decl = true
			}
			decls = append(decls, v)
		case Tag:
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
		switch v := e.(type) {
		case Decl:
			decl = true
			decls = append(decls, v)
		case Tag:
			if decl {
				return nil, nil, ErrTailTag
			}
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
		switch v := e.(type) {
		case Decl:
			decl = true
			decls = append(decls, v)
		case Tag:
			if decl {
				return tags, decls, es[i:], nil
			}
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
