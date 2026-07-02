// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"reflect"
	"testing"
)

// TestShiftEmptyPile exercises shift() on an empty pile (Ruby Array#shift on []
// returns nil; here "").
func TestShiftEmptyPile(t *testing.T) {
	p := &parser{}
	if s := p.shift(); s != "" {
		t.Fatalf("shift on empty pile = %q, want %q", s, "")
	}
}

// TestAssignResultNilOption is a no-op guard (option not found in @switches).
func TestAssignResultNilOption(t *testing.T) {
	p := &parser{assigns: NewValueMap()}
	p.assignResult(nil, "x")
	if len(p.assigns.Keys()) != 0 {
		t.Fatalf("assignResult(nil) mutated assigns: %v", p.assigns.Keys())
	}
}

// TestCurrentIsSwitchTreatedAsValue: after an inline "=" value is unshifted as a
// value, the next classification must treat it as a non-switch.
func TestCurrentIsSwitchTreatedAsValue(t *testing.T) {
	p := &parser{isTreatedAsValue: true, pile: []string{"--x"}}
	match, isSwitch := p.currentIsSwitch()
	if match || isSwitch {
		t.Fatalf("isTreatedAsValue: got match=%v isSwitch=%v, want false,false", match, isSwitch)
	}
}

// TestCurrentIsValueAfterTerminator: with parsing_options off, current_is_value?
// is unconditionally true.
func TestCurrentIsValueAfterTerminator(t *testing.T) {
	p := &parser{parsingOptions: false, pile: []string{"--still-a-value"}}
	if !p.currentIsValue() {
		t.Fatalf("currentIsValue after terminator = false, want true")
	}
}

// TestSwitchOptionNoOrSkipDirect: a literal "--no-foo" switch that is itself
// registered wins over the "--foo" fallback.
func TestSwitchOptionNoOrSkipDirect(t *testing.T) {
	noFoo := opt(t, "no-foo", Option{Typ: Boolean})
	p := &parser{switches: map[string]*Option{"--no-foo": noFoo}}
	if got := p.switchOption("--no-foo"); got != noFoo {
		t.Fatalf("switchOption(--no-foo) = %v, want the direct registration", got)
	}
}

// TestParseArrayEmpty: an array switch followed by a token that looks like a
// switch (dash-led) but isn't switch-formatted causes parse_array to consume no
// values, yielding an empty (non-nil) slice (arr==nil arm).
func TestParseArrayEmpty(t *testing.T) {
	list := opt(t, "list", Option{Typ: Array})
	res := mustParse(t, []*Option{list}, []string{"--list", "--foo--bar"})
	v := getVal(t, res, "list")
	arr, ok := v.([]string)
	if !ok || arr == nil || len(arr) != 0 {
		t.Fatalf("empty array = %#v, want empty non-nil []string", v)
	}
}

// TestParseNoSkipStringBareReturnsNil: a "--no-<string>" switch given bare
// resolves through parse_peek's no_or_skip arm to a nil value (option cleared).
func TestParseNoSkipStringBareReturnsNil(t *testing.T) {
	// Register both "--foo" (string) and let "--no-foo" resolve to it. Given
	// "--no-foo" bare, parse_peek returns nil for the human name "foo".
	foo := opt(t, "foo", Option{Typ: String})
	res := mustParse(t, []*Option{foo}, []string{"--no-foo"})
	v, ok := res.Options.Get("foo")
	if !ok {
		t.Fatalf("foo missing after --no-foo; keys=%v", res.Options.Keys())
	}
	if v != nil {
		t.Fatalf("--no-foo bare: foo=%#v, want nil", v)
	}
}

// TestParseNoSkipStringWithValue: "--no-<string>" with a following value goes
// through parse_string's no_or_skip arm and yields nil (value not consumed).
func TestParseNoSkipStringWithValue(t *testing.T) {
	foo := opt(t, "foo", Option{Typ: String})
	res := mustParse(t, []*Option{foo}, []string{"--no-foo", "bar"})
	v, ok := res.Options.Get("foo")
	if !ok || v != nil {
		t.Fatalf("--no-foo bar: foo=%#v ok=%v, want nil", v, ok)
	}
	if !reflect.DeepEqual(res.Args, []string{"bar"}) {
		t.Fatalf("args=%v, want [bar]", res.Args)
	}
}

// TestParseNumericLazyDefaultBare: a numeric option with a lazy default, given
// bare at the end, uses the lazy default via parse_peek's lazy_default arm.
func TestParseNumericLazyDefaultBare(t *testing.T) {
	n := opt(t, "count", Option{Typ: Numeric, LazyDefault: int64(7)})
	res := mustParse(t, []*Option{n}, []string{"--count"})
	if v := getVal(t, res, "count"); v != int64(7) {
		t.Fatalf("--count bare with lazy default = %#v, want 7", v)
	}
}

// TestParseByTypeUnknownType: an Option carrying a bogus (non-standard) Typ
// falls through parse_by_type to a nil value (Type is a public string field, so
// callers can construct one outside the five valid types).
func TestParseByTypeUnknownType(t *testing.T) {
	p := &parser{}
	got, err := p.parseByType(Type("bogus"), "--x", &Option{Name: "x", Typ: Type("bogus")})
	if err != nil || got != nil {
		t.Fatalf("parseByType(bogus) = (%#v, %v), want (nil, nil)", got, err)
	}
}

// TestDefaultBannerUnknownType: defaultBanner on a bogus Typ returns "" via its
// final fallback.
func TestDefaultBannerUnknownType(t *testing.T) {
	o := &Option{Name: "x", Typ: Type("bogus")}
	if b := o.defaultBanner(); b != "" {
		t.Fatalf("defaultBanner(bogus) = %q, want %q", b, "")
	}
}

// TestRunUnknownSwitchThenArgs: an unknown switch-formatted token (match=true,
// isSwitch=false) is pushed to extra, then following non-dash args are swept in.
func TestRunUnknownSwitchThenArgs(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--unknown", "a", "b"})
	if !reflect.DeepEqual(res.Args, []string{"--unknown", "a", "b"}) {
		t.Fatalf("args=%v, want [--unknown a b]", res.Args)
	}
}
