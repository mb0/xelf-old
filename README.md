xelf
====

xelf is a cross-platform expression language framework. It is meant to be a simple notation for all
things crossing process or language boundaries.

A small set of opinionated packages are the foundation to build simple domain specific languages,
that provide expression resolution and a practical type system with inference out of the box.

It is in an alpha state and can be used for experimentation.

Overview [![GoDoc](https://godoc.org/github.com/mb0/xelf?status.svg)](https://godoc.org/github.com/mb0/xelf)
--------

 * [cor](https://godoc.org/github.com/mb0/xelf/cor):
   minimal runtime utilities for generated code
 * [bfr](https://godoc.org/github.com/mb0/xelf/bfr):
   common interface for buffered writers and bytes buffer pool
 * [lex](https://godoc.org/github.com/mb0/xelf/lex):
   token lexer and tree scanner
 * [typ](https://godoc.org/github.com/mb0/xelf/typ):
   composable type system and a parser, comparison and unification
 * [lit](https://godoc.org/github.com/mb0/xelf/lit):
   literal parser, adapters and support for comparison and conversion
 * [exp](https://godoc.org/github.com/mb0/xelf/exp):
   simple extensible expression language
 * [std](https://godoc.org/github.com/mb0/xelf/std):
   built-in expression resolvers
 * [utl](https://godoc.org/github.com/mb0/xelf/utl):
   extra utilities and resolvers

Motivation
----------

The author envisioned this tool, while building a typical back-office software:

 * where configuration files could benefit from simple expressions
 * templates could help generating html and PDFs on both the server and client
 * complex queries for data heavy pages would accumulate expression like parameters
 * data schema was needed at runtime and to generate code for different languages

For all of those cases there are projects and solutions available that can be used and implemented.
Each has its own environment with a different syntax and limitations. That is fine at first.

Then you want to format the text of some ingredients bold on both its product label and the label
preview or use the domain model to create HTML views in your client.
Or really any situation where you need to share data and behaviour between different environments.
The resulting hacked up adapters and solutions continue to grow and keep you busy writing similar
boilerplate all over in each environment for each change.

It usually starts with a repetitive pattern on either the client or the server, after some
abstraction you have a small package API that is needed on the other side as well.
Now you rewrite that code in another programming language and add an ugly JSON based data format.
After some time you have multiple tiny DSLs that are neither well-defined, pretty to look at nor
easy to work with, all whilst repeating similar data validation and manipulation with some
bug-prone permutations.

JSON was the minimal go-to data format used by the author. While not elegant or comfortable, every
environment supports it well. Using JSON, however, is not ergonomic, when used as a product label
layout language or in any other more involved situation.

After about two years of experiments of varying degree I naturally arrived at a Lisp-style syntax,
using a simple, yet powerful type system in combination with JSON compatible literals, a small set
of built-in operators and expressions and an extensible evaluation process, that can be used to
liberally change or extend the language.

The result is not a Lisp and is much more restricted than one, primarily to make it as easy as
possible to translate expressions to idiomatic code in different languages, even SQL.

License
-------

Copyright (c) 2019 Martin Schnabel. All rights reserved.
Use of the source code is governed by a BSD-style license that can found in the LICENSE file.
