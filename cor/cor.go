/*
Package cor has some utility function for working with numeric and character literal values.

The literal values of raw, uuid, time and span all have a string parser, formatter
and a parser function that returns a pointer or nil on error.

The primitive go types bool, int64, float64 and string have a function that returns a
pointer to the passed in value to aid code generation for literal values.
*/
package cor

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Int returns a pointer to v.
func Int(v int64) *int64 { return &v }

// Real returns a pointer to v.
func Real(v float64) *float64 { return &v }

// Str returns a pointer to v.
func Str(v string) *string { return &v }

// Any returns a pointer to v.
func Any(v interface{}) *interface{} { return &v }
