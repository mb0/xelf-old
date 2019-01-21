/*
Package exp is a simple and extensible expression language, built on xelf types and literals.
It is meant to be used as a base for domain specific languages for a variety of tasks.

Language elements share the common interface El.
Atom elements are represented by a symbol, a literal or, for some types, a special type expression.
The atoms are:
    Type: a type definition as defined by package typ
    Lit:  a literal value as defined by package lit
    Ref:  a referenced and unresolved atom or spec

Expression elements are enclosed in parenthesis and usually start with a reference to a spec.
    Expr: an unresolved expression where the spec name is known
    Dyn:  a unresolved, dynamic expression where the spec name is yet unknown
    Tag:  a tagged argument group that applies to the parent expression
    Decl: a declaration group that applies to the parent expression

The parsing and resolution process is very abstract and uses following rules:
Tag and declaration symbols are always treated as sub expression.
    (eq (e :a 1 :b +d :c 3 - :f) (e (:a 1) (:b) (+d (:c 3)) (:f)))
The parser reads literal and type symbols as such and otherwise looks up a resolver by name.
Expressions starting with a type are rewritten to the 'as'-expression
    (eq (str 'test') (as str 'test'))
Expressions starting with a literal are rewritten to the 'combine'-expression
    (eq ('a' 'b' 'c') (combine 'a' 'b' 'c'))

Languages built on this package can choose to use the built-in std resolvers or use custom
implementations.

*/
package exp
