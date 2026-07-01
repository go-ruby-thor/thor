// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

// ErrorKind classifies a Thor parse/dispatch error by the gem exception class
// it corresponds to, so a host can re-raise the matching Ruby exception.
type ErrorKind int

const (
	// KindRequiredArgumentMissing maps to Thor::RequiredArgumentMissingError.
	KindRequiredArgumentMissing ErrorKind = iota
	// KindMalformattedArgument maps to Thor::MalformattedArgumentError.
	KindMalformattedArgument
	// KindUnknownArgument maps to Thor::UnknownArgumentError.
	KindUnknownArgument
	// KindExclusiveArgument maps to Thor::ExclusiveArgumentError.
	KindExclusiveArgument
	// KindAtLeastOneRequiredArgument maps to Thor::AtLeastOneRequiredArgumentError.
	KindAtLeastOneRequiredArgument
	// KindUndefinedCommand maps to Thor::UndefinedCommandError.
	KindUndefinedCommand
	// KindAmbiguousCommand maps to Thor::AmbiguousCommandError.
	KindAmbiguousCommand
	// KindArgument maps to Ruby's ArgumentError (invalid declaration).
	KindArgument
)

// String returns the fully-qualified gem exception class name for the kind.
func (k ErrorKind) String() string {
	switch k {
	case KindRequiredArgumentMissing:
		return "Thor::RequiredArgumentMissingError"
	case KindMalformattedArgument:
		return "Thor::MalformattedArgumentError"
	case KindUnknownArgument:
		return "Thor::UnknownArgumentError"
	case KindExclusiveArgument:
		return "Thor::ExclusiveArgumentError"
	case KindAtLeastOneRequiredArgument:
		return "Thor::AtLeastOneRequiredArgumentError"
	case KindUndefinedCommand:
		return "Thor::UndefinedCommandError"
	case KindAmbiguousCommand:
		return "Thor::AmbiguousCommandError"
	case KindArgument:
		return "ArgumentError"
	}
	return "Thor::Error"
}

// Error is a Thor parse or dispatch error. Msg is byte-identical to the message
// the gem raises, and Kind names the gem exception class so the host can
// re-raise it faithfully.
type Error struct {
	Kind ErrorKind
	Msg  string
}

// Error implements the error interface, returning the gem-exact message.
func (e *Error) Error() string { return e.Msg }

func newError(kind ErrorKind, msg string) *Error { return &Error{Kind: kind, Msg: msg} }
