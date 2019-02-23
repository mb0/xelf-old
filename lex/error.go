package lex

import (
	"fmt"
	"strings"

	"github.com/mb0/xelf/cor"
	"golang.org/x/xerrors"
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
	Want  rune
	err   error
	frame xerrors.Frame
}

// Error builds and returns an error string of e.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d:%d: ", e.Line, e.Col))
	b.WriteString(e.err.Error())
	if e.Want != 0 {
		b.WriteString(" want token ")
		b.WriteString(TokStr(e.Want))
	}
	if e.Tok != 0 {
		b.WriteString(" got ")
		b.WriteString(TokStr(e.Tok))
	}
	return b.String()
}

func (e *Error) Format(f fmt.State, c rune) {
	xerrors.FormatError(e, f, c)
}

func (e *Error) FormatError(p xerrors.Printer) error {
	p.Print(e.Error())
	if p.Detail() {
		e.frame.Format(p)
	}
	return nil
}

func (e *Error) Unwrap() error {
	return e.err
}

func (*Error) Is(t error) bool {
	_, ok := t.(*Error)
	return ok
}

func ErrorAt(t Token, err error) error {
	return ErrorSkip(t, err, 0, 2)
}

func ErrorWant(t Token, err error, want rune) error {
	return ErrorSkip(t, err, want, 2)
}

func ErrorSkip(t Token, err error, want rune, skip int) error {
	return &Error{t, want, err, xerrors.Caller(skip)}
}
