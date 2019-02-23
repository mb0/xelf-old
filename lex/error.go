package lex

import (
	"fmt"
	"strings"

	"github.com/mb0/xelf/cor"
)

var (
	// ErrUnexpected denotes an unexpected input rune.
	ErrUnexpected = cor.StrError("unexpected")
	// ErrUnterminated denotes an unterminated quote or open bracket.
	ErrUnterminated = cor.StrError("unterminated")
	// ErrExpectDigit denotes missing digits in a floating point format.
	ErrExpectDigit = cor.StrError("expect digit")
)

// Error is a special lexer error with token information.
type Error struct {
	Token
	Err  error
	Want rune
}

// Error builds and returns an error string of e.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(e.Unwrap().Error())
	if e.Want != 0 {
		b.WriteString(" want token ")
		b.WriteString(TokStr(e.Want))
	}
	if e.Tok != 0 {
		b.WriteString(" at: ")
		b.WriteString(e.Token.String())
		b.WriteString(fmt.Sprintf(" :%d:%d", e.Line, e.Col))
	}
	return b.String()
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	if e.Err == nil {
		return ErrUnexpected
	}
	return e.Err
}
