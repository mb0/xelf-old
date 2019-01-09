package lex

import (
	"fmt"
)

const (
	_ rune = -iota
	EOF
	Num
	Str
	Sym
)

// TokStr returns a string representation of token rune t
func TokStr(r rune) string {
	switch r {
	case EOF:
		return "EOF"
	case Num:
		return "Num"
	case Str:
		return "Str"
	case Sym:
		return "Sym"
	}
	return fmt.Sprintf("%q", r)
}

// Pos represents a file position by line number and column in bytes
type Pos struct {
	Line, Col int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

// Token represent a token recognized by the lexer with start offset and line position.
// The tok field hold either the special rune Num, Str or Sym or is itself the input rune.
// Special tokens also contain the read token input as string.
type Token struct {
	Tok  rune
	Off  int
	Line int
	Col  int
	Val  string
}

func (t Token) String() string {
	s := TokStr(t.Tok)
	if len(t.Val) == 0 {
		return s
	}
	if total := len(s) + len(t.Val); total > 30 {
		return s + ": " + t.Val[:30-len(s)] + "…"
	}
	return s + ": " + t.Val
}

// Tree represents either a single token or a sequence of trees starting with an open bracket.
// Trees contain the tokens file positions. A sequence tree token and position refers to the
// opening bracket and the end tree to its matching closing bracket.
type Tree struct {
	Token
	Seq []*Tree
	End *Token
}

// Err wraps and returns the given err as a token error with position information.
func (t *Tree) Err(err error) error {
	return &Error{t.Token, err, 0}
}
