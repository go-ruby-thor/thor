// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import "testing"

func TestNewOptionDefaultType(t *testing.T) {
	o, err := NewOption("name", Option{})
	if err != nil {
		t.Fatal(err)
	}
	if o.Typ != String {
		t.Fatalf("typ=%q", o.Typ)
	}
	if o.Banner != "NAME" { // string default banner is the upcased human name
		t.Fatalf("banner=%q", o.Banner)
	}
}

func TestNewOptionExplicitEmptyBannerKept(t *testing.T) {
	o, err := NewOption("name", Option{Typ: String, Banner: "", BannerSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if o.Banner != "" {
		t.Fatalf("banner=%q", o.Banner)
	}
}

func TestDasherizedName(t *testing.T) {
	o := mo(t, "--weird", Option{Typ: String})
	if o.SwitchName() != "--weird" {
		t.Fatalf("switch=%q", o.SwitchName())
	}
	if o.HumanName() != "weird" {
		t.Fatalf("human=%q", o.HumanName())
	}
}

func TestSwitchAndHumanNameUndasherized(t *testing.T) {
	o := mo(t, "my_flag", Option{Typ: String})
	if o.SwitchName() != "--my-flag" {
		t.Fatalf("switch=%q", o.SwitchName())
	}
	if o.HumanName() != "my_flag" {
		t.Fatalf("human=%q", o.HumanName())
	}
	// Single-char name dasherizes to a single dash.
	one := mo(t, "x", Option{Typ: String})
	if one.SwitchName() != "-x" {
		t.Fatalf("switch=%q", one.SwitchName())
	}
}

func TestUndasherize(t *testing.T) {
	if undasherize("--foo") != "foo" {
		t.Fatalf("got %q", undasherize("--foo"))
	}
}

func TestDefaultBanners(t *testing.T) {
	cases := map[Type]string{
		Boolean: "",
		Numeric: "N",
		Hash:    "key:value",
		Array:   "one two three",
	}
	for typ, want := range cases {
		o := mo(t, "opt", Option{Typ: typ})
		if o.Banner != want {
			t.Errorf("banner[%s]=%q want %q", typ, o.Banner, want)
		}
	}
	// String banner is the upcased human name.
	s := mo(t, "name", Option{Typ: String})
	if s.Banner != "NAME" {
		t.Fatalf("string banner=%q", s.Banner)
	}
}

func TestUsageForceNoExtraSwitches(t *testing.T) {
	f := mo(t, "force", Option{Typ: Boolean})
	if f.Usage(0) != "[--force]" {
		t.Fatalf("force=%q", f.Usage(0))
	}
}

func TestUsageNoPrefixNameNoExtra(t *testing.T) {
	np := mo(t, "no-color", Option{Typ: Boolean})
	if np.Usage(0) != "[--no-color]" {
		t.Fatalf("np=%q", np.Usage(0))
	}
	sk := mo(t, "skip_it", Option{Typ: Boolean})
	if sk.Usage(0) != "[--skip-it]" {
		t.Fatalf("sk=%q", sk.Usage(0))
	}
}

func TestUsageBooleanExpands(t *testing.T) {
	b := mo(t, "loud", Option{Typ: Boolean})
	if b.Usage(0) != "[--loud], [--no-loud], [--skip-loud]" {
		t.Fatalf("got %q", b.Usage(0))
	}
}

func TestUsageRequiredNotBracketed(t *testing.T) {
	r := mo(t, "name", Option{Typ: String, Required: true})
	if r.Usage(0) != "--name=NAME" {
		t.Fatalf("got %q", r.Usage(0))
	}
}

func TestUsageExplicitEmptyBanner(t *testing.T) {
	// Explicit empty banner -> sample is the bare switch, bracketed as optional.
	e := mo(t, "flag", Option{Typ: String, Banner: "", BannerSet: true})
	if e.Usage(0) != "[--flag]" {
		t.Fatalf("got %q", e.Usage(0))
	}
}

func TestUsagePadding(t *testing.T) {
	o := mo(t, "name", Option{Typ: String, Aliases: []string{"-n"}})
	// Padding widens the aliases column.
	if got := o.Usage(6); got != "-n,   [--name=NAME]" {
		t.Fatalf("got %q", got)
	}
}

func TestHasNoSkipPrefix(t *testing.T) {
	for _, p := range []string{"no-x", "no_x", "skip-x", "skip_x"} {
		if !hasNoSkipPrefix(p) {
			t.Errorf("%q should match", p)
		}
	}
	if hasNoSkipPrefix("plain") {
		t.Error("plain should not match")
	}
}

func TestLjust(t *testing.T) {
	if ljust("ab", 5) != "ab   " {
		t.Fatalf("got %q", ljust("ab", 5))
	}
	if ljust("abcde", 3) != "abcde" {
		t.Fatalf("got %q", ljust("abcde", 3))
	}
}

func TestShowDefault(t *testing.T) {
	cases := []struct {
		def  any
		want bool
	}{
		{true, true},
		{false, true},
		{nil, false},
		{"x", true},
		{"", false},
		{[]string{"a"}, true},
		{[]string{}, false},
		{int64(3), true},
	}
	for _, c := range cases {
		o := mo(t, "o", Option{Typ: String, Default: c.def})
		if got := o.ShowDefault(); got != c.want {
			t.Errorf("ShowDefault(%#v)=%v want %v", c.def, got, c.want)
		}
	}
	// Non-empty OrderedMap default shows; empty does not.
	m := NewOrderedMap()
	oEmpty := mo(t, "o", Option{Typ: Hash, Default: m})
	if oEmpty.ShowDefault() {
		t.Fatal("empty map should not show default")
	}
	m.Set("k", "v")
	oFull := mo(t, "o", Option{Typ: Hash, Default: m})
	if !oFull.ShowDefault() {
		t.Fatal("non-empty map should show default")
	}
}

func TestPrintDefault(t *testing.T) {
	arr := mo(t, "arr", Option{Typ: Array, Default: []string{"x", "y"}})
	if arr.PrintDefault() != `"x" "y"` {
		t.Fatalf("arr=%q", arr.PrintDefault())
	}
	// Array type with a non-[]string default falls through to valueToDisplay.
	weird := mo(t, "arr", Option{Typ: Array, Default: int64(5)})
	if weird.PrintDefault() != "5" {
		t.Fatalf("weird=%q", weird.PrintDefault())
	}
	num := mo(t, "n", Option{Typ: Numeric, Default: int64(7)})
	if num.PrintDefault() != "7" {
		t.Fatalf("num=%q", num.PrintDefault())
	}
	str := mo(t, "s", Option{Typ: String, Default: "hi"})
	if str.PrintDefault() != "hi" {
		t.Fatalf("str=%q", str.PrintDefault())
	}
}

func TestEnumToS(t *testing.T) {
	o := mo(t, "o", Option{Typ: String, Enum: []string{"a", "b", "c"}})
	if o.EnumToS() != "a, b, c" {
		t.Fatalf("got %q", o.EnumToS())
	}
}
