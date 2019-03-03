/*
Package typ provides a restricted but combinable type system.

There are five groups of types that share some behaviour: numeric, character, list, dictionary and
special types. Special types are void, any, the special typ type represing a type literals and type
references, as well as the exp types dyn, form, func, tag and decl.

The other four groups each have a base type, which are num, char, list and dict. Base types are
usually only used as long as no specific type cound be resolved. Their explicit use is discouraged.
Specific types should be used whenever possible.

There is a number of specific types for each base type:
    bool, int, real, time and span are numeric types
    str, raw, uuid, time and span are character types
    arr and obj are list types
    map and obj are dictionary types

The character, numeric and the special types any and void are represented by their name:

    bool, any, time

The time and span type represents a time or a time duration/interval/delta that can either be
represented as numeric value in milliseconds since epoch and ms delta or in a character format
as specified in the cor package. Their default representation is the character format.

The arr and map types have a type slot and can be nested at most seven times. These types are
represented by their name appended by a pipe and the slot type:

    arr|int, map|bool, arr|arr|arr|int, map|arr|map|arr|map|arr|str


The obj type has a sequence of fields, that can be accessed by name or index, therefor an obj is
both a list and dictionary type. The obj type must have field declarations and be enclosed in
parenthesis. A field declaration consists of the declaration name starting with plus sign and the
field type definition. Optional fields have names ending with a question mark, otherwise a field is
required.

    (obj +x +y +z? int), (arr|obj +name str +val any +extra? any)

Optional types are nullable type-variants. The any, list, dict, arr, map and exp types are always
optional and the void and typ and exp types never are. All the other primitive, object and reference
types can be marked as optional by a question mark suffix.

    (obj +top10 (arr|obj? +name str +score int?) +err str?)

The exp types form and func aso use the field declaration syntax used as parameter and result type
signature. The last field signifies the result type and is usually unnamed. All other fields
represent the parameters. Function parameters must have a type and can be named. The form type
must be name and it can use untype parameters.

Type references start with an at sign '@name' and represent the type of what 'name' resolves to.
References need to be resolved in a declaration context for this reason.

An unnamed type reference '@' means the type must be inferred and replaced by the resolver.

The @name form can be used as an alias in place of any type definition. Type references in unnamed
field declarations are embedded. Embedding an obj type is the same as copying all of its fields
into the new obj type, while for all other references fields are named by the last simple name part
in the reference name.

    (let
        +kind  int
    	+named (obj +id uuid +name str?)
	+cat   (obj + @named @kind? +prods arr|@named)
    )

Self references of the form '@1' are a special references to the current '@0' or the parent '@1' or
the grand parent '@2' and so on for the whole object type ancestry.

Schema types are also a kind of type references, that are treated as a specific type in most cases.
Schema types reference a global type definition and as such must be resolved. Other than normal
references the identifier is kept alongside the full type data after resolution.

    flag is a named int type bit-set that consists of multiple bit constants
    enum is a named str type that consists of one string constant of an enumeration
    rec  is a named obj type that has additional type and field details

The global identifier allows users to associate extra data and behaviour to these types.

All non-special types and the any type are called literal types. Concrete literal types are all
literal types except the base types. All the other special types are not considered literal types.
Even though reference may resolve to a literal and types themself can be considered a literal.

Minimum restrictions apply for compatibility with JavaScript and JSON:

    map  keys are restricted to string
    int  is limited to 53bit precision
    real has no NaN or infinities
    time has millisecond granularity

Types don't usually need to be written as JSON, because both client and server expect a given schema.
But when they need to be serialized, it should look like this:

    {
    	"kind": {"typ": "int"},
	"named":{"typ": "obj", "fields": [
		{"name": "id", "typ": "uuid"},
		{"name": "name", "typ": "str?"}
	]},
	"cat":  {"typ": "obj", "fields": [
		{"typ": "ref", "ref": "named"},
		{"typ": "ref?", "ref": "kind"},
		{"name": "prods", "typ": "arr|ref", "ref": "named"}
    	]}
    }

*/
package typ
