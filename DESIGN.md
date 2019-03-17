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
   list for JSON arrays
   dict for JSON objects

However we usually need more specific type information that can not be represented in JSON.
The type system chose a number of specific primitive types. The specific types of num are int, real
and flag for bit sets, the specific types of char are str, raw for bytes, uuid, enum, time and span
for durations. The bool type is also considered a numeric type, because some environments might not
have a dedicated bool type, indexedDB in browsers comes to mind. Both span and time are usually
represented in a text format but can also be converted to an integer, representing milliseconds.
The time type as num are milliseconds since the unix epoch.

The selection is based on what types strings and number literals are commonly used for, that
differentiate enough in comparison or manipulation behavior. It is heavily informed by types
supported by full featured databases like postgresql. We do not bother with different number sizes
or text length because it only complicates what xelf tries to provide and is only a concern to a
validation or storage layer. For most uses a couple of wasted bytes is not an issue.

There are two distinct behavior of container types called idxer or keyer. An idxer provides access
to its elements by index, and a keyer by a string key. The idxer base type list can have elements of
any type, while the arr type takes an explicit element type. The keyer types are the base type dict,
the map type with explicit element type, as well as obj and rec with field info. Because object
fields are inherently ordered, both the obj and rec types do implement idxer as well. Dict is
implemented to preserve field order for this reason, but does not provide the idxer interface.

Apart from the any type all primitive and object types can be optional. Optional type variants have
a question mark suffix and translate to pointer, option or nullable types in the target
environments. Note that there are also optional object fields that mark the field itself and not its
value optional.

To rely on type information without specifying it explicitly in every case, for possibly large
composite literals, xelf must allow to explicitly refer to and infer from existing types. This is
covered by type embedding and type references. Recursive obj and rec declaration also use special
type references to itself or an ancestor. Type references are implicitly used in the type inference
and resolution phase.

The flag, enum and rec types are called schema types and are global references to a type definition
with additional information. The flag type is a integer type used as bit set with associated
constants. The enum type is a string type with associated constants. A record is an object with a
known schema.

The void, typ and exp types are special types and are useful in a language context. The exp types
provide a type signature for form, func types. They are otherwise used as an universal indicator to
differentiate between all language elements. This allows us to replace some type switches and
assertions to use a simple interface call instead.

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

Xelf literal parsing rules are a superset of JSON with minor additional rules to address specific
issues, mainly to make writing literals by hand easier.

Char literals can be single quoted escaped literals. Xelf is more often than not handled as string
inside other environments. Single quoted string alleviate escaping quotes in double quoted strings.
Single quoted char literals are therefor the default xelf format.

Char literals can be back-tick quoted multi line raw literals without escape sequences. This is
especially useful when use in templates or everywhere else with large pre-formated char literals.

List and dict literals can omit commas, because it does not fit in with the lisp style expression
syntax as later discussed. And simple dict keys can be symbols and do not need any quotes.

Composite literals can only contain literals. Any opening square or curly braces always start a
literal. Expression resolvers are used to construct literals from expressions, instead of reusing
the literal syntax. This makes it visually more obvious whether something is a literal.

Types, functions and forms do implement the main literal interface and can be used as literals in
some cases. This also simplifies the already heavy resolution API, as a successful resolution will
always return a valid literal, and alleviates another check after each resolution step.

All literals except booleans are parsed as the base types any, num, char, list and dict.
The literals are usually converted to a specific type inferred from the expression context.

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

Convertible types cover arr, map and obj types whose element types also convertible.

Checked convertible type might be convertible, but need to check the literal value to decide if they
actually are. This is the case for types containing unresolved type reference, if the source type is
the any, list or dict type and should convert to a more specific type, or is an arr, map or obj
type, whose element types are checked convertible, or is the char type that should convert to a
specific type with a strict format (raw, uuid, time and span).

Symbols, Names and Keys
-----------------------

Xelf symbols are ASCII identifiers that allow a large number of punctuation characters. Some
punctuation characters already have a designated meaning. As prefix ':+-@~', as suffix '?' and the
prefixes '$/.' for special scope lookups. All other punctuation characters can be used in client
libs. Built-in expression resolvers all use short ascii names instead of punctuation.

By using only the ASCII character set we can avoid any encoding issues or substitutions in
environments without unicode identifier support.

Xelf will need to work in environment that are case-sensitive and case-insensitive. To address
that, cased names are usually used for declarations and are then automatically lowercased. All
lookups in the resolution environment, map elements or object fields must use lowercase keys. This
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
we naturally choose lisp style parenthesis enclosed expressions. The parenthesis are also used for
defining complex types. But apart from that always indicate that a resolver is called with the
expression element.

Lisp is great. However, many key concepts of lisp languages are not easily expressed in simple
environments. Xelf builds on JSON and adds a notation for types and expressions on top of it.

Xelf has special handling for tag and declaration symbols within expressions. This is to avoid
excessive nesting of expressions and to achieve a comfortable level of expressiveness in a variety
of contexts. Tag symbols that start with a colon and can be used for named arguments, node
properties or similar things. Declaration symbols starting with a plus sign are used to signify
variables, parameters or field names in declarations or when setting object fields or map elements
by key.

Predefined Symbols
------------------

Only the literal symbols null, false and true as well as the void type are hard keywords. All other
types are just built-in definitions that can be overwritten in sub environment.

The schema prefix can be used to refer to built-in types, even if shadowed by a another definition.

Types are mostly used in context where they are explicitly expected. Maybe we can introduce a rule
stating that in those cases simple names are always treated as type, while every other case resolves
the name in its environment.

Type References
---------------

There are four kinds of reference types.

Self and ancestor reference refer to the current or ancestor object type and are used for recursive
type definitions.

Schema types are a kind of reference and need to be resolved. The reference name of schema types
refers to a global type schema and uses the schema prefix '~' for lookups from the environment.

Unnamed references are inferred types and need to be inferred from context. They are mostly used in
the resolution phase and may represent poly types by collecting candidates in the companion field
list normally used by object types.

Normal type references refer to any literal in scope and represent the type of that literal. Type
references referring to type literals do use the underlying type and not the special typ type.

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

The tilde '~' prefix is used for schema lookups and is followed by a model name that is usually
qualified by its schema name or needs a context that can infer a default schema.

Starting dots '.' are used for data lookups. The dot starts a relative path to a data scope. One dot
represents the immediate data scope's literal, each additional dot moves one data scope up. If the
first dot is followed by a question mark '?' the default resolver tries each data scope, starting
with the immediate one.

The dollar '$' and slash '/' prefixes have a similar effect as the data scope, but are exclusively
used for parameter and result paths. Both use the immediate environment supporting the prefix and
can be followed by dots to select a supporting parent. A double prefix '$$' or '//' will select the
outermost supporting environment.

The with expression starting with a plain literal argument creates a data scope. The literal
argument can still be followed by declarations or have nested let expression that define normal
scope resolvers. Normal functions and loop expressions provide their arguments as a data scope as
well.

Type references not starting with dots first try to resolve as relative path symbol and after that
as plain symbol. Schema prefix could become completely unnecessary at some point, because the
plain symbol namespace is much less crowded and we can use it for schema types as well. We also
moved the type parser to the exp package, that allows us to look up schema data from the resolution
environment. There might not even any name conflicts because we only use references to models or
their contents and not the schema, if we otherwise disallow dots in resolver names.

Expressions
-----------

Xelf language elements can be literals including types, symbols or expressions. Expression can
either be tag and declaration expressions or dynamic and normal expressions. All elements share a
common interface, that includes a sting and write bfr method as well as a type method. The returned
type identifies the kind of the language element.

Tag and declaration expressions start with a tag or declaration symbol respectively and are handled
by the parent expression's resolver. They are formed automatically by the parser from tag and
declaration symbols in expression arguments.

Dynamic expressions are expressions, where the resolver is yet unknown and may start with a literal
or sub expression. Dynamic expressions starting with a literal are resolved with the configurable
dyn resolver.

Normal expressions are all expressions where an expression resolver was found.

Expression Resolution
---------------------

Dynamic expressions direct a considerable part of the resolution process, and provide a configurable
way to extend the language with new syntax. Because the dyn resolver plays this central role it does
not use a lookup from the environment on every call and instead uses a reference in the resolution
context or uses the default dyn resolver. Changing the dyn resolver is still possible by copying the
context.

The default dyn resolver resolves the first arg and delegates to a resolver based on it. If the
first argument is func or form it is called directly. If it is a type the expression is treated as
the 'as' type conversion resolver. For other literals a appropriate combination operator is used if
available. Users can redefine and reuse the dyn resolver to add custom delegations.

A resolver is free to choose how its arguments are typed and resolved and can even change the
environment for subexpressions. There should only be one resolver interface for all aspects of
the resolution process. For that reason a resolution context is passed in that indicates whether
an expression can execute or can return a partially resolved expression. The context also
encapsulates the default resolution machinery that resolvers can choose to resolve arguments.

The form resolvers provided by xelf are grouped into the core, std and library resolvers. The core
built-ins include basic operators, conditional and the dyn and as resolvers. The std built-ins
have basic resolvers that include declarations. The library built-ins are usually provide extra
functionality centered around a type of literals.

Function Type and Literal
-------------------------

Xelf tried to build around the concept of resolvers that do not differentiate between variables,
functions or built-in expressions. However this meant that even simple resolvers, that could be
expressed in xelf expressions, needed to be implemented in every environment used. The concept
of function is also common enough to justify special handling to reduce code duplication.

Functions as literals also imply a function type that represents the type's signature. A declared
signature allows us to factor out the default type checking and inference and provide a more stable
and comfortable user experience. Functions are called only if all their arguments are successfully
resolved and then use their declaration environment for evaluation.

Resolvers that need to have control over type checking or can partially resolve their arguments
must be implemented as form resolver.

Because we have different kinds of function implementation we use a common function literal type,
that implements the literal and expr resolver interface and delegates expression resolution to a
function body implementation. The different are kinds built-in functions with custom or reflection
base resolvers and normal function bodies with a list of expression elements.

Functions are allowed to access the environment chain they were declared in. This is only useful
for normal functions and means their implementations need to remember the declaration environment.
A special function scope provides the parameter declarations that were resolved in the calling
environment.

Normal functions need to be inlined in environments without function literals. For this reason they
are allowed to execute without exec context. Other functions only call the body if exec is true.

A new resolver 'fn' is used to construct function literal. It has the same parameter and result
declaration as the function type syntax but ends in a tail of body elements. Simple function
expression should be able to omit and infer the function signature.

If we have a full function type as hint, inferring the signature could be as simple as checking if
all parameter references work with the declared type and whether the result type if comparable. The
parameter syntax can use either index or key notation to refer to the parameters.

To infer the signature without any hint we must deduce all parameter references and their order as
well as the result type. The prefix allow us to identify all parameter references. We can use index
parameters to explicitly order some of the parameters append named ones in order.

Functions should be used by other expressions, that need to execute in an isolated and parameterized
environment like loop actions or the with expressions.

Form Type and Literal
---------------------

Form literals are used for expression resolvers that cannot be expressed as function. Form literals
are a more general expression resolver and can direct most aspects of the resolution process. Using
literals for all expression resolvers allows us to use reference resolution to lookup the resolver.

Most of the built-in resolvers do not conform to a simple type signature and need to resolve their
arguments for type information anyway. Modeling these signatures with the type system in the same
way as function types would have no clear benefit. It was therefor decided to introduce another
quasi-literal for those cases.

Form types have a reference name, primarily to be printable. This name is not meant to be resolved,
but should match the definition name of that form.

Expressions have different layouts regarding the number and shape of their arguments. Function for
example accept leading plain elements and optionally tagged elements for named parameters. Forms
can have any layout including not only tag expressions but also declaration expressions.

Form types have a signature of type hints. Layouts can be used to validate and resolve args against
the form signature. And the type resolution machinery uses specific result types.

Type Inference
--------------

This is a work in progress.

After some study over the hindley-milner type system, I come to the conclusion, that it does not
lend itself to be faithfully applied to xelf. We cannot separate type checking from the resolution
process, because xelf allows resolvers to direct most aspects of the process. Form resolvers often
need to resolve their arguments to provide the result type information, expressing their signatures
in type variables and constraints would be more work without significant value. We also have
auto conversion rules between base type and comparable special types.

Instead we need a way to check and infer types within the same resolution context and process. It
should be flexible enough to cover functions or references and help form expression infer their
types. Luckily we already have type references, that we can use as type variables and as poly type
behind the scenes, a context to stash intermediate results and an environment to look it all up.
Using function signatures and implementing type resolution in all form resolvers potentially
provides use with all the types we need.

We pass a type hint to resolvers, so it can be considered when inferring types and encapsulate their
type resolution. The other option would be to handle type checking and inference at the call site,
but this would limit what resolvers could infer.

The type hint is interpreted differently than normal types and should work the same way as types in
form parameters. This is done to make hints more expressive for the common cases. Type hints can
be transformed into a corresponding poly type. If type hint should explicitly represent the void,
any or a base type it can be expressed as poly type with the hint as only option.

Void indicates a lack of type expectations and means the resolver can disregard the hint completely.
The any type indicates a common literal type excluding the special typ, func and form types. The
base types represents the type itself as well as all its specific types.
