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

One aspect of xelf is to provide a unified way to work with data across different environments
and needs to choose the data format that poses the least resistance.

The most basic environment to think of here is a some-what idiomatic sqlite3 target.

JSON emerged quite early and organically as the most commonly implemented and used object format.
It is very basic, well-known, and programmers are already used to map their data to a JSON
representation one way or another. Recently more and database solution got JSON support.

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
   list for JSON arrays
   dict for JSON objects

However we usually need more specific type information that can not be represented in JSON.
The type system chose a number of specific primitive types. The specific types of num are int, real
and flag for bit sets, the specific types of char are str, raw for bytes, uuid, enum, time and span
for durations. The bool type is also considered a numeric type, because some environments might not
have a dedicated bool type, indexedDB in browsers comes to mind. Both span and time are usually
represented as in a text format but can also be converted to an integer, representing milliseconds.
The time type number represents milliseconds since the unix epoch.

The selection is based on what types strings and number literals are commonly used for that
differentiate enough in comparison or manipulation behavior. It is heavily informed by types
supported full featured databases like postgresql. We do not bother with different number sizes
or text length because it only complicates what xelf tries to provide and is only a concern to a
validation or storage layer. For most uses a couple of wasted bytes is not an issue.

The specific container types are a typed arr|T or unordered map|T with string keys, as well
as the somewhat special obj and rec types, with field access by both index and key.

Apart from the any type all primitive and object types can be optional. Optional type variants
have a question mark suffix and translate to pointer, option or nullable types in the target
environments. There are also optional object fields that mark the field itself and not its value
optional.

The 'null' literal turns out to be very useful if treated as a universal zero value.

To rely on type information without specifying it explicitly in every case for possibly large
composite literals, xelf must allow to explicitly refer to and infer from existing types. This is
covered by type embedding and type references. Recursive obj and rec declaration also use special
type references to itself or an ancestor. Type references are implicitly used in the type inference
and resolution phase.

To make working with types easier in a language context there are also the void and typ type.

The flag, enum and rec types are named schema types that globally reference a type definition with
additional information. The flag type is a integer type used as bit set with associated constants.
The enum type is a string type with associated constants. A record is an object with a known
schema.

The final naming of the types took a long time and serious consideration. Real does not imply
a floating point precision as much as float does and fits the short naming schema. Dict implies
an ordered list if of keyed entries, while map implies a key is mapped to a kind of value. List is
just a sequence of values and arr is hopefully more associated with a specific element type. An
object has fields or properties, while a record implies multiple instances and a significant type.

To efficiently work with types most of the information is encoded in the 'kind' bit set. This bit
set is very compact and can quickly answer questions in varying granularity. For example:
An encoder is only interested in the base type, while the resolver only wants to check if its a
reference type. The reference name and field information are stored in a companion object.

The kind encodes the element type for arr and map types. Each type slot uses one byte and we can
only assume 53bit precision when a float64 may be involved. It follows that arr and map types can
only be nested six levels deep. This could be worked around, but I do not expect the situation to
arise.

Literals
--------

Xelf literal parsing rules are a superset of json with minor additional rules to address specific
issues, mainly to make writing literals by hand easier.

Char literals can be single quoted escaped literals. Xelf is more often than not handled as string
inside other environments. Single quoted string alleviate escaping quotes in double quoted strings.
Single quoted char literals are therefor the default xelf format.

Char literals can be back-tick quoted multi line raw literals without escape sequences. This is
especially useful when use in templates or everywhere else with large pre-formated char literals.

List and dict literals can omit commas, because it does not fit in with the lisp style expression
syntax as later discussed. And simple dict keys can be symbols and do not need any quotes.

All literals except booleans are parsed as the base types any, num, char, list and dict.
The literals are usually converted to a specific type inferred from the expression context.

Composite literals can only contain literals. Any opening square or curly braces always start a
literal. Expression resolvers are used to construct literals from expressions, instead of reusing
the literal syntax. This makes it visually more obvious whether something is a literal.

Every environment working with xelf literals requires some adapter code. The xelf go package
provides interfaces for each class of literal behaviors and both generic and proxy adapters.

Types do implement the main literal interface and can be used as literals in some cases. This also
simplifies the already heavy resolution API, as a successful resolution will always return a valid
literal, an alleviates another check after each resolution step.

Symbols, names and keys
-----------------------

Xelf symbols are ascii identifiers that allow a large number of punctuation characters. Some
punctuation characters already have a designated meaning. As prefix ':+-@', as suffix '?' and
'$/.~' for special scope lookups. All other punctuation characters can be used in client libs.
Built-in expression resolvers all have short ascii names for that reason.

By using only the ascii character set we can avoid any encoding issues or substitutions in
environments without unicode identifier support.

Xelf will need to work in environment that are case-sensitive and case-insensitive. To address
that cased names are usually used that will automatically be lowercased to its key form. All
lookups in the resolution environment, map elements or object fields must use the key form. This
way we do not have to use configurable casing rules to generate idiomatic code for the go target.

Compound names that would use either CamelCase, snake_case or kebab-case depending on the
environment like ClientID are instead used as cased name for the go target and simply lowercased
for all others. This avoids a lot of busy code to convert from one identifier flavor to another
and avoids potentially even more confusion.

Symbols in literals must be simple alphanumeric names and only allow the underscore as valid
punctuation. This makes a potentially problematic colon parsing in map literals a non-issue.

Expression Syntax
-----------------

Xelf must be very simple to parse. Infix notation is always harder to parse than s-expressions. So
we naturally choose lisp style parenthesis enclosed expressions. The parenthesis always signify an
expressions.

I would love to incorporate more parts of the lisp family of languages, but in many cases that does
not work well when translating to the simpler targeted environments. Starting with logical
expressions not resulting in booleans and ending at macros. Lisps are simply too powerful.

Xelf has special handling for tag and declaration symbols within expressions. This is to avoid
excessive nesting of expressions and achieve a comfortable level of expressiveness in a variety of
contexts. Tag symbols that start with a colon and can be used for named arguments, node properties
or similar things. Declaration symbols starting with a plus sign are used to signify variable,
parameter or field names in declarations or when setting object fields or map elements by key.

There is a couple of predefined forms that resolvers can use to validate expression arguments.

Predefined Symbols
------------------

Both the literal symbols null, false, true as well as the type names are considered keywords and
always refer to the predefined meaning regardless of the context. While all resolvers are just
built-ins in the root environment and can be replaced in sub environments.

Treating literal symbols and type names as keyword substantially simplifies symbol resolution.
However the type names contain common short names that may occur in user code and models, for that
reason symbols referring to scoped variables or resolver can use a path syntax to force scope
lookup. This path syntax can also be used to select shadowed variables from a parent scope.

Type References
---------------

There are four kinds of reference types.

Self and ancestor reference refer to the current or ancestor object type and are used for recursive
type definitions.

Schema types are a kind of reference and need to be resolved. The reference name of schema types
refers to a global type schema and uses the schema prefix '~' for lookups from the environment.

Unnamed references are inferred types and need to be inferred from context. They are mostly used
in the resolution phase and may represent poly types by collecting candidates in the companion
info object field list normally used by object types.

Normal type references refer to a literal in scope and represent the type of that literal.
For type literals the referred to type is the types identity and not the typ type.

Symbol Resolution
-----------------

The normal symbol resolution process usually inspects the symbol for known special prefixes.
The common prefixes are:

The tilde '~' prefix is used for schema lookups and is followed by a model name that is usually
qualified by a schema name or needs a context that can infer a default schema.

The dollar '$' prefix is used for parameter lookups and is treated as a literal itself.

The slash '/' prefix is used for result lookups and is also treated as a literal.

Starting dots '.' are used for relative lookups. A single dot indicates that the following symbol
can only be found in the current environment, each additional dot moves one parent up the ancestry.
The special prefix '.?' can be used to force the default scope lookup for predefined name like list
which is also a type name.

If a symbol does not start with a special lookup modifier prefix and instead starts with a name
start character, it is parsed as a path in case it contains interior dots. And each path segment
is resolved after another.  The first segment path segment can only be a key in this context.
This key is used to lookup the resolver. Following path segments are used to select into the
resolved literal. This implies that resolver names cannot contains dots and instead should
use other punctuation without special meaning. Library resolvers use a colon for that reason.


Expression Resolution
---------------------

Xelf expression elements can be literals including types, references or sub expressions.

Sub expression can either be tag and declaration expressions or dynamic and normal expressions.

Normal expressions start with a symbol that can be used to lookup a resolver from the environment.

Tag and declaration expressions start with a tag or declaration symbol respectively and are handled
by the parent expression's resolver. They are formed automatically by the parser from tag and
declaration symbols in expression arguments.

Dynamic expressions start with a literal or sub expression. Dynamic expressions starting with a
literal or sub expression are transformed to a normal expression with the predefined name 'dyn'.

The default 'dyn' resolver again delegates to other resolvers based on the arguments. If the first
argument is a type the expression is treated as the 'as' type conversion and constructor resolver.
For other literals a appropriate combination operator is used if available. Users can redefine and
reuse the 'dyn' resolver to add custom delegations.

Dynamic expressions fill an otherwise corner-case of the syntax with a configurable way to provide
syntax sugar for common type casting or literal operations.

A resolver is free to choose how its arguments are typed and resolved and can even change the
environment for subexpressions. There should only be one resolver interface for all aspects of
the resolution process. For that reason a resolution context is passed in that indicates whether
an expression can execute or can return a partially resolved expression. The context also
encapsulates the default resolution machinery that resolvers can choose to resolve arguments.

The resolvers provided by xelf are grouped into the core, std and library resolvers. The core
built-ins include basic operators, conditional and the dyn and as resolvers. The std built-ins
have basic resolvers that include declarations. The library built-ins are usually provide extra
functionality centered around a type of literals.

Function type
-------------

This is a work in progress.

In the beginning xelf tried to build around the concept of resolvers that do not differentiate
between variables, functions or built-in expressions. This was by design, as heavy use of function
was thought to complicate translations to environments that do not support these contexts.

However this meant that even simple resolvers that could be expressed in xelf expressions needed to
be implemented in every environment used. We face much of the same concerns for loop actions of any
form. Functions are also handy to factor out common expressions from large queries.

Functions as literals also imply a function type that represents the type's signature. A declared
signature allows us to factor the default type checking and inference out of each resolver and
provide a more stable and comfortable user experience. Each resolver can decide whether to call
into the default type checking machinery. The resolver only returns a custom function literal when
treated as a symbol reference.

Function can be inlined in environments without function literals and can avoid work along the way.

I am not sure if functions should be allowed to access or modify their non-immediate environment
and work like closures or be completely side effect free and treated as an isolated program. The
isolation approach is much simpler to reason about for sure and allows us to reuse the parameter
lookup prefix.

Type inference
--------------

This is a work in progress.

After some study over the hindley-milner type system, I come to the conclusion, that it does not
lend itself to be faithfully applied to xelf.

If we embrace function types for all resolvers we already have a enough type information for all
language elements. We still might need a type hint because of the automatic base type conversion
rules. We can store type options in unnamed inferred reference type to work with intermediate
poly types in the resolution phase or express overloaded resolvers like the arithmetic resolvers.
