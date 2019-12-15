/*
Package exp is a simple and extensible expression language, built on xelf types and literals.
It is meant to be used as a base for domain specific languages for a variety of tasks.

Language elements implement the interface El and are pointer of the following types:
    Atom  for literals, including type literals
    Sym   for unresolved symbols
    Dyn   for unresolved expressions
    Call  for expressions with a resolved form or func specification
    Named for special tag syntax elements used in call arguments

Parse reads a tree and returns atoms, symbols, named or dynamic expressions.
Literals and type symbols as well as type expressions are parsed as atoms.
Tag symbols associate to the neighoring elements.

Prog is a program with a type context used to resolve or evaluate elements.

Resl resolves symbol types and dyn expressions to calls or atoms.

The resolution uses the spec signature and is automatic for all function specs and most form specs.

Dynamic expressions are resolved to call expressions. Dynamic expressions starting with a spec
use that spec directly, starting with a literal or type lookup a built-in spec. Language
extensions can change the dynamic lookup in the program context.

Eval evaluates elements resulting in an atom or partially resolved element.
*/
package exp
