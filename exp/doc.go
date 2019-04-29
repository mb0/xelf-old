/*
Package exp is a simple and extensible expression language, built on xelf types and literals.
It is meant to be used as a base for domain specific languages for a variety of tasks.

Language elements share the common interface El.
Atom elements are represented by a symbol, a literal or, for some types, a special type expression.
The atoms are:
    Type: a type definition as defined by package typ
    Lit:  a literal value as defined by package lit
    Sym:  a referenced and unresolved atom or spec

Expression elements are enclosed in parenthesis and usually start with a reference to a spec.
    Call:  an unresolved expression where the spec is known
    Dyn:   a unresolved, dynamic call where the spec is yet unknown
    Named: a tagged argument or declaration that applies to the parent expression

The parsing and resolution process is very abstract and uses following rules:
Literals and type symbols as well as type expressions are parsed normally.
Tag symbols associate to the next element, unless it is the end, another tag or declaration.
Declaration symbols start an implicit sub expression with all following elements until the end or
another declaration, except for the naked '-' sign, that never groups elements.
	(eq (e +a :x 1 :y :z + 2 +c - 4 5 6) (e (+a :x 1 :y :z) (+ 2) (+c) 4 5 6))

All other symbols are parsed as such. Expressions not starting with a symbol, are parsed as dynamic
expression.

The layout uses the spec signature to decide how to parse arguments.

Dynamic expressions starting with a literal or type are resolved as the 'dyn' expression. Languages
built on this package can choose to use the built-in std resolver or use a custom implementation.

Dynamic expressions starting with a tag or declaration are invalid. The only other allowed
start elements are another unresolved or dynamic expression.
*/
package exp
