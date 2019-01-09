/*
Package typ provides a restricted but combinable type system.

There are five groups of types that share behaviour or format:
    special, numeric, character, sequence, dictionary

The special types void, any and ref are synthetic and cover special cases.

The other four groups each have a base type: num, char, seq and dict.
The base types are use when no specific type is not known or required. Base types can not be used
in obj field, list or map declarations, but can be use in variable or parameter declarations.

There is a number of specific types for base type:
    bool, int and float are numeric types
    str, uuid, time and span are character types
    list and obj are sequence types
    map and obj are dictionary types

The character and numeric types and the special types any and void are represented by their name:
    bool, any, time

The list and map types have a single type argument. These types are represented by their name
appended by a colon and the element type:
    list:int, map:list:str

The obj type has field arguments. Fields are a sequence of name-type pairs and can be accessed
either by name or by index, therefor the obj is both a sequence and dictionary type.
An obj type must be followed by its field declarations and is enclosed in parenthesis.
A field declaration consists of a declaration name starting with plus sign and a type definition.
Optional fields have names ending with a question mark, otherwise a field is required.
    (obj +x +y +z? int), (list:obj +name str +val any +extra? any)

Optional types on the other hand are nullable type-variants. The any, seq, dict, list and map
types are always optional and the void type never is. All other types can be marked as optional by
a question mark suffix.

    (obj +top10 (list:obj? +name str +score int?) +err str?)

Minimum restrictions apply mainly for compatibility with javascript and JSON:

    map keys are restricted to string
    int is limited to 53bit precision
    float has no NaN or infinities
    time has millisecond granularity

Type references start with an at sign '@name' and represent the type of what 'name' resolves to.
References need to be resolved in a declaration context for this reason.

An unnamed type reference '@' means the type must be inferred and replaced by the resolver.

The @name form can be used as an alias in place of any type definition. Type references in unnamed
field declarations are embedded. Embedding an obj type is the same as copying all of its fields
into the new obj type, while for all other references fields are named by the last simple name part
in the reference name.

    (obj
        +kind  int
    	+named (obj +id uuid +name str?)
	+cat   (obj + @named @kind? +prods list:@named)
    )

Types don't usually need to be written as JSON, because both client and server expect a given schema.
But when they need to be serialized, the should look like this:

    {"typ": "obj", "fields": [
	["kind": {"typ": "int"}],
    	["named":{"typ": "obj", "fields": [
		{"name": "id", "typ": "uuid"},
		{"name": "name", "typ": "str?"}
	]}],
	["cat":  {"typ": "obj", "fields": [
		{"typ": "ref", "ref": "named"},
		{"typ": "ref?", "ref": "kind"},
		{"name": "prods", "typ": "list:ref", "ref": "named"}
    	]}]
    ]}

*/
package typ
