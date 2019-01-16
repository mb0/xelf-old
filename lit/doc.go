/*
Package lit provides code for working with literals.

The Lit interface defines the common behaviour of all literals. There are additional interfaces for
each base type: Opter, Numer, Charer, Idxer or Keyer.

The go implementation provides adapter types that implement those interfaces. There are:

   concrete adapters for the base, numeric and character typed literals
   abstract adapters for uncommon, nested or custom types using reflection

Literal can be read from and written as JSON or a xelf extension with four new syntax features:

   single quoted strings:               'no need to escape a double quote _"_'
   raw multi-line strings:              `\r\a\w \s\t\r\i\n\g`
   optional commas (as in whitespace):  [1 2], {'a': 1 'b' :2}
   simple dict key notation:            {+foo 1 'with space': 23 "json key": 42}

Literals without context default to their base type. Base types in typed context are automatically
converted to the context. Type expressions can otherwise be used to convert literals to a specific
type. For example:

    (real 1)
    (int 1e6)
    (time "2018-11-16T23:52:20")
    (span "7:07:40")
    ((obj +id uuid +name str?) ["68986386-46ac-47f5-bf47-198ab20e594b" "foo"])

The type usually applies to the whole literal and defines all nested fields types. Which works for
all fields except any typed fields, those can not fully be represented in the JSON literal format
and need explicit conversion.

    (arr|@prod {})
    (arr|any (float 23) (span '7h'))

*/
package lit
