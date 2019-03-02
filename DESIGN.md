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
The time type number represents milliseconds since the unix epoch.

The selection is based on what types strings and number literals are commonly used for, that
differentiate enough in comparison or manipulation behavior. It is heavily informed by types
supported by full featured databases like postgresql. We do not bother with different number sizes
or text length because it only complicates what xelf tries to provide and is only a concern to a
validation or storage layer. For most uses a couple of wasted bytes is not an issue.

There are two distinct behavior of container types called idxer or keyer. An idxer provides access
to its elements by index, and a keyer by a string key. The idxer base type list can have elements of
any type, while the arr type takes an explicit element type. The keyer types are the base type dict,
the map type with explicit element type, as well as obj and rec with field info.  Because object
fields are inherently ordered, both the obj and rec types do implement idxer as well.  Dict is
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

To make working with types easier in a language context there are also the void and typ type.

The flag, enum and rec types are called schema types and are global references a type definition
with additional information. The flag type is a integer type used as bit set with associated
constants. The enum type is a string type with associated constants. A record is an object with a
known schema.

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

Composite literals can only contain literals. Any opening square or curly braces always start a
literal. Expression resolvers are used to construct literals from expressions, instead of reusing
the literal syntax. This makes it visually more obvious whether something is a literal.

Types, functions and forms do implement the main literal interface and can be used as literals in
some cases. This also simplifies the already heavy resolution API, as a successful resolution will
always return a valid literal, an alleviates another check after each resolution step.

All literals except booleans are parsed as the base types any, num, char, list and dict.
The literals are usually converted to a specific type inferred from the expression context.

Every environment working with xelf literals requires some adapter code to convert, compare or
otherwise work with literal. The xelf go package provides interfaces for each class of literal
behaviors and both generic and proxy adapters.

Null literal
------------

The 'null' literal turns out to be very useful if treated as a universal zero value. That means
that it can be used in every type context as an appropriate zero value. This helps to translate
the concept to language that do not have a null pointers and use none and some for optional types
and do not provide.

Type Conversion
---------------

We need flexible type conversion rules, mostly because xelf is a typed language using untyped json
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

Xelf symbols are ascii identifiers that allow a large number of punctuation characters. Some
punctuation characters already have a designated meaning. As prefix ':+-@', as suffix '?' and
'$/.~' for special scope lookups. All other punctuation characters can be used in client libs.
Built-in expression resolvers all have short ascii names for that reason.

By using only the ascii character set we can avoid any encoding issues or substitutions in
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

I would love to incorporate more parts of the lisp family of languages, but in many cases that does
not work well when translating to the simpler targeted environments. Starting with logical
expressions not resulting in booleans and ending at macros. Lisps are simply too powerful.

Xelf has special handling for tag and declaration symbols within expressions. This is to avoid
excessive nesting of expressions and to achieve a comfortable level of expressiveness in a variety
of contexts. Tag symbols that start with a colon and can be used for named arguments, node
properties or similar things. Declaration symbols starting with a plus sign are used to signify
variables, parameters or field names in declarations or when setting object fields or map elements
by key.

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

There are five kinds of reference types.

Self and ancestor reference refer to the current or ancestor object type and are used for recursive
type definitions.

Schema types are a kind of reference and need to be resolved. The reference name of schema types
refers to a global type schema and uses the schema prefix '~' for lookups from the environment.

Form types are flagged as reference types, they represent built-in resolvers that might not conform
to a specific type signature and must be resolved as part of an expression.

Unnamed references are inferred types and need to be inferred from context. They are mostly used in
the resolution phase and may represent poly types by collecting candidates in the companion field
list normally used by object types.

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
is resolved after another.  The first path segment can only be a key in the current context.
This key is used to lookup the resolver. Following path segments are used to select into the
resolved literal. This implies that resolver names cannot contains dots and instead should
use other punctuation without special meaning. Library resolvers use a colon for that reason.

To avoid checking the symbol in each environment the default resolution should do most of the heavy
lifting. It should check if a symbol has a special prefix or is a path, and with that information
select the appropriate environment from the whole ancestry of the current one. Environments need a
a way to advertise special behaviour for this to work. Because interface calls where tested to be
cheaper than type checks we add a method the environment, instead of using marker interfaces. We
could expand the use of types for all expression elements to avoid type checking where the typed
value is not required anyway.

Expression Resolution
---------------------

Xelf expression elements can be literals including types, references or sub expressions.

Sub expression can either be tag and declaration expressions or dynamic and normal expressions.

Tag and declaration expressions start with a tag or declaration symbol respectively and are handled
by the parent expression's resolver. They are formed automatically by the parser from tag and
declaration symbols in expression arguments.

Dynamic expressions are expressions, where the resolver is yet unknown and may start with a literal
or sub expression. Dynamic expressions starting with a literal are transformed are resolved with a
configurable dyn resolver.

The default dyn resolver resolves the first arg and delegates to another resolvers based on it. If
the first argument is a type the expression is treated as the 'as' type conversion resolver. For
other literals a appropriate combination operator is used if available. Users can redefine and
reuse the dyn resolver to add custom delegations.

Dynamic expressions direct a considerable part of the resolution process, and provide a configurable
way to extend the language with new syntax. Because the dyn resolver plays this central role
it would be appropriate to avoid the lookup from the environment on every call. Changing the dyn
resolver should be possible, but is unusual enough as to justify new resolution context. The
context can have a custom dyn resolver set and otherwise falls back on a built-in default.

Normal expressions are all expressions where the resolver is known.

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
functions or built-in expressions. However this meant that even simple resolvers that could be
expressed in xelf expressions needed to be implemented in every environment used. The concept
of function is also common enough to justify special handling to reduce code duplication.

Functions as literals also imply a function type that represents the type's signature. A declared
signature allows us to factor out the default type checking and inference and provide a more stable
and comfortable user experience. Functions are called only if all their arguments are successfully
resolved and then use their declaration environment for evaluation.

Resolvers that need to have control over type checking or can partially resolve their arguments
must be implemented as form resolver.

Because we have different kinds of function implementation we use a common function literal type,
that implements the literal and resolver interface and delegates expression resolution to a
function body implementation. The different are kinds built-in functions with custom or reflection
base resolvers and normal function bodies with a list of expression elements.

Functions are allowed to access the environment chain they were declared in. This is only useful
for normal functions and means their implementations need to remember the declaration environment.
A special function scope provides the parameter declarations that were resolved in the calling
environment.

Normal functions need to be inlined in environments without function literals, but can avoid
work along the way.

A new resolver 'fn' is used to construct function literal. It has the same parameter and result
declaration as the function type syntax but ends in a tail of body elements. Simple function
expression should be able to omit and infer the function signature.

If we have a full function type as hint, inferring the signature could be as simple as checking if
all parameter references work with the declared type and whether the result type if comparable. This
either implies that we must have names for every parameter or use a new syntax that allows
parameter references by index.

To infer the signature without any hint we must deduce all parameter references and their order as
well as the result type. We can assume that all free references for which no resolver is found in
the declaration scope are parameters, but then we still need to figure out the order and would be
effectively limited to functions that take at most one parameter or happen to refer to parameters
in the desired order.

Using a special prefix that marks a references to parameters and allows to refer by index, would
make things easier. We have the program parameter prefix that already supports most of the
requirements. However, because normal functions are closures, using them as-is would be confusing
when parameter environments are nested. Using a new prefix does not solve the problem. What we need
is to refine the parameter syntax so we can explicitly refer to a specific parameter environment.

We already have the concept of relative paths to select a specific environment and can reuse it for
parameter and maybe even the result path resolution. The parameter prefix should lead the symbol to
clearly identify it as parameter path. It can then be followed by dots, each dot selects the next
parameter environment. Parameters are in general only looked up from one parameter environment and
none of its parents. A parameter without dots should therefor refer to the immediate environment,
one dot to its parent and so on.  The program parameters are special in that reference to them will
most likely end up in deeply nested environments. A double parameter prefix '$$' could more clearly
identify program parameters in those cases.

Functions should be used to model all expressions that use their own isolated and parameterized
environment like loop actions or the with expressions.

Form Type and Literal
---------------------

Form literals are used for expression resolvers that cannot be expressed as function. Form literals
are a more general expression resolver and can direct most aspects of the resolution process. Using
literals for all expression resolvers allows us to use reference resolution to lookup the resolver.

Most of the built-in resolvers do not conform to a simple type signature and need to resolve their
arguments for type information anyway. Modeling these signatures with the type system in the same
way as function types would have no clear benefit. It was therefor decided to introduce another
quasi-literal for those cases resolvers.

Form types can have a signature and the default resolution does use result types if specified.
Form parameters could be formalized and used to validate the form arguments at some point, but are
only used as documentation hint for now. The plan is to interpret base types in form signatures as
type hints.

Form types have a reference name, primarily to be printable. This name is not meant to be resolved,
but usually matches the definition name of that form.

Form Parsing
------------

This is a work in progress.

Expressions have different layouts regarding the number and shape of their arguments. Function
for example accept leading plain elements and optionally tagged elements for named parameters.
Forms can have any layout including not only tag expressions but also declaration expressions.

We started with a couple of helper functions, that validate expression arguments against common
form layouts. Since we now have form signatures with type hints, we could use these signatures
to direct the argument validation and check their types at the same time. Because parameter names
have no other use in form parameters we use them together with the parameter type to sufficiently
identify the desired layout.

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
types. Luckily we already have type references, that we can use as type variables, the infer type,
that can act as a poly type behind the scenes, a context to stash intermediate results and an
environment to look it all up. Using function signatures and implementing type resolution in all
form resolvers potentially provides use with all the types we need.

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
