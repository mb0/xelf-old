// Package lex provides a token and tree lexer, tree splitter and string quoting code.
package lex

import (
	"bufio"
	"io"
	"sort"
	"strings"

	"github.com/mb0/xelf/cor"
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
// If simple is true, symbols can only be ascii names.
func (l *Lexer) Lex(simple bool) (Token, error) {
	r := l.next()
	for cor.Space(r) {
		r = l.next()
	}
	switch r {
	case EOF:
		t := l.tok(r)
		return t, ErrorAt(t, l.err)
	case ',', ';', '(', ')', '[', ']', '{', '}':
		return l.tok(r), nil
	case '"', '\'', '`':
		return l.lexString()
	}
	if cor.NameStart(r) {
		return l.lexSymbol(simple)
	}
	if cor.Digit(r) || r == '-' && cor.Digit(l.nxt) {
		return l.lexNumber()
	}
	if cor.Punct(r) {
		if simple {
			return l.tok(r), nil
		}
		return l.lexSymbol(false)
	}
	t := l.tok(r)
	return t, ErrorAt(t, ErrUnexpected)
}

// Scan scans and returns the next tree or an error.
// Symbols nested in sequences with curly or square brackets are read as simple names.
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

// pos returns a new pos at the current offset.
func (l *Lexer) pos() Pos {
	n, c := 1, l.idx
	if i := sort.SearchInts(l.lines, l.idx); i > 0 {
		n += i
		c -= l.lines[i-1]
	}
	return Pos{uint32(l.idx), uint16(n), uint16(c)}
}

// tok returns a new token at the current offset.
func (l *Lexer) tok(r rune) Token {
	p := l.pos()
	return Token{Tok: r, Src: Src{Pos: p, End: p.add(1)}}
}

// tokval returns a new value token at the current offset.
func (l *Lexer) val(t Token, val string) (Token, error) {
	t.Raw, t.End = val, l.pos()
	return t, nil
}

// lexString reads and returns a string token starting at the current offset.
func (l *Lexer) lexString() (Token, error) {
	t := l.tok(String)
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
		t, _ = l.val(t, b.String())
		return t, ErrorWant(t, ErrUnterminated, q)
	}
	b.WriteRune(q)
	return l.val(t, b.String())
}

// lexSymbol reads and returns a symbol token starting at the current offset.
// If simple is true, it only accepts ascii letters and digits.
func (l *Lexer) lexSymbol(simple bool) (t Token, _ error) {
	var b strings.Builder
	switch l.cur {
	case ':':
		t = l.tok(Tag)
	case '+', '-':
		t = l.tok(Decl)
	default:
		t = l.tok(Symbol)
	}
	b.WriteRune(l.cur)
	for cor.NamePart(l.nxt) || !simple && cor.Punct(l.nxt) {
		b.WriteRune(l.next())
	}
	return l.val(t, b.String())
}

// lexNumber reads and returns a number token starting at the current offset.
func (l *Lexer) lexNumber() (Token, error) {
	var b strings.Builder
	t := l.tok(Number)
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
			t, _ = l.val(t, b.String())
			return t, ErrorAtPos(l.pos(), ErrExpectDigit)
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
			t, _ = l.val(t, b.String())
			return t, ErrorAtPos(l.pos(), ErrExpectDigit)
		}
	}
	return l.val(t, b.String())
}

// lexDigits reads the next digits and writes the to b.
// It returns false if no digit was read.
func (l *Lexer) lexDigits(b *strings.Builder) bool {
	if !cor.Digit(l.nxt) {
		return false
	}
	for ok := true; ok; ok = cor.Digit(l.nxt) {
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
	simple := end != ')'
	t, err := l.Lex(simple)
	if err != nil {
		return res, err
	}
	for t.Tok != end && t.Tok != EOF {
		a, err := l.scanTree(t)
		if err != nil {
			return a, err
		}
		res.Seq = append(res.Seq, a)
		t, err = l.Lex(simple)
		if err != nil {
			return res, err
		}
	}
	res.End = t.Pos
	if t.Tok != end {
		return res, ErrorWant(t, ErrUnterminated, end)
	}
	return res, nil
}
