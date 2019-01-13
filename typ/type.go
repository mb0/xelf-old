package typ

// Type represents the full type details. It consists of a kind and additional information.
type Type struct {
	Kind `json:"typ"`
	*Info
}

// Info represents the reference name and obj fields for the ref and obj types.
type Info struct {
	Ref    string  `json:"ref,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

// Field represents an obj field with a name and type.
type Field struct {
	Name string `json:"name,omitempty"`
	Type
}
