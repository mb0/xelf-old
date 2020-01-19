// Package lex provides a token and tree lexer, tree splitter and string quoting code.
package lex

import (
	"bufio"
	"io"
	"sort"
	"strings"

	"github.com/mb0/xelf/cor"
)

// Read returns a Tree read from r or an error.
func Read(r io.Reader) (*Tree, error) {
	return New(r).Tree()
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

// Token reads and returns the next token or an error.
func (l *Lexer) Token() (Token, error) {
	r := l.next()
	for cor.Space(r) {
		r = l.next()
	}
	switch r {
	case EOF:
		t := l.tok(r)
		return t, ErrorAt(t, l.err)
	case ':', ';', ',', '(', ')', '[', ']', '{', '}', '<', '>':
		return l.tok(r), nil
	case '"', '\'', '`':
		return l.lexString()
	}
	if cor.Digit(r) || r == '-' && cor.Digit(l.nxt) {
		return l.lexNumber()
	}
	if cor.NameStart(r) || cor.Punct(r) {
		return l.lexSymbol()
	}
	t := l.tok(r)
	return t, ErrorAt(t, ErrUnexpected)
}

// Tree scans and returns the next tree or an error.
func (l *Lexer) Tree() (*Tree, error) {
	t, err := l.Token()
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
	t.Raw, t.End = val, l.pos().add(1)
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
func (l *Lexer) lexSymbol() (t Token, _ error) {
	var b strings.Builder
	t = l.tok(Symbol)
	b.WriteRune(l.cur)
	for cor.NamePart(l.nxt) || cor.Punct(l.nxt) {
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
	} else if cor.Digit(l.nxt) {
		return t, cor.Errorf("number zero must be separated by whitespace")
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
// If the token is an open paren, trees are scanned until a matching closing paren.
// Enclosed trees are separated by white-spaces or comma.
// Colons and semicolons associate with neighboring trees to form tags.
// Tags are a way to structure tree elements as named values.
// 	{"name":"value"}
// We allow tags starting with symbols in xelf object literals:
// 	{name:"value"}
// In call expressions we allow short tags, where the leading symbol is only a name without value:
// 	(spec name;) equals (spec name:,)
// We use capitalization of the tag name for declaration tags as convention.
// 	(schema faq help:'faq schema has sub models'
// 		Cat:(ID:uuid Name:str
//			help:'this is a model with a help tag and capitalized field tags'
// 		)
// 		Help:(ID:uuid Name: str
//			help:'this is a model with the same name as a schema field but capitalized'
//		)
// 	)
func (l *Lexer) scanTree(t Token) (*Tree, error) {
	res := &Tree{Token: t}
	end := end(t.Tok)
	if end == 0 {
		return res, nil
	}
	t, err := l.Token()
	if err != nil {
		return res, err
	}
	for t.Tok != end && t.Tok != EOF {
		switch t.Tok {
		case ':', ';', ',':
			return res, ErrorAt(t, ErrUnexpected)
		}
		a, err := l.scanTree(t)
		if err != nil {
			return a, err
		}
		t, err = l.Token()
		if err != nil {
			return res, err
		}
		switch t.Tok {
		case ':':
			switch a.Tok {
			case Symbol, String:
			default:
				return res, ErrorWant(a.Token, ErrUnexpected, Symbol)
			}
			tt := &Tree{Token: Token{Tok: Tag, Raw: ":", Src: t.Src}}
			tt.Pos = a.Pos
			res.Seq = append(res.Seq, tt)
			t, err = l.Token()
			if err != nil {
				return res, err
			}
			switch t.Tok {
			case Number, String, Symbol, '[', '{', '<', '(':
				b, err := l.scanTree(t)
				if err != nil {
					return res, err
				}
				tt.Seq = []*Tree{a, b}
				tt.End = b.End
				t, err = l.Token()
				if err != nil {
					return res, err
				}
			default:
				tt.Seq = []*Tree{a}
			}
		case ';':
			switch a.Tok {
			case Symbol:
			default:
				return res, ErrorWant(a.Token, ErrUnexpected, Symbol)
			}
			tt := &Tree{Token: Token{Tok: Tag, Raw: ";", Src: t.Src}, Seq: []*Tree{a}}
			tt.Pos = a.Pos
			res.Seq = append(res.Seq, tt)
			t, err = l.Token()
			if err != nil {
				return res, err
			}
			continue
		default:
			res.Seq = append(res.Seq, a)
		}
		switch t.Tok {
		case ',':
			t, err = l.Token()
			if err != nil {
				return res, err
			}
		}
	}
	res.End = t.End
	if t.Tok != end {
		return res, ErrorWant(t, ErrUnterminated, end)
	}
	return res, nil
}

func end(start rune) rune {
	switch start {
	case '(':
		return ')'
	case '<':
		return '>'
	case '{':
		return '}'
	case '[':
		return ']'
	}
	return 0
}
