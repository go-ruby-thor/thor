// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"reflect"
	"testing"
)

// opt is a test helper that builds an option or fails the test.
func opt(t *testing.T, name string, o Option) *Option {
	t.Helper()
	op, err := NewOption(name, o)
	if err != nil {
		t.Fatalf("NewOption(%q): %v", name, err)
	}
	return op
}

// parse builds an Options parser over opts and parses argv.
func parse(t *testing.T, opts []*Option, argv []string) (*Result, error) {
	t.Helper()
	return NewOptions(opts, nil, false, false, Relations{}).Parse(argv)
}

func mustParse(t *testing.T, opts []*Option, argv []string) *Result {
	t.Helper()
	res, err := parse(t, opts, argv)
	if err != nil {
		t.Fatalf("parse %v: %v", argv, err)
	}
	return res
}

func getVal(t *testing.T, res *Result, human string) any {
	t.Helper()
	v, ok := res.Options.Get(human)
	if !ok {
		t.Fatalf("option %q missing; keys=%v", human, res.Options.Keys())
	}
	return v
}

func TestParseString(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--name", "David", "extra"})
	if v := getVal(t, res, "name"); v != "David" {
		t.Fatalf("name=%v", v)
	}
	if !reflect.DeepEqual(res.Args, []string{"extra"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseStringEquals(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--name=David"})
	if v := getVal(t, res, "name"); v != "David" {
		t.Fatalf("name=%v", v)
	}
}

func TestParseStringLazyDefaultOnBareSwitch(t *testing.T) {
	// Optional string given bare -> HumanName (Thor's :string fallback).
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--name"})
	if v := getVal(t, res, "name"); v != "name" {
		t.Fatalf("name=%v", v)
	}
}

func TestParseStringLazyDefaultValue(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, LazyDefault: "lazy"})
	res := mustParse(t, []*Option{name}, []string{"--name"})
	if v := getVal(t, res, "name"); v != "lazy" {
		t.Fatalf("name=%v", v)
	}
}

func TestParseStringDefaultOnBareSwitch(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Default: "dfl"})
	res := mustParse(t, []*Option{name}, []string{"--name"})
	if v := getVal(t, res, "name"); v != "dfl" {
		t.Fatalf("name=%v", v)
	}
}

func TestParseRequiredStringMissingValue(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Required: true})
	_, err := parse(t, []*Option{name}, []string{"--name"})
	assertErr(t, err, KindMalformattedArgument, "No value provided for option '--name'")
}

func TestParseNumericInt(t *testing.T) {
	n := opt(t, "count", Option{Typ: Numeric})
	res := mustParse(t, []*Option{n}, []string{"--count", "42"})
	if v := getVal(t, res, "count"); v != int64(42) {
		t.Fatalf("count=%v (%T)", v, v)
	}
}

func TestParseNumericFloat(t *testing.T) {
	n := opt(t, "ratio", Option{Typ: Numeric})
	res := mustParse(t, []*Option{n}, []string{"--ratio", "3.5"})
	if v := getVal(t, res, "ratio"); v != 3.5 {
		t.Fatalf("ratio=%v", v)
	}
}

func TestParseNumericShortMerged(t *testing.T) {
	n := opt(t, "count", Option{Typ: Numeric, Aliases: []string{"-c"}})
	res := mustParse(t, []*Option{n}, []string{"-c5"})
	if v := getVal(t, res, "count"); v != int64(5) {
		t.Fatalf("count=%v", v)
	}
}

func TestParseNumericMalformatted(t *testing.T) {
	n := opt(t, "count", Option{Typ: Numeric})
	_, err := parse(t, []*Option{n}, []string{"--count", "abc"})
	assertErr(t, err, KindMalformattedArgument, `Expected numeric value for '--count'; got "abc"`)
}

func TestParseBoolean(t *testing.T) {
	b := opt(t, "loud", Option{Typ: Boolean})
	res := mustParse(t, []*Option{b}, []string{"--loud"})
	if v := getVal(t, res, "loud"); v != true {
		t.Fatalf("loud=%v", v)
	}
}

func TestParseBooleanNoPrefix(t *testing.T) {
	b := opt(t, "loud", Option{Typ: Boolean})
	res := mustParse(t, []*Option{b}, []string{"--no-loud"})
	if v := getVal(t, res, "loud"); v != false {
		t.Fatalf("loud=%v", v)
	}
}

func TestParseBooleanSkipPrefix(t *testing.T) {
	b := opt(t, "loud", Option{Typ: Boolean})
	res := mustParse(t, []*Option{b}, []string{"--skip-loud"})
	if v := getVal(t, res, "loud"); v != false {
		t.Fatalf("loud=%v", v)
	}
}

func TestParseBooleanExplicitValues(t *testing.T) {
	b := opt(t, "loud", Option{Typ: Boolean})
	for _, tc := range []struct {
		arg  string
		want bool
	}{
		{"true", true}, {"TRUE", true}, {"t", true}, {"T", true},
		{"false", false}, {"FALSE", false}, {"f", false}, {"F", false},
	} {
		res := mustParse(t, []*Option{b}, []string{"--loud", tc.arg})
		if v := getVal(t, res, "loud"); v != tc.want {
			t.Fatalf("loud %s -> %v", tc.arg, v)
		}
	}
}

func TestParseBooleanFollowedByNonBool(t *testing.T) {
	// A non-true/false value after a bool switch yields true and stays in args.
	b := opt(t, "loud", Option{Typ: Boolean})
	res := mustParse(t, []*Option{b}, []string{"--loud", "word"})
	if v := getVal(t, res, "loud"); v != true {
		t.Fatalf("loud=%v", v)
	}
	if !reflect.DeepEqual(res.Args, []string{"word"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseBooleanNoPrefixWithValueToken(t *testing.T) {
	// --no-loud followed by a value token: parseBoolean's currentIsValue path
	// with a non-recognized token and a no/skip switch returns false.
	b := opt(t, "loud", Option{Typ: Boolean})
	res := mustParse(t, []*Option{b}, []string{"--no-loud", "word"})
	if v := getVal(t, res, "loud"); v != false {
		t.Fatalf("loud=%v", v)
	}
}

func TestParseArray(t *testing.T) {
	a := opt(t, "files", Option{Typ: Array})
	res := mustParse(t, []*Option{a}, []string{"--files", "a", "b", "c"})
	if v := getVal(t, res, "files"); !reflect.DeepEqual(v, []string{"a", "b", "c"}) {
		t.Fatalf("files=%v", v)
	}
}

func TestParseArrayBareErrors(t *testing.T) {
	// Real Thor: a bare non-string switch with no value raises
	// "No value provided for option '<switch>'" (parse_peek default branch).
	a := opt(t, "files", Option{Typ: Array})
	_, err := parse(t, []*Option{a}, []string{"--files"})
	assertErr(t, err, KindMalformattedArgument, "No value provided for option '--files'")
}

func TestParseArrayEmptyFromEquals(t *testing.T) {
	// --files= yields a single empty-string token, so parse_array collects one
	// empty value (which is skipped by the enum guard) into a one-element slice.
	a := opt(t, "files", Option{Typ: Array})
	res := mustParse(t, []*Option{a}, []string{"--files="})
	if v := getVal(t, res, "files"); !reflect.DeepEqual(v, []string{""}) {
		t.Fatalf("files=%#v", v)
	}
}

func TestParseArrayEnum(t *testing.T) {
	a := opt(t, "colors", Option{Typ: Array, Enum: []string{"red", "green"}})
	res := mustParse(t, []*Option{a}, []string{"--colors", "red", "green"})
	if v := getVal(t, res, "colors"); !reflect.DeepEqual(v, []string{"red", "green"}) {
		t.Fatalf("colors=%v", v)
	}
	_, err := parse(t, []*Option{a}, []string{"--colors", "blue"})
	assertErr(t, err, KindMalformattedArgument,
		"Expected all values of '--colors' to be one of red, green; got blue")
}

func TestParseHash(t *testing.T) {
	h := opt(t, "meta", Option{Typ: Hash})
	res := mustParse(t, []*Option{h}, []string{"--meta", "a:1", "b:2"})
	m, ok := getVal(t, res, "meta").(*OrderedMap)
	if !ok {
		t.Fatalf("not a map: %T", getVal(t, res, "meta"))
	}
	if !reflect.DeepEqual(m.Keys(), []string{"a", "b"}) {
		t.Fatalf("keys=%v", m.Keys())
	}
	if v, _ := m.Get("b"); v != "2" {
		t.Fatalf("b=%v", v)
	}
}

func TestParseHashStopsAtNonColon(t *testing.T) {
	h := opt(t, "meta", Option{Typ: Hash})
	res := mustParse(t, []*Option{h}, []string{"--meta", "a:1", "plain"})
	m := getVal(t, res, "meta").(*OrderedMap)
	if m.Len() != 1 {
		t.Fatalf("len=%d", m.Len())
	}
	if !reflect.DeepEqual(res.Args, []string{"plain"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseHashDuplicateKey(t *testing.T) {
	h := opt(t, "meta", Option{Typ: Hash})
	_, err := parse(t, []*Option{h}, []string{"--meta", "a:1", "a:2"})
	assertErr(t, err, KindMalformattedArgument,
		"You can't specify 'a' more than once in option '--meta'; got a:1 and a:2")
}

func TestParseEnumStringOK(t *testing.T) {
	s := opt(t, "mode", Option{Typ: String, Enum: []string{"fast", "slow"}})
	res := mustParse(t, []*Option{s}, []string{"--mode", "fast"})
	if v := getVal(t, res, "mode"); v != "fast" {
		t.Fatalf("mode=%v", v)
	}
}

func TestParseEnumStringBad(t *testing.T) {
	s := opt(t, "mode", Option{Typ: String, Enum: []string{"fast", "slow"}})
	_, err := parse(t, []*Option{s}, []string{"--mode", "medium"})
	assertErr(t, err, KindMalformattedArgument,
		"Expected '--mode' to be one of fast, slow; got medium")
}

func TestParseEnumNumericBad(t *testing.T) {
	n := opt(t, "count", Option{Typ: Numeric, Enum: []string{"1", "2"}})
	_, err := parse(t, []*Option{n}, []string{"--count", "5"})
	assertErr(t, err, KindMalformattedArgument,
		"Expected '--count' to be one of 1, 2; got 5")
}

func TestParseEnumNumericOK(t *testing.T) {
	n := opt(t, "count", Option{Typ: Numeric, Enum: []string{"1", "2"}})
	res := mustParse(t, []*Option{n}, []string{"--count", "2"})
	if v := getVal(t, res, "count"); v != int64(2) {
		t.Fatalf("count=%v", v)
	}
}

func TestParseAliases(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Aliases: []string{"-n", "alias"}})
	res := mustParse(t, []*Option{name}, []string{"-n", "Bob"})
	if v := getVal(t, res, "name"); v != "Bob" {
		t.Fatalf("name=%v", v)
	}
	// The "alias" alias is normalized to "-alias".
	if name.Aliases[1] != "-alias" {
		t.Fatalf("aliases=%v", name.Aliases)
	}
}

func TestParseShortSquashed(t *testing.T) {
	a := opt(t, "a", Option{Typ: Boolean, Aliases: []string{"-a"}})
	b := opt(t, "b", Option{Typ: Boolean, Aliases: []string{"-b"}})
	res := mustParse(t, []*Option{a, b}, []string{"-ab"})
	if getVal(t, res, "a") != true || getVal(t, res, "b") != true {
		t.Fatalf("a=%v b=%v", getVal(t, res, "a"), getVal(t, res, "b"))
	}
}

func TestParseShortSquashedUnknownGoesToExtra(t *testing.T) {
	a := opt(t, "a", Option{Typ: Boolean, Aliases: []string{"-a"}})
	// -xy has no matching switch -> treated as extra, not expanded.
	res := mustParse(t, []*Option{a}, []string{"-xy"})
	if !reflect.DeepEqual(res.Args, []string{"-xy"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseDoubleDashTerminator(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--name", "x", "--", "--name", "y"})
	if v := getVal(t, res, "name"); v != "x" {
		t.Fatalf("name=%v", v)
	}
	if !reflect.DeepEqual(res.Args, []string{"--name", "y"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseDoubleDashAtEnd(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--name", "x", "--"})
	if len(res.Args) != 0 {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseUnknownSwitchKept(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"--unknown", "--name", "x"})
	if v := getVal(t, res, "name"); v != "x" {
		t.Fatalf("name=%v", v)
	}
	if !reflect.DeepEqual(res.Args, []string{"--unknown"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseStopOnUnknown(t *testing.T) {
	name := opt(t, "name", Option{Typ: String})
	o := NewOptions([]*Option{name}, nil, true, false, Relations{})
	res, err := o.Parse([]string{"cmd", "--name", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res.Args, []string{"cmd", "--name", "x"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseNonOptionRunAfterMatch(t *testing.T) {
	// A leading non-switch token then trailing plain words: the check_unknown
	// "match" branch swallows following non-dash tokens into extra.
	name := opt(t, "name", Option{Typ: String})
	res := mustParse(t, []*Option{name}, []string{"plain1", "plain2", "--name", "x"})
	if v := getVal(t, res, "name"); v != "x" {
		t.Fatalf("name=%v", v)
	}
	if !reflect.DeepEqual(res.Args, []string{"plain1", "plain2"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestParseRequiredMissing(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Required: true})
	_, err := parse(t, []*Option{name}, []string{})
	assertErr(t, err, KindRequiredArgumentMissing,
		"No value provided for required options '--name'")
}

func TestParseRequiredProvided(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Required: true})
	res := mustParse(t, []*Option{name}, []string{"--name", "x"})
	if v := getVal(t, res, "name"); v != "x" {
		t.Fatalf("name=%v", v)
	}
}

func TestParseDisableRequiredCheck(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Required: true})
	o := NewOptions([]*Option{name}, nil, false, true, Relations{})
	if _, err := o.Parse([]string{}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestParseDefaultsHashClearsRequired(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Required: true})
	dfl := NewValueMap()
	dfl.Set("name", "seeded")
	o := NewOptions([]*Option{name}, dfl, false, false, Relations{})
	res, err := o.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if v, _ := res.Options.Get("name"); v != "seeded" {
		t.Fatalf("name=%v", v)
	}
}

func TestParseExclusive(t *testing.T) {
	// Multi-char names so "--foo"/"--bar" actually match their switches and the
	// exclusive check fires (single-char names have switch "-x" and "--x" is an
	// unknown long switch, so the assigns hash would stay empty).
	a := opt(t, "foo", Option{Typ: Boolean})
	b := opt(t, "bar", Option{Typ: Boolean})
	rel := Relations{Exclusive: [][]string{{"foo", "bar"}}}
	o := NewOptions([]*Option{a, b}, nil, false, false, rel)
	_, err := o.Parse([]string{"--foo", "--bar"})
	assertErr(t, err, KindExclusiveArgument, "Found exclusive options '--foo', '--bar'")
}

func TestParseExclusiveOK(t *testing.T) {
	a := opt(t, "foo", Option{Typ: Boolean})
	b := opt(t, "bar", Option{Typ: Boolean})
	rel := Relations{Exclusive: [][]string{{"foo", "bar"}}}
	o := NewOptions([]*Option{a, b}, nil, false, false, rel)
	if _, err := o.Parse([]string{"--foo"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestParseAtLeastOne(t *testing.T) {
	a := opt(t, "foo", Option{Typ: Boolean})
	b := opt(t, "bar", Option{Typ: Boolean})
	rel := Relations{AtLeastOne: [][]string{{"foo", "bar"}}}
	o := NewOptions([]*Option{a, b}, nil, false, false, rel)
	_, err := o.Parse([]string{})
	assertErr(t, err, KindAtLeastOneRequiredArgument,
		"Not found at least one of required options '--foo', '--bar'")
}

func TestParseAtLeastOneOK(t *testing.T) {
	a := opt(t, "foo", Option{Typ: Boolean})
	b := opt(t, "bar", Option{Typ: Boolean})
	rel := Relations{AtLeastOne: [][]string{{"foo", "bar"}}}
	o := NewOptions([]*Option{a, b}, nil, false, false, rel)
	if _, err := o.Parse([]string{"--bar"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestParseRepeatableArray(t *testing.T) {
	a := opt(t, "tag", Option{Typ: String, Repeatable: true})
	res := mustParse(t, []*Option{a}, []string{"--tag", "x", "--tag", "y"})
	v := getVal(t, res, "tag")
	arr, ok := v.([]any)
	if !ok || len(arr) != 2 || arr[0] != "x" || arr[1] != "y" {
		t.Fatalf("tag=%#v", v)
	}
}

func TestParseRepeatableHash(t *testing.T) {
	a := opt(t, "meta", Option{Typ: Hash, Repeatable: true})
	res := mustParse(t, []*Option{a}, []string{"--meta", "a:1", "--meta", "b:2"})
	m := getVal(t, res, "meta").(*OrderedMap)
	if !reflect.DeepEqual(m.Keys(), []string{"a", "b"}) {
		t.Fatalf("keys=%v", m.Keys())
	}
}

func TestNewOptionBooleanRequiredRejected(t *testing.T) {
	_, err := NewOption("x", Option{Typ: Boolean, Required: true})
	assertErr(t, err, KindArgument, "An option cannot be boolean and required.")
}

func TestSelectNonEmptyDropsEmptyGroups(t *testing.T) {
	rel := Relations{Exclusive: [][]string{{}, {"a"}}, AtLeastOne: [][]string{{}, {"b"}}}
	out := selectNonEmpty(rel)
	if len(out.Exclusive) != 1 || len(out.AtLeastOne) != 1 {
		t.Fatalf("out=%#v", out)
	}
}

func TestLooksLikeUnknownSwitch(t *testing.T) {
	cases := map[string]bool{
		"--foo":     true,
		"-f":        true,
		"plain":     false,
		"--a--b":    false, // an interior "--" run disqualifies it
		"---triple": true,  // "---triple": after two dashes, "-triple" has no "--"
	}
	for in, want := range cases {
		if got := looksLikeUnknownSwitch(in); got != want {
			t.Errorf("looksLikeUnknownSwitch(%q)=%v want %v", in, got, want)
		}
	}
}

func TestValueLooksLikeSwitch(t *testing.T) {
	cases := map[string]bool{
		"plain":  false,
		"--foo":  true,
		"-f":     true,
		"---bar": true,
	}
	for in, want := range cases {
		if got := valueLooksLikeSwitch(in); got != want {
			t.Errorf("valueLooksLikeSwitch(%q)=%v want %v", in, got, want)
		}
	}
}

func TestSprintfMsgNoPlaceholder(t *testing.T) {
	if got := sprintfMsg("no placeholders", "a", "b", "c"); got != "no placeholders" {
		t.Fatalf("got %q", got)
	}
}

func TestContainsHelpers(t *testing.T) {
	if !contains([]string{"a", "b"}, "b") || contains([]string{"a"}, "z") {
		t.Fatal("contains broken")
	}
	if !reflect.DeepEqual(subtract([]string{"a", "b"}, []string{"b"}), []string{"a"}) {
		t.Fatal("subtract broken")
	}
	if !reflect.DeepEqual(intersect([]string{"a", "b"}, []string{"b", "c"}), []string{"b"}) {
		t.Fatal("intersect broken")
	}
}

func TestValidateEnumNilOption(t *testing.T) {
	if err := validateEnum(nil, "x", "y", "msg"); err != nil {
		t.Fatal(err)
	}
	if err := validateEnumValue(nil, "x", int64(1), "msg"); err != nil {
		t.Fatal(err)
	}
}

// assertErr checks a *Error's kind and message.
func assertErr(t *testing.T, err error, kind ErrorKind, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %q, got nil", msg)
	}
	te, ok := err.(*Error)
	if !ok {
		t.Fatalf("not a *Error: %T", err)
	}
	if te.Kind != kind {
		t.Fatalf("kind=%v want %v (msg=%q)", te.Kind, kind, te.Msg)
	}
	if te.Msg != msg {
		t.Fatalf("msg=%q want %q", te.Msg, msg)
	}
}
