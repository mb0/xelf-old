// Package bfr provides a common interface for buffered writers and bytes.Buffer pool
package bfr

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{New: func() interface{} {
	return &bytes.Buffer{}
}}

// Get returns a bytes.Buffer from the pool
func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

// Put a Buffer back into the pool
func Put(b *bytes.Buffer) {
	b.Reset()
	pool.Put(b)
}

// Grow grows the buffer by n if it implements a Grow(int) method
// both bytes.Buffer and strings.Builder implement that method
func Grow(b B, n int) {
	if v, ok := b.(interface{ Grow(int) }); ok {
		v.Grow(n)
	}
}

// B is the common interface of bytes.Buffer, strings.Builder and bufio.Writer
type B interface {
	Write([]byte) (int, error)
	WriteByte(byte) error
	WriteRune(rune) (int, error)
	WriteString(string) (int, error)
}
