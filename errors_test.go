// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import "testing"

func TestErrorKindString(t *testing.T) {
	cases := map[ErrorKind]string{
		KindRequiredArgumentMissing:    "Thor::RequiredArgumentMissingError",
		KindMalformattedArgument:       "Thor::MalformattedArgumentError",
		KindUnknownArgument:            "Thor::UnknownArgumentError",
		KindExclusiveArgument:          "Thor::ExclusiveArgumentError",
		KindAtLeastOneRequiredArgument: "Thor::AtLeastOneRequiredArgumentError",
		KindUndefinedCommand:           "Thor::UndefinedCommandError",
		KindAmbiguousCommand:           "Thor::AmbiguousCommandError",
		KindArgument:                   "ArgumentError",
		ErrorKind(999):                 "Thor::Error",
	}
	for k, want := range cases {
		if got := k.String(); got != want {
			t.Errorf("%d.String()=%q want %q", int(k), got, want)
		}
	}
}

func TestErrorError(t *testing.T) {
	e := newError(KindMalformattedArgument, "boom")
	if e.Error() != "boom" {
		t.Fatalf("Error()=%q", e.Error())
	}
	if e.Kind != KindMalformattedArgument {
		t.Fatalf("kind=%v", e.Kind)
	}
}
