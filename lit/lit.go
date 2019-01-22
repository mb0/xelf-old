package lit

import (
	"errors"

	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/typ"
)

// BreakIter is a special error value that can be returned from iterators.
// It indicates that the iteration should be stopped even though no actual failure occurred.
var BreakIter = errors.New("break iter")

// Lit is the common interface for all literal adapters.
// A nil Lit represents an absent value.
type Lit interface {
	// Typ returns the defined type of the literal.
	Typ() typ.Type
	// IsZero returns whether the literal value is the zero value.
	IsZero() bool
	// WriteBfr writes to a bfr ctx either as strict JSON or xelf representation.
	WriteBfr(bfr.Ctx) error
	// String returns the xelf representation as string.
	String() string
	// MarshalJSON returns the JSON representation as bytes.
	MarshalJSON() ([]byte, error)
}

// Opter is the interface for literals with an optional type.
type Opter interface {
	Lit
	// Some returns the wrapped literal or nil
	Some() Lit
}

// Numer is the common interface for numeric literals.
type Numer interface {
	Lit
	// Num returns the numeric value of the literal as float64.
	Num() float64
	// Val returns the simple go value representing this literal.
	// The type is either bool, int64, float64, time.Time or time.Duration
	Val() interface{}
}

// Charer is the common interface for character literals.
type Charer interface {
	Lit
	// Char returns the character format of the literal as string.
	Char() string
	// Val returns the simple go value representing this literal.
	// The type is either string, []byte, [16]byte, time.Time or time.Duration.
	Val() interface{}
}

// Idxer is the common interface for container literals with elements accessible by index.
type Idxer interface {
	Lit
	// Len returns the number of contained elements.
	Len() int
	// Idx returns the literal of the element at idx or an error.
	Idx(idx int) (Lit, error)
	// SetIdx sets the element value at idx to l or returns an error.
	SetIdx(idx int, l Lit) error
	// IterIdx iterates over elements, calling iter with the elements index and literal value.
	// If iter returns an error the iteration is aborted.
	IterIdx(iter func(int, Lit) error) error
}

// Keyer is the common interface for container literals with elements accessible by key.
type Keyer interface {
	Lit
	// Len returns the number of contained elements.
	Len() int
	// Keys returns a string slice of all keys.
	Keys() []string
	// Key returns the literal of the element with key key or an error.
	Key(key string) (Lit, error)
	// SetKey sets the elements value with key key to l or returns an error.
	SetKey(key string, l Lit) error
	// IterKey iterates over elements, calling iter with the elements key and literal value.
	// If iter returns an error the iteration is aborted.
	IterKey(iter func(string, Lit) error) error
}

// Arr is the interface for arr literals.
type Arr interface {
	Idxer
	// Elem returns the arr element type.
	Elem() typ.Type
}

// Map is the interface for map literals.
type Map interface {
	Keyer
	// Elem returns the map element type.
	Elem() typ.Type
}

// Obj is the interface for obj literals.
type Obj interface {
	Lit
	// Len returns the number of fields.
	Len() int
	// Keys returns a string slice of all field keys.
	Keys() []string
	// Key returns the literal of the field with key key or an error.
	Key(key string) (Lit, error)
	// SetKey sets the fields value with key key to l or returns an error.
	SetKey(key string, l Lit) error
	// Idx returns the literal of the field at idx or an error.
	Idx(idx int) (Lit, error)
	// SetIdx sets the field value at idx to l or returns an error.
	SetIdx(idx int, l Lit) error
	// IterKey iterates over fields, calling iter with the fields key and literal value.
	// If iter returns an error the iteration is aborted
	IterKey(iter func(string, Lit) error) error
	// IterIdx iterates over fields, calling iter with the fields index and literal value.
	// If iter returns an error the iteration is aborted.
	IterIdx(iter func(int, Lit) error) error
}
