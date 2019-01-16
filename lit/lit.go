package lit

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// Lit is the common interface for all literal adapters
// A nil Lit represents an absent value.
type Lit interface {
	// Typ returns the defined type of the literal
	Typ() typ.Type
	// IsZero returns whether the literal value is the zero value
	IsZero() bool
	// WriteBfr writes to a bfr ctx either as strict JSON or xelf representation
	WriteBfr(bfr.Ctx) error
	// String returns the xelf representation as string
	String() string
	// MarshalJSON returns the JSON representation as bytes
	MarshalJSON() ([]byte, error)
}

// Opter represents a literal with an optional type
type Opter interface {
	Lit
	// Some returns the wrapped literal or nil
	Some() Lit
}

// Numer represents a literal with a numeric type
type Numer interface {
	Lit
	// Num returns the numeric value of the literal as float64
	Num() float64
	// Val returns the simple go value representing this literal
	// The type is either bool, int64, float64, time.Time or time.Duration
	Val() interface{}
}

// Charer represents a literal with a character type
type Charer interface {
	Lit
	// Char returns the character format of the literal as string
	Char() string
	// Val returns the simple go value representing this literal
	// The type is either string, []byte, [16]byte, time.Time or time.Duration
	Val() interface{}
}

// Idxer represents a containter literal with elements accessible by index
type Idxer interface {
	Lit
	// Len returns the number of contained elements
	Len() int
	// Idx returns the literal of the element at idx or an error
	Idx(idx int) (Lit, error)
	// SetIdx sets the element value at idx to l or returns an error
	SetIdx(idx int, l Lit) error
	// IterIdx iterates over elements, calling iter with the elements index and literal value
	// If iter returns false the iteration is aborted
	IterIdx(f func(int, Lit) bool) error
}

// Keyer represents a containter literal with elements accessible by key
type Keyer interface {
	Lit
	// Len returns the number of contained elements
	Len() int
	// Keys returns a string slice of all keys
	Keys() []string
	// Key returns the literal of the element with key k or an error
	Key(k string) (Lit, error)
	// SetKey sets the elements value with key k to l or returns an error
	SetKey(k string, l Lit) error
	// IterKey iterates over elements, calling iter with the elements key and literal value
	// If iter returns false the iteration is aborted
	IterKey(iter func(string, Lit) bool) error
}
