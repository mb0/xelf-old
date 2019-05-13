Design
======

This document discusses and sometimes explains the design trade-offs chosen for xelf. Unanswered
questions about the design of xelf can be raised with the author and will be incorporated into
this document. This document along with each package documentation acts as informal and incomplete
specification as long as the project is a work in progress.

The main mantra for xelf is simplicity and practicality for use with web-centric technologies.

Simplicity is a difficult goal by itself, but especially for a generic and extensible tool
like xelf. The project obviously has to provide value for common use-cases to be feasible,
which means code that provides that value. It should try to separate each feature as much as
possible and is allowed to be opinionated in feature selection.

Xelf should try to make gradual implementation easy where complexity cannot be avoided.

JSON
----

One aspect of xelf is to provide a unified way to work with data across different environments.
It needs to choose the data format that poses the least resistance.

The most basic environment to think of here is a some-what idiomatic sqlite3 target.

JSON emerged quite early and organically as the most commonly implemented and used object format.
It is very basic, well-known, and programmers are already used to map their data to a JSON
representation one way or another. Recently more and more database solution got JSON support.

Accepting plain JSON as valid literal syntax makes it very simple to interoperate with existing
code, databases, and user assumptions.

Building xelf around JSON at its core is very practical, but also provides limitations that need
to be accepted or worked around.

Types
-----

Xelf must have type information to work with data. Using JSON as literal syntax implies a number of
base types:

   any  for a 'null' without context
   bool for 'true' and 'false'
   num  for JSON numbers
   char for JSON strings
   idxr for JSON arrays
   keyr for JSON objects

However we usually need more specific type information that can not be represented in JSON.
The type system chose a number of specific primitive types. The specific types of num are int, real
and flag for bit sets, the specific types of char are str, raw for bytes, uuid, enum, time and span
for durations. The bool type is also considered a numeric type, because some environments might not
have a dedicated bool type, indexedDB in browsers comes to mind. Both span and time are usually
represented in a text format but can also be converted to an integer, representing milliseconds.
The numeric value of a time are the milliseconds from the unix epoch.

The selection is based on what types character and numeric literals are commonly used for, that
differentiate enough in comparison or manipulation behavior. It is heavily informed by types
supported by full featured databases like postgresql. We do not bother with different number sizes
or text length because it only complicates what xelf tries to provide and is only a concern to a
validation or storage layer. For most uses a couple of wasted bytes is not an issue.

There are two distinct behavior of container types called idxr and keyr. An indexer provides access
to its elements by index, and a keyer by a string key. The idxr and keyr base type can hold any
elements, while the list and dict type take an optional element type. They use different types to
avoid mix-ups with the record types rec and obj that do implement both indexer and keyer interface,
because record fields are inherently ordered. Idxr should be implemented to preserve field order for
this reason, but does not provide the idxr interface. Another reason in go to avoid maps most
literals in xelf is that they do not support interior pointers.

Apart from the any type all primitive and record types can be optional. Optional type variants have
a question mark suffix and translate to pointer, option or nullable types in the target
environments. Note that there are also optional record fields that mark the field itself and not its
value optional.

To rely on type information without specifying it explicitly in every case, for possibly large
composite literals, xelf must allow to explicitly refer to and infer from existing types. This is
covered by type embedding and type references. Recursive record declaration also use special
type references to itself or an ancestor. Type variables and alternatives are implicitly used in the
type inference and resolution phase.

The flag, enum and obj types are called schema types and are global references to a type definition
with additional information. The flag type is a integer type used as bit set with associated
constants. The enum type is a string type with associated constants. A object is a record type with
a known schema and possibly additional information attached to it.

The void, typ and exp types are special types and are useful in a language context. The exp types
provide a type signature for form, func types. They are otherwise used as an universal indicator to
differentiate between all language elements. This allows us to replace some type switches and
assertions to use a simple interface call instead.

The final naming of the types took a long time, some changes and serious consideration. Real does
not imply a floating point precision as much as float does and fits the short naming schema. Dict
implies an ordered list if of keyed entries, while map implies a key is mapped to a kind of value.
List is a sequence of values and array is more associated with a specific element type. A record has
is a structure of fields, while an object implies more 'real-world' properties.

To efficiently work with types most of the information is encoded in the 'kind' bit set. This bit
set is very compact and can quickly answer questions in varying granularity. For example:
An encoder is only interested in the base type, while the resolver only wants to check if its a
reference type. The reference name and field information are stored in a companion object.

Literals
--------

Xelf literal parsing rules are a superset of JSON with minor additional rules to address specific
issues, mainly to make writing literals by hand easier.

Char literals can be single quoted escaped literals. Xelf is more often than not handled as string
inside other environments. Single quoted string alleviate escaping quotes in double quoted strings.
Single quoted char literals are therefor the default xelf format.

Char literals can be back-tick quoted multi line raw literals without escape sequences. This is
especially useful when use in templates or everywhere else with large pre-formated char literals.

Idxr and Keyr literals can omit commas, because it does not fit in with the lisp style expression
syntax as later discussed. And simple dict keys can be symbols and do not need any quotes.

Composite literals can only contain literals. Any opening square or curly braces always start a
literal. Expression resolvers are used to construct literals from expressions, instead of reusing
the literal syntax. This makes it visually more obvious whether something is a literal.

Types, functions and forms do implement the main literal interface and can be used as literals in
some cases. This also simplifies the already heavy resolution API, as a successful resolution will
always return a valid literal, and alleviates another check after each resolution step.

All literals except booleans are parsed as the base types any, num, char, idxr and keyr. They are
later converted to a specific type inferred from the expression context.

Every environment working with xelf literals requires some adapter code to convert, compare or
otherwise work with literal. The xelf go package provides interfaces for each class of literal
behaviors and both generic and proxy adapters. The generic adapters provide an abstract
representation for all literals. Proxy adapters implement the literal interface but write through to
to any compatible backing value in the native environment.

Null literal
------------

The 'null' literal turns out to be very useful if treated as a universal zero value. That means
that it can be used in every type context as an appropriate zero value. This helps to translate
the concept to languages that do not have a null pointers and use none and some for optional types.

Type Conversion
---------------

We need flexible type conversion rules, mostly because xelf is a typed language using untyped JSON
literals. The conversion rules are a bit more involved for that reason.

Allowed conversions are encoded by the compare function in the typ package. It returns a comparison
bit-set, that indicates not only whether, but in what way a type can be converted to another. The
possible conversions are grouped into levels: equal, comparable, convertible or checked convertible.

Equal types indicate the same types or that the destination type is inferred.

Comparable types can automatically convert to the destination. All literal types are comparable to
the any type. Specific types are comparable to their base or opt types, while the primitive base
types are comparable to their specific type, unless the specific type requires a strict format. That
means that char is comparable to str or enum, and the num type is comparable to any specific numeric
type.

Convertible types cover list, dict and record types whose element types also convertible.

Checked convertible type might be convertible, but need to check the literal value to decide if they
actually are. This is the case for types containing unresolved type reference, if the source type is
the any, idxr or keyr type and should convert to a more specific type, or is an list, dict or record
type, whose element types are checked convertible, or is the char type that should convert to a
specific type with a strict format (raw, uuid, time and span).

Symbols, Names and Keys
-----------------------

Xelf symbols are ASCII identifiers that allow a large number of punctuation characters. Some
punctuation characters already have a designated meaning. As prefix ':+-@~', as suffix '?' and the
prefixes '$/.' for special scope lookups. All other punctuation characters can be used in client
libs. Built-in expression resolvers all use short ASCII names instead of punctuation.

By using only the ASCII character set we can avoid any encoding issues or substitutions in
environments without unicode identifier support.

Xelf will need to work in environment that are case-sensitive and case-insensitive. To address that,
cased names are usually used for declarations and are then automatically lowercased for all lookups
in the resolution environment. This way we do not have to use configurable casing rules to generate
idiomatic code for the go target.

Compound names, that would use either CamelCase, `snake_case` or kebab-case depending on the
environment, like ClientID, are instead used as cased name for the go target and simply lowercased
for all others. This avoids a lot of busy code to convert from one identifier flavor to another
and avoids potentially even more confusion.

Symbols in literals must be simple alphanumeric names and only allow the underscore as valid
punctuation. This makes a potentially problematic colon parsing in keyer literals a non-issue.

Expression Syntax
-----------------

Xelf must be very simple to parse. Infix notation is always harder to parse than s-expressions. So
we naturally choose LISP style parenthesis enclosed expressions. The parenthesis are also used for
defining complex types. But apart from that always indicate that a resolver is called with the
expression element.

LISP languages are great. However, many key concepts of LISP-languages are not easily expressed in
simple environments. Xelf builds on JSON and adds a notation for types and expressions on top of it.

Xelf has special handling for tag and declaration symbols within expressions. This is to avoid
excessive nesting of expressions and to achieve a comfortable level of expressiveness in a variety
of contexts. Tag symbols that start with a colon and can be used for named arguments, node
properties or similar things. Declaration symbols starting with a plus sign are used to signify
variables, parameters or field names in declarations or when setting elements by key.

Predefined Symbols
------------------

Only the literal symbols null, false and true as well as the void type are hard keywords. All other
types are just built-in definitions that can be overwritten in sub environment.

The schema prefix can be used to refer to built-in types, even if shadowed by a another definition.

Type References and Variables
-----------------------------

Type references refer to any literal in scope and represent the underlying type of the literal. They
start with the at-sign followed by a symbol '@name' and can refer to a generalized element type of
any container with a underscore path segment `@mylist._`.

Type variables represent unresolved types and are used during type inference and checking. They
start with reference prefix followed by a numeric type variable id '@123'. The naked at-sign '@'
without id represents a new type variable with any new id.

Type variables can have one parameter acting as constraint, that is usually a base type or a type
alternative. For example '@1:num' or '(@:alt str raw list)'.

Schema types are a kind of reference and need to be resolved. The name of schema types refers to a
global type schema and uses the schema prefix '~schema.model' for lookups from the environment.

Self and ancestor references point to the current or ancestor record type and are used for recursive
type definitions. They use the schema prefix followed by a number '~1'.

The schema prefix also qualifies all named types in an expression context, it can generally omitted
for prominent types or in a type context.

Symbol Resolution
-----------------

Some type names and core expressions have names commonly used in application data. Xelf uses a
number of prefixes, to avoid shadowing, long names, implicit, or explicit namespaces. The schema and
type reference prefixes do only refer to types, while path prefixes starting with a dot, dollar or
slash only refer to literal data. Symbols without prefix are most importantly used for predefined
types and expression resolvers, or otherwise explicitly declared resolvers.

To avoid checking for prefixes in each environment the default resolution checks for prefixes and
selects the appropriate environment. The environments implement a simple method to indicate whether
they support one of the special prefixes.

The tilde '~' prefix qualifies a type and must be followed by a symbol in an expression context.
The type is either a basic type or a schema type, that needs to be looked up.

Starting dots '.' are used for data lookups. The dot starts a relative path to a data scope. One dot
represents the immediate data scope's literal, each additional dot moves one data scope up. If the
first dot is followed by a question mark '?' the default resolver tries each data scope, starting
with the immediate one.

The dollar '$' and slash '/' prefixes have a similar effect as the data scope, but are exclusively
used for parameter and result paths. Both use the immediate environment supporting the prefix and
can be followed by dots to select a supporting parent. A double prefix '$$' or '//' will select the
outermost supporting environment.

The with expression takes a literal as first argument that creates a data scope. Normal functions
and loop expressions provide their arguments as a data scope as well.

Type references not starting with dots first try to resolve as relative path symbol and after that
as plain symbol.

Expressions
-----------

Xelf language elements can be literals including types, symbols or expressions. Expressions can
either be named, dynamic or call expressions. All elements share a common interface, that includes a
sting and write bfr method as well as a type method. The returned type identifies the kind of the
language element.

Named expressions start with a tag or declaration symbol and are handled by the parent's
specification. They are formed automatically by the parser from tag and declaration tokens.

Dynamic expressions are expressions, where the resolver is yet unknown and may start with a literal
or sub expression. Dynamic expressions starting with a literal are resolved with the configurable
dyn resolver.

Calls are all expressions where a specification was found.

Expression Resolution
---------------------

Dynamic expressions direct a considerable part of the resolution process, and provide a configurable
way to extend the language with new syntax. Because the dyn resolver plays this central role it does
not use a lookup from the environment on every call and instead uses a reference in the resolution
context or uses the default dyn resolver. Changing the dyn resolver is still possible by copying the
context.

The default dyn resolver resolves the first arg and delegates to a resolver based on it. If the
first argument is form or function it is called directly. If it is a type the expression is treated
as the 'as' type conversion form. For other literals a appropriate combination operator is used if
available. Users can redefine and reuse the dyn resolver to add custom delegations.

There is only one resolver interface for all aspects of the resolution process to keep it simple.
For that reason a context is passed in that indicates whether an expression can execute or can
return a partially resolved expression. The context also encapsulates the default resolution
machinery that resolvers can choose to resolve arguments.

The form resolvers provided by xelf are grouped into the core, std and library resolvers. The core
built-ins include basic operators, conditional and the dyn and as resolvers. The std built-ins have
basic resolvers that include declarations. The library built-ins are provide extra functionality
centered around one type.

Forms and Functions
-------------------

Form and function types provide a signature that can be used to direct most aspects of the
resolution process. The signature allows us to factor out the default type checking and inference
and provide a more stable and comfortable user experience. Forms signatures accept special
parameter names, that allow tags and declaration parsing.

Form and function types have a reference name, primarily to be printable. This name is not meant to
be resolved, but should match the definition.

A form is free to choose how its arguments are typed and resolved and can even change the
environment for subexpressions.

Functions are called only if all their arguments are successfully resolved and then use their
declaration environment for evaluation, acting as closures. If the last function argument parameter
has an list type, it can be called as variadic parameter - meaning multiple arguments can be used
instead of the expected list. When exactly one argument is used that is convertible to the list type
it is used-as, other cases are treated as element.

Specification
-------------

Specifications are quasi-literals with a form or function type and a resolver.

The 'fn' form can be used to construct function literals. Simple function expression should be able
to omit and infer the function signature.

If we have a full function type as hint, inferring the signature could be as simple as checking if
all parameter references work with the declared type and whether the result type if comparable. The
dot prefix is used and allows path to use either keys or indices to refer to the parameters. The
underscore refers to the first parameter which allows use to easy infer the type signatures with two
parameters.

To infer the signature without any hint we must deduce all parameter references and their order as
well as the result type. The prefix allow us to identify all parameter references. We can use index
parameters to explicitly order some of the parameters append named ones in order.

Type Inference
--------------

Xelf uses a Hindley-Milner based type system with some modifications to accommodate the complex type
conversion rules. Type alternatives are used internally to collect all possible options and type
variables can be constrained. The unification process chooses the most specific type that satisfies
all alternatives.

Form signatures can be used to automatically infer the type for some arguments most of the time.
Some forms, like the built-in 'and', 'if', 'map' and others, cannot express their type as signature
and must handle the unification in the resolver.

Resolvers are passed a type hint to unify with. Type hints can be of any kind, but are usually type
variables created in the parent's resolver. Void hints indicates a lack of type expectations and
means the resolver can disregard the hint completely.

