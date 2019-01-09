package bfr

import "strings"

// Ctx is serialization context with output configuration flags
type Ctx struct {
	B
	JSON bool
}

// Writer is an interface for types that can write to a Ctx
type Writer interface {
	WriteBfr(Ctx) error
}

// String writes w and returns the result as string ignoring any error
func String(w Writer) string {
	var b strings.Builder
	_ = w.WriteBfr(Ctx{B: &b})
	return b.String()
}

// JSON writes w with the json flag and returns the result as bytes or an error
func JSON(w Writer) ([]byte, error) {
	b := Get()
	defer Put(b)
	err := w.WriteBfr(Ctx{B: b, JSON: true})
	return b.Bytes(), err
}
