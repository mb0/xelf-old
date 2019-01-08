xelf
====

xelf is a cross-environment expression language framework.

It is currently a work in progress.

Overview
--------

 * cor: minimal runtime utilities for generated code
 * bfr: common interface for buffered writers and bytes buffer pool
 * lex: lexer, scanner and string quoting code
 * typ: composable type system
 * lit: literal parser, adapters and support code like comparison and conversion
 * xpr: simple expression language, standard built-ins and resolvers

Motivation
----------

The author found envisioned this tool, while building a typical back-office software:

 * Config files for multiple deployments are easier to manage with merging and data validation.
 * Template languages are used for HTML, emails, PDFs or even receipts on both server and client.
 * Complex queries for other data heavy pages accumulate parameters and quasi-expressions.
 * Event sourcing composite field updates, e.g. tag list, have special notations sooner or later.
 * Domain models are useful at compile and runtime both on the server and the client.

For all of those cases there are projects and solutions available that can be used and implemented.
Each has its own environment with a different syntax and limitations. That is fine at first.

Then you want to format the text of some ingredients bold on both its product label and the label
preview or use the domain model to create HTML views in your client.
Or really any situation where you need to share data and behaviour between different environments.
The resulting hacked up adapters and solutions keep growing and breaking and keep you busy writing
similar boilerplate all over in each environment for each change.

It usually starts with repetitive pattern on either the client or the server, after some
abstraction you have a small package API that you actually could use on the other side as well.
Now you rewrite that code in another programming language and add an ugly JSON based data format.
After some time to grow, you find yourself with multiple tiny DSLs that are neither well-defined,
pretty to look at nor easy to work with, all whilst writing the same data validation and
manipulation code over and over with some bug-prone permutations.

JSON was the minimal go-to data format used by the author. While not elegant or comfortable, every
environment supports it well. Using JSON, however, is not ergonomic when used as a product label
layout language or any other more complex situation.

After about two years of experiments of varying degree I naturally arrived at a Lisp-style syntax,
using a simple, yet powerful type system in combination with JSON compatible literals, a tiny set
of built-in operators and expressions and an extensible evaluation process, that can be used to
liberally change or extend the language.

The result is not a Lisp and much more restricted than that, primarily to make it as easy as
possible to translate expressions to idiomatic code in different languages, even SQL.


Copyright (c) 2019 Martin Schnabel. All rights reserved.
Use of the source code is governed by a BSD-style license that can found in the LICENSE file.
