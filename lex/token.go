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
	// Tag rune indicates a tag token.
	Tag
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
	case Tag:
		return "Tag"
	}
	return fmt.Sprintf("%q", r)
}

// Pos represents a file position by offset, line and column in bytes.
type Pos struct {
	Off  uint32
	Line uint16
	Col  uint16
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

// Src represents a file span with a start and end position
type Src struct {
	Pos
	End Pos
}

func (s Src) Source() Src { return s }

// Token represent a token recognized by the lexer with start offset and line position.
// The tok field hold either the special rune Number, String or Symbol or is itself the input rune.
// Special tokens also contain the read token input as string.
type Token struct {
	Tok rune
	Src
	Raw string
}

func (t Token) String() string {
	switch t.Tok {
	case EOF:
		return "EOF"
	case Number, String, Symbol:
		return t.Raw
	}
	return string(t.Tok)
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
func (t *Tree) WriteBfr(b *bfr.Ctx) (err error) {
	switch t.Tok {
	case Number, String, Symbol:
		_, err = b.WriteString(t.Raw)
	case EOF:
	case Tag:
		for i, c := range t.Seq {
			err = c.WriteBfr(b)
			if err != nil {
				break
			}
			if i == 0 {
				b.WriteString(t.Raw)
			}
		}
	default:
		b.WriteRune(t.Tok)
		end := closing(t.Tok)
		if end == 0 {
			break
		}
		for i, c := range t.Seq {
			if i > 0 {
				b.WriteByte(' ')
			}
			err = c.WriteBfr(b)
			if err != nil {
				return err
			}
		}
		_, err = b.WriteRune(end)
	}
	return err
}
