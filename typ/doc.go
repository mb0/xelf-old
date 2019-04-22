/*
Package typ provides a restricted but combinable type system.

There are five groups of types that share some behaviour: numeric, character, indexer, keyer and
special types. Special types are void, any, the special typ type representing a type literals and
type variables, alternatives and references, as well as the exp types dyn, form, func, tag and decl.

The other four groups each have a base type, which are num, char, idxr and keyr. Base types are
usually only used as long as no specific type could be resolved. Their explicit use is discouraged
and needs to use the schema prefix: ~num ~keyr.

There is a number of specific types for each base type:
    bool, int, real, time and span are numeric types
    str, raw, uuid, time and span are character types
    list, rec and obj are indexer types
    dict, rec and obj are keyer types

The character, numeric types and void type are represented by their name:
    bool, void, time

The plain special types and base types use the schema prefix, as not to pollute the namespace.
    ~any, ~num, ~keyr, ~dyn

The time and span type represents a time or a time duration/interval/delta that can either be
represented as numeric value in milliseconds since epoch and ms delta or in a character format
as specified in the cor package. Their default representation is the character format.

The list and dict type can have a type parameters and can be nested.

    list, list|int, dict|bool, list|list|list|int, dict|list|dict|list|dict|list|str


The record types rec and obj have type parameters representing fields, that can be accessed by key
or index, therefor a record type is both an indexer and keyer type. Records must have at least one
parameter and must be enclosed in parenthesis. A parameter declaration consists of an optional name
tag starting with a colon followed by a type definition. Optional fields have names ending with a
question mark, otherwise a field is considered required.

    (rec :x :y :z? int), (list|rec :name str :val any :extra? any)

Optional types are nullable type-variants. The any, list, dict and exp types are always optional and
the void and typ and exp types never are. All the other primitive, record and reference types can be
marked as optional by a question mark suffix.

    (rec :top10 (list|rec? :name str :score int?) :err str?)

The exp types form and func also use the type parameters syntax used as argument and result type
signature. The last parameter signifies the result type and is usually unnamed. All other parameters
represent the form arguments. Function parameters must have a type and may be named. Form parameters
must be name and can omit the type.

Type variables start with an at sign '@123' followed by a type id. They represent an unresolved
type during type inference. An variable without var id '@' means a new id must be assigned.

Type references start with an at sign '@name' followed by a symbol and represent the type of what
the symbol resolves to. References need to be resolved in a declaration context for this reason.

The @name form can be used as an alias in place of any type definition. Type references in unnamed
field declarations are embedded. Embedding a record type has the same effect as copying all of its
fields into the new type, while for all other references fields are named by the last simple
name part in the reference name.

    (let
        +kind  int
    	+named (rec +id uuid +name str?)
	+cat   (rec + @named @kind? +prods list|@named)
    )

Self references of the form '~1' are a special references to the current '~0' or the parent '~1' or
the grand parent '~2' type and so on for the whole record type ancestry.

Schema types are also a kind of type references, that are treated as a specific type in most cases.
Schema types reference a global type definition and as such must be resolved. Other than normal
references the identifier is kept alongside the full type data after resolution.

    flag is a named int type bit-set that consists of multiple bit constants
    enum is a named str type that consists of one string constant of an enumeration
    obj  is a named rec type that has additional type and field details

The global identifier allows users to associate extra data and behaviour to these types.

All non-special types and the any type are called literal types. Concrete literal types are all
literal types except the base types. All the other special types are not considered literal types.
Even though references may resolve to a literal type, they can be considered a literal.

Minimum restrictions apply for compatibility with JavaScript and JSON:

    map  keys are restricted to string
    int  is limited to 53bit precision
    real has no NaN or infinities
    time has millisecond granularity

Types don't usually need to be written as JSON, because both client and server expect a given schema.
But when they need to be serialized, it serializes the type in the xelf representations to a json
object field 'typ':

    {
    	"kind": {"typ": "int"},
	"named":{"typ": "(rec +id uuid +name str?)"},
	"cat":  {"typ": "(rec + @named + @kind? +prods list|@named)"}
    }

*/
package typ
