// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import "testing"

func TestRubyInspectEscapes(t *testing.T) {
	cases := map[string]string{
		"abc":    `"abc"`,
		`a"b`:    `"a\"b"`,
		`a\b`:    `"a\\b"`,
		"a\nb":   `"a\nb"`,
		"a\tb":   `"a\tb"`,
		"a\rb":   `"a\rb"`,
		"a\ab":   `"a\ab"`,
		"a\bb":   `"a\bb"`,
		"a\fb":   `"a\fb"`,
		"a\vb":   `"a\vb"`,
		"a\x00b": `"a\0b"`,
		"a\x1bb": `"a\eb"`,
		"héllo":  `"héllo"`,
	}
	for in, want := range cases {
		if got := rubyInspect(in); got != want {
			t.Errorf("rubyInspect(%q)=%q want %q", in, got, want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	cases := map[float64]string{
		3.0:   "3.0",
		3.5:   "3.5",
		100.0: "100.0",
		-2.5:  "-2.5",
		0.0:   "0.0",
	}
	for in, want := range cases {
		if got := formatFloat(in); got != want {
			t.Errorf("formatFloat(%v)=%q want %q", in, got, want)
		}
	}
}

func TestValueToDisplay(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hi", "hi"},
		{true, "true"},
		{false, "false"},
		{int64(7), "7"},
		{int(9), "9"},
		{2.5, "2.5"},
		{nil, ""},
		{[]string{"a"}, ""}, // unmapped type -> ""
	}
	for _, c := range cases {
		if got := valueToDisplay(c.in); got != c.want {
			t.Errorf("valueToDisplay(%#v)=%q want %q", c.in, got, c.want)
		}
	}
}
