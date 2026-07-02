// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"reflect"
	"testing"
)

func mkCmd(name, desc, usage string, opts []*Option) *Command {
	return NewCommand(name, desc, usage, opts)
}

func newTestBase() *Base {
	b := NewBase("myapp", Config{Basename: "myapp"})
	b.AddCommand(mkCmd("install", "Install it", "install", nil))
	b.AddCommand(mkCmd("init", "Init it", "init", nil))
	b.AddCommand(mkCmd("list", "List it", "list", nil))
	return b
}

func TestAddCommandNormalizesAndReplaces(t *testing.T) {
	b := NewBase("x", Config{})
	b.AddCommand(mkCmd("foo-bar", "d1", "foo-bar", nil))
	if b.Commands()[0].Name != "foo_bar" {
		t.Fatalf("name=%q", b.Commands()[0].Name)
	}
	// Re-adding the same key replaces in place, not appends.
	b.AddCommand(mkCmd("foo_bar", "d2", "foo_bar", nil))
	if len(b.Commands()) != 1 {
		t.Fatalf("commands=%d", len(b.Commands()))
	}
	if b.byName["foo_bar"].Description != "d2" {
		t.Fatalf("desc=%q", b.byName["foo_bar"].Description)
	}
}

func TestNormalizeCommandNameDefault(t *testing.T) {
	b := NewBase("x", Config{})
	name, err := b.NormalizeCommandName("")
	if err != nil || name != "help" {
		t.Fatalf("name=%q err=%v", name, err)
	}
	b.DefaultCommand = "do-thing"
	name, _ = b.NormalizeCommandName("")
	if name != "do_thing" {
		t.Fatalf("name=%q", name)
	}
}

func TestNormalizeCommandNamePrefix(t *testing.T) {
	b := newTestBase()
	name, err := b.NormalizeCommandName("ins")
	if err != nil || name != "install" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestNormalizeCommandNameExactWins(t *testing.T) {
	b := NewBase("x", Config{})
	b.AddCommand(mkCmd("in", "d", "in", nil))
	b.AddCommand(mkCmd("index", "d", "index", nil))
	// Exact match "in" resolves to itself even though "index" also has prefix.
	name, err := b.NormalizeCommandName("in")
	if err != nil || name != "in" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestNormalizeCommandNameAmbiguous(t *testing.T) {
	b := NewBase("x", Config{})
	b.AddCommand(mkCmd("start", "d", "start", nil))
	b.AddCommand(mkCmd("stop", "d", "stop", nil))
	_, err := b.NormalizeCommandName("st")
	assertErr(t, err, KindAmbiguousCommand, "Ambiguous command st matches [start, stop]")
}

func TestNormalizeCommandNameNoMatch(t *testing.T) {
	b := newTestBase()
	// No prefix match: returns the token unchanged (dashes normalized).
	name, err := b.NormalizeCommandName("zzz-thing")
	if err != nil || name != "zzz_thing" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestNormalizeCommandNameMapAlias(t *testing.T) {
	b := newTestBase()
	b.Map = map[string]string{"i": "install"}
	name, err := b.NormalizeCommandName("i")
	if err != nil || name != "install" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestNormalizeCommandNameHiddenSkipped(t *testing.T) {
	b := NewBase("x", Config{})
	hidden := mkCmd("secret", "d", "secret", nil)
	hidden.Hidden = true
	b.AddCommand(hidden)
	b.AddCommand(mkCmd("show", "d", "show", nil))
	// "s" only matches the visible "show".
	name, err := b.NormalizeCommandName("s")
	if err != nil || name != "show" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestFindCommandPossibilitiesMapUnique(t *testing.T) {
	b := NewBase("x", Config{})
	b.AddCommand(mkCmd("install", "d", "install", nil))
	// Two map keys both resolving to the same command collapse to one.
	b.Map = map[string]string{"ins1": "install", "ins2": "install"}
	poss := b.findCommandPossibilities("ins")
	if !reflect.DeepEqual(poss, []string{"install"}) {
		t.Fatalf("poss=%v", poss)
	}
}

func TestRetrieveCommandName(t *testing.T) {
	b := newTestBase()
	// Empty argv.
	args := []string{}
	if got := b.retrieveCommandName(&args); got != "" {
		t.Fatalf("empty -> %q", got)
	}
	// Plain first token consumed.
	args = []string{"install", "--x"}
	if got := b.retrieveCommandName(&args); got != "install" || !reflect.DeepEqual(args, []string{"--x"}) {
		t.Fatalf("got=%q args=%v", got, args)
	}
	// Leading switch: not consumed.
	args = []string{"--x", "y"}
	if got := b.retrieveCommandName(&args); got != "" || !reflect.DeepEqual(args, []string{"--x", "y"}) {
		t.Fatalf("got=%q args=%v", got, args)
	}
	// Map alias token consumed even though not a plain word.
	b.Map = map[string]string{"-i": "install"}
	args = []string{"-i"}
	if got := b.retrieveCommandName(&args); got != "-i" || len(args) != 0 {
		t.Fatalf("got=%q args=%v", got, args)
	}
}

func TestSplit(t *testing.T) {
	a, r := Split([]string{"a", "b", "--x", "c"})
	if !reflect.DeepEqual(a, []string{"a", "b"}) || !reflect.DeepEqual(r, []string{"--x", "c"}) {
		t.Fatalf("a=%v r=%v", a, r)
	}
	a, r = Split([]string{"--x"})
	if len(a) != 0 || !reflect.DeepEqual(r, []string{"--x"}) {
		t.Fatalf("a=%v r=%v", a, r)
	}
	a, r = Split([]string{"a", "b"})
	if !reflect.DeepEqual(a, []string{"a", "b"}) || len(r) != 0 {
		t.Fatalf("a=%v r=%v", a, r)
	}
}

func TestDispatchResolvesCommandAndOptions(t *testing.T) {
	b := NewBase("myapp", Config{Basename: "myapp"})
	nameOpt := opt(t, "name", Option{Typ: String})
	b.AddCommand(mkCmd("greet", "Greet", "greet", []*Option{nameOpt}))
	cmd, res, err := b.Dispatch([]string{"greet", "pos", "--name", "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Name != "greet" {
		t.Fatalf("cmd=%q", cmd.Name)
	}
	if v, _ := res.Options.Get("name"); v != "Bob" {
		t.Fatalf("name=%v", v)
	}
	if !reflect.DeepEqual(res.Args, []string{"pos"}) {
		t.Fatalf("args=%v", res.Args)
	}
}

func TestDispatchDefaultCommand(t *testing.T) {
	b := NewBase("myapp", Config{})
	b.AddCommand(mkCmd("help", "Help", "help", nil))
	cmd, _, err := b.Dispatch([]string{})
	if err != nil || cmd.Name != "help" {
		t.Fatalf("cmd=%v err=%v", cmd, err)
	}
}

func TestDispatchUndefinedCommand(t *testing.T) {
	b := NewBase("myapp", Config{})
	b.AddCommand(mkCmd("real", "d", "real", nil))
	_, _, err := b.Dispatch([]string{"ghost"})
	assertErr(t, err, KindUndefinedCommand,
		`Could not find command "ghost" in "myapp" namespace.`)
}

func TestDispatchAmbiguousPropagates(t *testing.T) {
	b := NewBase("myapp", Config{})
	b.AddCommand(mkCmd("start", "d", "start", nil))
	b.AddCommand(mkCmd("stop", "d", "stop", nil))
	_, _, err := b.Dispatch([]string{"st"})
	assertErr(t, err, KindAmbiguousCommand, "Ambiguous command st matches [start, stop]")
}

func TestDispatchParseErrorPropagates(t *testing.T) {
	b := NewBase("myapp", Config{})
	n := opt(t, "count", Option{Typ: Numeric})
	b.AddCommand(mkCmd("run", "d", "run", []*Option{n}))
	_, _, err := b.Dispatch([]string{"run", "--count", "abc"})
	assertErr(t, err, KindMalformattedArgument, `Expected numeric value for '--count'; got "abc"`)
}
