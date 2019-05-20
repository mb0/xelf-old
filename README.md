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
   literal parser, generic implementations and support for comparison and conversion
 * [prx](https://godoc.org/github.com/mb0/xelf/prx):
   literal adapters and proxies to native go data using reflection
 * [exp](https://godoc.org/github.com/mb0/xelf/exp):
   simple extensible expression language
 * [std](https://godoc.org/github.com/mb0/xelf/std):
   built-in expression resolvers
 * [utl](https://godoc.org/github.com/mb0/xelf/utl):
   extra utilities and resolvers

Motivation
----------

The author envisioned this tool, while building a typical back-office software, where:

 * configuration files could benefit from variables and simple expressions
 * templates could help generating html and PDFs on both the server and client
 * complex queries for data heavy pages would accumulate expression like parameters
 * data schema was needed at runtime and to generate code for different languages

For all of those cases there are projects and solutions available that can be used and implemented.
Each has its own environment with a different syntax and limitations. That is fine at first.

But then you want to customize one aspect of the library you are using for your problem or need to
use the same functionality in another language. The result is adapter code or handwritten niche
solutions, that continue to grow and keep you busy writing similar boilerplate for each environment,
for every change.

The vision for xelf is to have basic meta-language as versatile tool for creating simple domain
specific languages, that can be used as data format and translated to other language targets. In
contrast to other DSL frameworks, scripting languages or LISPs, the xelf language is specifically
designed to be easy to gradually implement and to work with in other languages.

After about two years of experiments of varying success, I naturally arrived at a Lisp-style syntax,
using a simple, yet powerful type system in combination with JSON compatible literals, a small set
of built-in operators and expressions and an extensible evaluation process, that can be used to
liberally change or extend the language.

The result is not a Lisp and is much more restricted than one, primarily to make it as easy as
possible to translate expressions to idiomatic code in different languages, even SQL.

Examples
--------

There are two projects in development, that will demonstrate how xelf can be used.

The [daql](https://github.com/mb0/daql) project provides tools to define, migrate and query domain
models as well as packages used for to facilitate these features.

The [layla](https://github.com/mb0/layla) project provides a layout format, that can be printed on a
specific label printer or rendered as PDF or HTML preview.

License
-------

Copyright (c) 2019 Martin Schnabel. All rights reserved.
Use of the source code is governed by a BSD-style license that can found in the LICENSE file.
