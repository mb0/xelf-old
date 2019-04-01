package bfr

import (
	"fmt"
	"strings"

	"github.com/mb0/xelf/lex"
)

// Ctx is serialization context with output configuration flags
type Ctx struct {
	B
	JSON  bool
	Depth int
	Tab   string
}

// Fmt writes the formatted string to the buffer or returns an error
func (c *Ctx) Fmt(f string, args ...interface{}) (err error) {
	if len(args) > 0 {
		_, err = fmt.Fprintf(c.B, f, args...)
	} else {
		_, err = c.WriteString(f)
	}
	return err
}

func (c *Ctx) Indent() bool {
	c.Depth++
	return c.Break()
}

func (c *Ctx) Dedent() bool {
	c.Depth--
	return c.Break()
}

func (c *Ctx) Break() bool {
	if c.Tab == "" {
		return false
	}
	c.WriteByte('\n')
	for i := c.Depth; i > 0; i-- {
		c.WriteString(c.Tab)
	}
	return true
}

// Quote writes the v as quoted string to the buffer or returns an error.
// The quote used is depending on the json context flag.
func (c *Ctx) Quote(v string) (err error) {
	if c.JSON {
		v, err = lex.Quote(v, '"')
	} else {
		v, err = lex.Quote(v, '\'')
	}
	if err != nil {
		return err
	}
	return c.Fmt(v)
}

// Writer is an interface for types that can write to a Ctx
type Writer interface {
	WriteBfr(*Ctx) error
}

// String writes w and returns the result as string ignoring any error
func String(w Writer) string {
	var b strings.Builder
	_ = w.WriteBfr(&Ctx{B: &b})
	return b.String()
}

// JSON writes w with the json flag and returns the result as bytes or an error
func JSON(w Writer) ([]byte, error) {
	b := Get()
	defer Put(b)
	err := w.WriteBfr(&Ctx{B: b, JSON: true})
	if err != nil {
		return nil, err
	}
	buf := b.Bytes()
	res := make([]byte, len(buf))
	copy(res, buf)
	return res, nil
}
