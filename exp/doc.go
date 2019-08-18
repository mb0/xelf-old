/*
Package exp is a simple and extensible expression language, built on xelf types and literals.
It is meant to be used as a base for domain specific languages for a variety of tasks.

Language elements implement the interface El and are pointer of the following types:
    Atom  for literals, including type literals
    Sym   for unresolved symbols
    Dyn   for unresolved expressions
    Call  for expressions with a resolved form or func specification
    Named for special tag and decl syntax elements used in call arguments

Parse reads a tree and returns atoms, symbols, named or dynamic expressions.
Literals and type symbols as well as type expressions are parsed as atoms.
Tag symbols associate to the next element, unless at the end, next to another tag or declaration.
Declaration symbols start an implicit sub expression with all following elements until the end or
another declaration, except for the naked '-' sign, that never groups elements.
	(eq (e +a :x 1 :y :z + 2 +c - 4 5 6) (e (+a :x 1 :y :z) (+ 2) (+c) 4 5 6))

Prog is a program with a type context used to resolve or evaluate elements.

Resl resolves symbol types and dyn expressions to calls or atoms.

The resolution uses the spec signature and is automatic for all function specs and most form specs.

Dynamic expressions are resolved to call expressions. Dynamic expressions starting with a spec
are use that spec directly, starting with a literal or type lookup a builtin spac. Languages
extensions can change the dynamic lookup in the program context.

Dynamic expressions starting with a tag or declaration are invalid. The only other allowed
start elements are other unresolved expressions.

Eval evaluates elements resulting in an atom or partially resolved element.
*/
package exp
