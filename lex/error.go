package lex

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrUnexpected denotes an unexpected input rune
	ErrUnexpected = errors.New("unexpected")
	// ErrUnterminated denotes an unterminated quote or open bracket
	ErrUnterminated = errors.New("unterminated")
	// ErrExpectDigit denotes missing digits in a floating point format
	ErrExpectDigit = errors.New("expect digit")
)

// Error is a special lexer error with token information
type Error struct {
	Token
	Err  error
	Want rune
}

// Error builds and returns an error string of e
func (e *Error) Error() string {
	var b strings.Builder
	if e.Err != nil {
		b.WriteString(e.Err.Error())
	} else {
		b.WriteString("unexpected token")
	}
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
