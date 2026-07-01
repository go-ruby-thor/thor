// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package thor is a pure-Go (CGO-free) MRI-faithful reimplementation of the
// deterministic core of Ruby's Thor CLI-framework gem (thor 1.5.0). It covers
// the three interpreter-independent halves of Thor a host embeds: declarative
// option/argument parsing ([Options], [Arguments]), the command registry and
// argv-to-command dispatch ([Base]), and byte-faithful usage/help generation.
//
// Running a command body invokes user Ruby (or Go) code; that is the host seam.
// This package resolves argv into (command, parsed-options, remaining-args) and
// renders help exactly as the gem does, so go-embedded-ruby can bind Thor with
// no Ruby runtime and reproduce the gem's output and error text byte-for-byte.
//
// # Value model
//
// Parsed option values use a small, fixed set of Go types so a host can map its
// object graph to and from this package. The mapping mirrors the one Thor's
// parser produces:
//
//	Ruby (Thor)      Go
//	-----------      --
//	nil              nil
//	true / false     bool
//	:string value    string
//	:numeric int     int64
//	:numeric float   float64
//	:array           []string
//	:hash            *OrderedMap (string->string, insertion order)
//
// A parse yields a [*Result]: its Options is an insertion-ordered map from the
// option human name to its value, and Args holds the non-option remainder.
//
// # Terminal width
//
// Help/table/wrap layout depends on terminal width. For deterministic,
// ruby-free golden output this package uses a fixed default of 80 columns,
// overridable per call via [Config.TerminalWidth] or the THOR_COLUMNS
// environment variable, matching Thor::Shell::Terminal.
package thor
