// Package lex provides a token and tree lexer, tree slitter and string quoting code.
package lex

import (
	"bufio"
	"io"
	"sort"
	"strings"
)

// Scan returns a Tree scanned from s or an error.
func Scan(s string) (*Tree, error) {
	return New(strings.NewReader(s)).Scan()
}

// Lexer is simple token lexer.
type Lexer struct {
	src      io.RuneScanner
	cur, nxt rune
	idx, nxn int
	err      error
	lines    []int
}

// New returns a new Lexer for Reader r.
func New(r io.Reader) *Lexer {
	l := &Lexer{idx: -1}
	if src, ok := r.(io.RuneScanner); ok {
		l.src = src
	} else {
		l.src = bufio.NewReader(r)
	}
	l.next()
	return l
}

// Lex reads and returns the next token or an error.
// If punct is true symbols can only be simple names optionally prefixed by a plus or minus sign.
func (l *Lexer) Lex(punct bool) (Token, error) {
	r := l.next()
	for IsSpace(r) {
		r = l.next()
	}
	switch r {
	case EOF:
		t := l.tok(r, "")
		return t, &Error{t, l.err, 0}
	case ',', '(', ')', '[', ']', '{', '}':
		return l.tok(r, ""), nil
	case '"', '\'', '`':
		return l.lexChar()
	}
	if IsLetter(r) || r == '+' {
		return l.lexSym(punct)
	}
	if IsDigit(r) || r == '-' && IsDigit(l.nxt) {
		return l.lexNum()
	}
	if IsPunct(r) {
		if punct {
			return l.tok(r, ""), nil
		}
		return l.lexSym(punct)
	}
	t := l.tok(r, "")
	return t, &Error{t, ErrUnexpected, 0}
}

// Scan scans and returns the next tree or an error.
// Symbols nested in sequences with curly or square brackets are read with the punct set to true.
func (l *Lexer) Scan() (*Tree, error) {
	t, err := l.Lex(false)
	if err != nil {
		return nil, err
	}
	return l.scanTree(t)
}

// next proceeds to and returns the next rune, updating the look-ahead.
func (l *Lexer) next() rune {
	if l.err != nil {
		return EOF
	}
	l.cur = l.nxt
	l.idx += l.nxn
	l.nxt, l.nxn, l.err = l.src.ReadRune()
	if l.cur == '\n' {
		l.lines = append(l.lines, l.idx)
	}
	return l.cur
}

// tok returns a new token at the current offset.
func (l *Lexer) tok(r rune, val string) Token {
	if r == 0 {
		r = l.cur
	}
	i := sort.SearchInts(l.lines, l.idx)
	n, c := 1, l.idx
	if i > 0 {
		n += i
		c -= l.lines[i-1]
	}
	return Token{r, l.idx, n, c, val}
}

// lexChar reads and returns a char token starting at the current offset.
func (l *Lexer) lexChar() (Token, error) {
	q := l.cur
	var b strings.Builder
	b.WriteRune(q)
	c := l.next()
	var esc bool
	for c != EOF && c != q || esc {
		esc = !esc && c == '\\' && q != '`'
		b.WriteRune(c)
		c = l.next()
	}
	if c == EOF {
		t := l.tok(Str, b.String())
		return t, &Error{t, ErrUnterminated, q}
	}
	b.WriteRune(q)
	return l.tok(Str, b.String()), nil
}

// lexSym reads and returns a sym token starting at the current offset.
func (l *Lexer) lexSym(punct bool) (Token, error) {
	var b strings.Builder
	t := l.tok(Sym, "")
	b.WriteRune(l.cur)
	for IsNamePart(l.nxt) || !punct && IsPunct(l.nxt) {
		b.WriteRune(l.next())
	}
	t.Val = b.String()
	return t, nil
}

// lexNum reads and returns a num token starting at the current offset.
func (l *Lexer) lexNum() (Token, error) {
	var b strings.Builder
	t := l.tok(Num, "")
	if l.cur == '-' {
		b.WriteRune(l.cur)
		l.next()
	}
	b.WriteRune(l.cur)
	if l.cur != '0' {
		l.lexDigits(&b)
	}
	if l.nxt == '.' {
		b.WriteRune(l.nxt)
		l.next()
		if ok := l.lexDigits(&b); !ok {
			l.next()
			return t, &Error{t, ErrExpectDigit, 0}
		}
	}
	if l.nxt == 'e' || l.nxt == 'E' {
		b.WriteRune('e')
		l.next()
		if l.nxt == '+' || l.nxt == '-' {
			b.WriteRune(l.nxt)
			l.next()
		}
		if ok := l.lexDigits(&b); !ok {
			l.next()
			return t, &Error{t, ErrExpectDigit, 0}
		}
	}
	t.Val = b.String()
	return t, nil
}

// lexDigits reads the next digits and writes the to b.
// It returns false if no digit was read.
func (l *Lexer) lexDigits(b *strings.Builder) bool {
	if !IsDigit(l.nxt) {
		return false
	}
	for ok := true; ok; ok = IsDigit(l.nxt) {
		b.WriteRune(l.nxt)
		l.next()
	}
	return true
}

// scanTree returns a token tree constructed from t or an error.
// If the token is an open paren trees are scanned until a matching closing paren.
// Symbols are read with punctuation only inside round parenthesis.
func (l *Lexer) scanTree(t Token) (*Tree, error) {
	res := &Tree{Token: t}
	var end rune
	switch t.Tok {
	case '(':
		end = ')'
	case '{':
		end = '}'
	case '[':
		end = ']'
	default:
		return res, nil
	}
	t, err := l.Lex(end != ')')
	if err != nil {
		return nil, err
	}
	for t.Tok != end && t.Tok != EOF {
		a, err := l.scanTree(t)
		if err != nil {
			return a, err
		}
		res.Seq = append(res.Seq, a)
		t, err = l.Lex(end != ')')
		if err != nil {
			return nil, err
		}
	}
	if t.Tok != end {
		return res, &Error{t, ErrUnterminated, end}
	}
	res.End = &t
	return res, nil
}
