/*
Package typ provides a restricted but combinable type system.

There are five groups of types that share some behaviour: numeric, character, list, dictionary and
special types. Special types are void, any and ref. The other four groups each have a base type:
    num, char, list and dict

The base types are use when a specific type is not known or required. Base types can not be used in
obj field, list or map declarations, but can be use in variable or parameter declarations.

There is a number of specific types for each base type:
    bool, int, real and span are numeric types
    str, uuid, time and span are character types
    arr and obj are list types
    map and obj are dictionary types

The character, numeric and the special types any and void are represented by their name:

    bool, any, time

The span type represents a time duration/interval/delta that can either be represented as
numeric value in milliseconds or in the character formats '-1234h5m6s789ms' or '-1234:05:06.789'.

The arr and map types have a type slot and can be nested at most seven times. These types are
represented by their name appended by a colon and the slot type:

    arr:int, arr:arr:arr:int, map:arr:map:arr:map:arr:str


The obj type has sequence of field, that can be accessed by name or index, therefor an obj is both
a list and dictionary type. The obj type must have field declarations and be enclosed in
parenthesis. A field declaration consists of the declaration name starting with plus sign and the
field type definition. Optional fields have names ending with a question mark, otherwise a field is
required.

    (obj +x +y +z? int), (arr:obj +name str +val any +extra? any)

Optional types are nullable type-variants. The any, list, dict, arr and map types are always
optional and the void type never is. All other types can be marked as optional by a question mark
suffix.

    (obj +top10 (arr:obj? +name str +score int?) +err str?)


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
	+cat   (obj + @named @kind? +prods arr:@named)
    )

There are more quasi-reference types, that are treated as a specific type in most cases, but carry
more information for handling in a domain modeled declaration context.

    flag is a named int type bit-set that consists of multiple bit constants
    enum is a named str type that consists of one string constant of an enumeration
    rec  is a named obj type that has additional type and field details

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
		{"name": "prods", "typ": "arr:ref", "ref": "named"}
    	]}
    }

*/
package typ
