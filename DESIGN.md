Design
======

This document discusses and sometimes explains the design trade-offs chosen for xelf. Unanswered
questions about the design of xelf can be raised with the author and will be incorporated into
this document. This document along with each package documentation acts as informal and incomplete
specification as long as the project is a work in progress.

The main mantra for xelf is simplicity and practicality in the context of web-centric technologies.
The different packages build on each other and can be gradually implemented on other platforms.

Xelf provides runtime code and development tools to work with data in a unified way in -and generate
schemas and code for- targeted platforms. Web apps usually consist of a server program, a web client
often using javascript and persistent data storage like a sql database. The most basic platform to
think of here is a some-what idiomatic sqlite3 target.

JSON
----

One aspect of xelf is to provide a unified way to work with data across different environments.
It needs to choose the data format that poses the least resistance.

JSON emerged quite early and organically as the most commonly implemented and used object format.
It is very basic, well-known, and programmers are already used to map their data to a JSON
representation one way or another. Recently more and more database solution got JSON support.

Accepting plain JSON as valid literal syntax makes it very simple to interoperate with existing
code, databases, and user assumptions. Building xelf around JSON at its core is very practical, but
also provides limitations that need to be accepted or worked around.

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

However we usually need more specific type information that can not be represented in JSON. The type
system chose a number of specific primitive types. The specific types of num are int, real and bits
type for bit sets and span for time durations. The specific types of char are str, raw for bytes,
uuid, enum and time. The bool type is also considered a numeric type, because some environments
might not have a dedicated bool type, indexedDB in browsers comes to mind. Both span and time are
usually represented in a text format but can also be represented as integer, representing
milliseconds. The numeric value of a time are the milliseconds since the unix epoch.

The selection is based on what types character and numeric literals are commonly used for, that
differentiate enough in comparison or manipulation behavior. It is heavily informed by types
supported by full featured databases like postgresql. We do not bother with different number sizes
or text length because it only complicates what xelf tries to provide and is only a concern to a
validation or storage layer. For most uses a couple of wasted bytes is not an issue.

There are two distinct behavior of container types called idxr and keyr. An indexer provides access
to its elements by index, and a keyer by a string key. The idxr and keyr base type can hold any
literals, while the list and dict type take an optional element type. They use different types to
avoid mix-ups with the record types, that do implement both indexer and keyer interface.
Because record fields are inherently ordered, dict is implemented to preserve field order, but does
not provide the idxr interface. Another reason in go to avoid maps, is that they do not support
interior pointers, which complicates working wit proxy literals.

Apart from the any type all primitive and record types can be optional. Optional type variants have
a question mark suffix and translate to pointer, option or nullable types for the target platform.
Note: there are also optional record fields that mark the field itself and not its value optional.

To rely on type information without specifying it explicitly in every case, for possibly large
composite literals, xelf must allow to explicitly refer to and infer from existing types. This is
covered by type embedding and type references. Recursive record declarations also use special
type references to itself or an ancestor. Type variables and alternatives are implicitly used in the
type inference and resolution phase.

The bits, enum and obj types are called schema types and are global references to a type definition
with additional information. The bits type is a integer type used as bit set with associated
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
especially useful when used in templates or anywhere else for large pre-formated char literals.

Idxr and Keyr literals can omit commas, because it does not fit in with the lisp style expression
syntax as later discussed. And simple dict keys can be plain symbols without quotes.

Composite literals can only contain literals. Any opening square or curly braces always start a
literal. Expression resolvers are used to construct literals from expressions, instead of reusing
the literal syntax. This makes it visually more obvious whether something is a literal.

Types, functions and forms specs do implement the main literal interface and can be used as literals
in some cases. This also simplifies the already heavy resolution API, as a successful resolution
will always return a valid literal, and alleviates another check after each resolution step.

All literals except booleans are parsed as the types any, num, char, list and dict. They are later
converted to a specific type inferred from the expression context.

Every platform working with literals requires some adapter code to convert, compare or otherwise
work with them. Xelf provides generic literal implementations in the lit package, that can be used
as abstract representation for all literals. And reflection based proxy literals in package prx,
that can be used to adapt and write through to compatible typed go data structures.

The 'null' literal is used as universal zero value. It can be used in typed context and represents
the specific zero value for that type. The zero value for a str is and empty string while the zero
value for an optional string is a null pointer. This helps to translate the concept to languages
that do not have a null pointers and use none and some for optional types.

Type Conversion
---------------

Xelf has flexible type conversion rules. Those rules make it possible to use untyped JSON literals
in in a typed language. The conversion rules are rather complex, but not really avoidable for the
stated reason.

Allowed conversions are encoded by the compare function in the typ package. Compare returns a
comparison bit-set, that indicates not only whether, but in what way a type can be converted to
another. The possible conversions are grouped into levels: equal, comparable, convertible or checked
convertible.

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
prefixes '$/.' for special scope lookups. All other punctuation characters can be used in language
extensions. Built-in expression resolvers all use short ASCII names instead of punctuation.

By using only the ASCII character set we can avoid any encoding issues or substitutions in
environments without unicode identifier support.

Xelf will need to work in environments that are case-sensitive and case-insensitive. To address
that, cased tag symbols are usually automatically lowercased for all lookups in the resolution
environment. This way we do not have to use configurable casing rules to generate idiomatic code for
the go target.

Compound names, that would use either CamelCase, `snake_case` or kebab-case depending on the
environment, like ClientID, are instead used as cased name for the go target and simply lowercased
for all others. This avoids a lot of busy code to convert from one identifier flavor to another
and avoids potentially even more confusion.

Symbols in literals must be simple alphanumeric names and only allow the underscore as valid
punctuation. This makes a potentially problematic colon parsing in keyer literals a non-issue.

Expression Syntax
-----------------

Xelf uses prefix notation for expressions and is very simple to parse. Infix notation is always
harder to parse than s-expressions, so we naturally choose LISP style parenthesis enclosed
expressions.  The parenthesis are also used for complex type definitions, but otherwise always
indicate a call to a function or form resolver.

LISP languages are great. However, many key concepts of LISP-languages are not easily expressed in
other environments. Xelf builds on JSON and adds a notation for types and expressions on top of it.

Predefined Symbols
------------------

The literal symbols null, false and true as well as common type names are hard keywords. All other
types are just built-in symbols that can be overwritten in sub environment.

Type References and Variables
-----------------------------

Type references refer to any literal in scope and represent the underlying type of the literal. They
start with the at-sign followed by a symbol '@name' and can refer to a generalized element type of
any container with a underscore path segment `@mylist._`.

Type variables represent unresolved types and are used during type inference and checking. They
start with reference prefix followed by a numeric id '@123'. The naked at-sign '@' without id
represents a new type variable with any new id.

Type variables can have one parameter acting as constraint, that is usually a base type or a type
alternative. For example '@1|num' or '(@|alt str raw list)'.

Schema types are a kind of reference and need to be resolved. The name of schema types refers to a
global type schema and uses the schema prefix '~schema.model' for lookups from the environment.

Self and ancestor references point to the current or ancestor record type and are used for recursive
type definitions. They use the schema prefix followed by a number '~1'.

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
The type is a schema type, that needs to be looked up.

Starting dots '.' are used for data lookups. The dot starts a relative path to a data scope. One dot
represents the immediate data scope's literal, each additional dot moves one data scope up. If the
first dot is followed by a question mark '?' the default resolver tries each data scope, starting
with the immediate one.

The dollar '$' and slash '/' prefixes are used for global parameter and result paths.

The with expression takes a literal as first argument that creates a data scope. Normal functions
provide their arguments as a data scope as well.

Type references not starting with dots first try to resolve as relative path symbol and after that
as plain symbol.

Expressions
-----------

Xelf language elements can be atoms, symbols, named elements or dynamic and call expressions. All
elements share a common interface, that includes a string and write bfr method as well as a traverse
and type method. The returned type identifies the kind of the language element.

Named elements start with a tag symbol and are handled by the parent's
specification. They are formed automatically by the parser.

Dynamic expressions are expressions, where the resolver is yet unknown and may start with a literal
or sub expression. Dynamic expressions starting with a literal are resolved with the configurable
dyn resolver.

Calls are expressions where a specification was found and the arguments layout constructed.

Expression Resolution
---------------------

Resolution and evaluation require a program that holds a type context and encapsulates the
resolution machinery.

Dynamic expressions resolve the first argument and construct a call based on it. If the first
argument is form or function it is called directly. If it is a type the expression is treated as the
'con' type construction or conversion form. For other literals a appropriate combination operator is
looked up from the program and used if available. Language extensions can change the dyn lookup.

The resolvers provided by xelf are grouped into the core, std and library resolvers. The core
built-ins include basic operators, conditional and the 'con' resolver. The std built-ins
have basic resolvers that include declarations. The library built-ins in package utl provide extra
functionality centered around one type.

Forms and Functions
-------------------

Form and function types provide a signature that can be used to direct most aspects of the
resolution process. The signature allows us to factor out the default type checking and inference
and provide a more stable and comfortable user experience. Forms signatures accept special
parameter names, that allow tags parsing.

Form and function types have a reference name, primarily to be printable. This name is not meant to
be resolved, but should match the definition.

A form is free to choose how its arguments are typed and resolved and can even change the
environment for subexpressions.

Functions are called only if all their arguments are successfully resolved and then use their
declaration environment for evaluation, acting as closures. If the last function argument parameter
has an list type, it can be called as variadic parameter - meaning multiple arguments can be used
instead of the expected list. When exactly one argument is used that is convertible to the list type
it is used-as, other cases are treated as element.

Specifications are quasi-literals with a form or function type and a resolver.

The 'fn' form can be used to construct function literals. Simple function expression can omit and
infer the function signature.

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

Syntax Changes
--------------

We could remove some ambiguity by using angle brackets for type literals. On the one hand these need
escaping in html contexts, on the other hand there are no other ascii enclosures left, type literals
should be mostly be avoidable in scripts.

I'm starting to be annoyed by the difference of tags starting with a colon and keyer fields
ending in a colon. Maybe tags should also end in a colon, because we cannot change literal parsing
without straying away from the json base concept. The lexer would always handle the colon
as special token creating a named pair with a name on its left and expression on the left side.

To tags without value we use a semi-colon with a leading symbol or string a; -> a:,.

I'm unsure whether the decls syntax is any good. It was eliminated for all forms except the daql dom
and query packages. The main goal was to handle xml-style like structure writable without additional
parenthesis. Decls are not used as declarations anymore and should therefor be at least
renamed as a concept. The plus and minus sign is therefor not appropriate and sticks out visually.
The only thing we use it for in the dom and qry packages is to collect a list of possibly named
subexpressions that are resolved by the parent form. 

I came to the conclusion to just use tags with some naming convention instead. My final idea is just
to use a case based discriminator for attribute and child tags in the dom and qry packages.

	(<rec a?:int>  $a)
	(model Test help:'This is the help attribute'
		ID:   (int pk;)
		Help: (str help:'the help field uses uppercase')
	)

Type Changes
------------

Apart from the mentioned syntax changes for types, we also want introduce expression bits that
allow to narrow down to what typed thing we expect. So we can express that we expect any list type,
or a specifically a constant literal or an unevaluated expression that resolves to a certain type.

One example is the qry packages where clause that is not evaluated but must evaluate to type bool.
We can express that in the form signature now and do not need custom resolution.

The other change is to better distribute the kind bitmask and clean up some ideas.

Planed Tasks
------------

We should use the element visitor in more places. One reason is that the interface calls to concrete
types is cheaper than type conversions and facilities using the visitor can more easily be reused
and extended.

Use the element visitor to build up the program input type, and spec literal parameter types.

With atoms now in place we could add a type field, that indicates the resolved literal type. This
would allow delayed literal conversions and would make it possible to drop at the base-type literals
char and num.

We could change the literal parser api to work with raw lex trees until a concrete literal is
required and then directly write into a provided proxy literal. This would would reduce allocations
especially for large container literals.

We should add more test and clear some Todo items. Especially record, function and form conversion
is not implemented, as are many edge-cases that generally need investigations.

It would be good to have an extensive test harness for all the std specs, that must be passed by all
targeted platforms. We need to test partial evaluation and that all specs respect the context exec
flag.

The expression resolution context should use context package for cancellation. A call to the context
err method a significant points in the evaluation process should be enough.

We should also populate the source information for all expressions. Partially evaluated expressions
should use the source positions of the origin expression. That means that the final result source
is likely the root expression or a branch for short circuiting forms like 'if'.

Explore and document the significance of the type unification order. We should think about how to
minimize type variables to make the type inference easier to grok. Maybe use local type contexts,
that we merge back into the larger program type context, omitting all locally inferable type vars.

We need better documentation and error messages for resolution and type errors, before the project
is advertised to a wider audience.

Implement planned tasks for the daql and layla example projects, to discover and fix potential
issues with xelf. If the work on those projects and examples using them stabilizes, we can invest
time cleaning up apis, writing a handbook and tutorials and think about releasing and evangelizing
the project.
