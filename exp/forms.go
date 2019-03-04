package exp

import "github.com/mb0/xelf/typ"

var (
	layoutArgs = []typ.Param{{Name: "args"}}
	layoutNode = []typ.Param{{Name: "tags"}, {Name: "rest"}}
	layoutUnis = []typ.Param{{Name: "unis"}}
	layoutFull = []typ.Param{{Name: "args"}, {Name: "decls"}, {Name: "tail"}}
)

// TagsForm accepts leading plain elements and then tag expressions after that.
// This is similar to how most languages use named function arguments.
// (form '_' +args -)
func TagsForm(es []El) ([]Tag, error) {
	lo, err := LayoutArgs(layoutArgs, es)
	if err != nil {
		return nil, err
	}
	return lo.Tags(0)
}

// NodeForm accepts optional starting tag expressions and then plain elements.
// This is similar to xml node syntax with tags and arguments and the elements as children.
// (form '_' +tags +rest -)
func NodeForm(es []El) ([]Tag, []El, error) {
	lo, err := LayoutArgs(layoutNode, es)
	if err != nil {
		return nil, nil, err
	}
	tags, err := lo.Tags(0)
	return tags, lo.Args(1), err
}

// UniDeclForm accepts only declarations with one argument.
// Naked declaration use the next declaration's argument.
// This is used for the let expression: (let +x +y 0)
// This is also the same syntax as xelf object type definitions.
// (form '_' +unis -)
func UniDeclForm(es []El) ([]Decl, error) {
	lo, err := LayoutArgs(layoutUnis, es)
	if err != nil {
		return nil, err
	}
	return lo.Unis(0)
}

// FullForm accepts leading plain elements, then tag, then declarations and then more elements.
// (form '_' +args +decls +tail -)
func FullForm(es []El) ([]Tag, []Decl, []El, error) {
	lo, err := LayoutArgs(layoutFull, es)
	if err != nil {
		return nil, nil, nil, err
	}
	tags, err := lo.Tags(0)
	if err != nil {
		return nil, nil, nil, err
	}
	decls, err := lo.Decls(1)
	if err != nil {
		return nil, nil, nil, err
	}
	tail := lo.Args(2)
	return tags, decls, tail, nil
}
