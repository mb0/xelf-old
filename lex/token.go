package lex

import (
	"fmt"

	"github.com/mb0/xelf/bfr"
)

// Special token runes. If not in this list the token rune represent an input rune.
const (
	_   rune = -iota
	EOF      // EOF rune indicates the end of file or another error.
	// Number rune indicates a number literal token.
	Number
	// String rune indicates a string literal token.
	String
	// Symbol rune indicates a identifier symbol token.
	Symbol

	Tag
	Decl
)

// TokStr returns a string representation of token rune t.
func TokStr(r rune) string {
	switch r {
	case EOF:
		return "EOF"
	case Number:
		return "Number"
	case String:
		return "String"
	case Symbol:
		return "Symbol"
	}
	return fmt.Sprintf("%q", r)
}

// Pos represents a file position by line number and column in bytes.
type Pos struct {
	Off  uint32
	Line uint16
	Col  uint16
}

func Offset(o int) Pos {
	return Pos{uint32(o), 1, uint16(o)}
}

func (p Pos) String() string {
	if p.Line > 0 {
		return fmt.Sprintf("%d:%d", p.Line, p.Col)
	}
	return fmt.Sprintf("0:%d", p.Off)
}

func (p Pos) add(col uint16) Pos {
	p.Off += uint32(col)
	p.Col += uint16(col)
	return p
}

type Src struct {
	Pos
	End Pos
}

// Token represent a token recognized by the lexer with start offset and line position.
// The tok field hold either the special rune Number, String or Symbol or is itself the input rune.
// Special tokens also contain the read token input as string.
type Token struct {
	Tok rune
	Src
	Raw string
}

func (t Token) String() string {
	s := TokStr(t.Tok)
	if len(t.Raw) == 0 {
		return s
	}
	if total := len(s) + len(t.Raw); total > 30 {
		return s + ": " + t.Raw[:30-len(s)] + "…"
	}
	return s + ": " + t.Raw
}

// Tree represents either a single token or a sequence of trees starting with an open bracket.
// Trees contain the tokens file positions. A sequence tree token and position refers to the
// opening bracket and the end tree to its matching closing bracket.
type Tree struct {
	Token
	Seq []*Tree
}

// Err wraps and returns the given err as a token error with position information.
func (t *Tree) Err(err error) error {
	return ErrorSkip(t.Token, err, 0, 2)
}

func (t *Tree) String() string { return bfr.String(t) }
func (t *Tree) WriteBfr(b *bfr.Ctx) error {
	switch t.Tok {
	case Number, String, Symbol, Tag, Decl:
		b.WriteString(t.Raw)
	case '(':
		b.WriteByte('(')
		for i, c := range t.Seq {
			if i > 0 {
				b.WriteByte(' ')
			}
			err := c.WriteBfr(b)
			if err != nil {
				return err
			}
		}
		return b.WriteByte(')')
	}
	return nil
}
