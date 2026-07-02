// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"strings"
	"testing"
)

func mo(t *testing.T, name string, o Option) *Option {
	t.Helper()
	op, err := NewOption(name, o)
	if err != nil {
		t.Fatal(err)
	}
	return op
}

func TestBasenameDefault(t *testing.T) {
	b := NewBase("x", Config{})
	if b.basename() != "thor" {
		t.Fatalf("basename=%q", b.basename())
	}
	b2 := NewBase("x", Config{Basename: "app"})
	if b2.basename() != "app" {
		t.Fatalf("basename=%q", b2.basename())
	}
}

func TestHelpClassLevel(t *testing.T) {
	b := NewBase("oracle.rb", Config{Basename: "oracle.rb", TerminalWidth: 80})
	b.ClassOptions = []*Option{mo(t, "verbose", Option{Typ: Boolean, Desc: "Be verbose", Aliases: []string{"-v"}})}
	b.AddCommand(NewCommand("greet", "Greet someone by NAME", "greet NAME", nil))
	b.AddCommand(NewCommand("list", "List things", "list", nil))
	want := "Commands:\n" +
		"  oracle.rb greet NAME  # Greet someone by NAME\n" +
		"  oracle.rb list        # List things\n" +
		"\n" +
		"Options:\n" +
		"  -v, [--verbose], [--no-verbose], [--skip-verbose]  # Be verbose\n" +
		"\n"
	if got := b.Help(); got != want {
		t.Fatalf("Help mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestHelpPackageName(t *testing.T) {
	b := NewBase("x", Config{TerminalWidth: 80, PackageName: "MyApp"})
	b.AddCommand(NewCommand("go", "Go", "go", nil))
	if !strings.HasPrefix(b.Help(), "MyApp commands:\n") {
		t.Fatalf("Help=%q", b.Help())
	}
}

func TestHelpHidesHiddenAndBlankDesc(t *testing.T) {
	b := NewBase("x", Config{TerminalWidth: 80})
	hidden := NewCommand("secret", "Secret", "secret", nil)
	hidden.Hidden = true
	b.AddCommand(hidden)
	nodesc := NewCommand("bare", "", "bare", nil)
	b.AddCommand(nodesc)
	got := b.Help()
	if strings.Contains(got, "secret") {
		t.Fatalf("hidden shown: %q", got)
	}
	// A command with no description shows just its banner (no "#"); the table
	// still pads the first column so there is trailing whitespace.
	if !strings.Contains(got, "  thor bare") || strings.Contains(got, "bare  #") {
		t.Fatalf("bare row wrong: %q", got)
	}
}

func TestCommandHelpGrouped(t *testing.T) {
	b := NewBase("app", Config{Basename: "app", TerminalWidth: 80})
	b.ClassOptions = []*Option{
		mo(t, "verbose", Option{Typ: Boolean, Desc: "Verbose"}),
		mo(t, "env", Option{Typ: String, Desc: "Env", Group: "Runtime"}),
	}
	b.AddCommand(NewCommand("go", "Run it", "go", nil))
	got, err := b.CommandHelp("go")
	if err != nil {
		t.Fatal(err)
	}
	want := "Usage:\n" +
		"  app go\n" +
		"\n" +
		"Options:\n" +
		"  [--verbose], [--no-verbose], [--skip-verbose]  # Verbose\n" +
		"\n" +
		"Runtime options:\n" +
		"  [--env=ENV]  # Env\n" +
		"\n" +
		"Run it\n"
	if got != want {
		t.Fatalf("mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestCommandHelpLongDescWrapped(t *testing.T) {
	b := NewBase("app", Config{Basename: "app", TerminalWidth: 30})
	c := NewCommand("go", "short", "go", nil)
	c.LongDescription = "one two three four five six seven eight nine ten"
	b.AddCommand(c)
	got, _ := b.CommandHelp("go")
	if !strings.Contains(got, "Description:\n") {
		t.Fatalf("no Description block: %q", got)
	}
	// Wrapped to width 30 minus indent 2 -> multiple lines, each indented by 2.
	if !strings.Contains(got, "\n  one two three") {
		t.Fatalf("wrap wrong: %q", got)
	}
}

func TestCommandHelpLongDescNoWrap(t *testing.T) {
	b := NewBase("app", Config{Basename: "app", TerminalWidth: 20})
	c := NewCommand("go", "short", "go", nil)
	c.LongDescription = "keep this whole line intact regardless of width"
	c.WrapLongDescription = false
	b.AddCommand(c)
	got, _ := b.CommandHelp("go")
	if !strings.Contains(got, "keep this whole line intact regardless of width\n") {
		t.Fatalf("nowrap wrong: %q", got)
	}
}

func TestCommandHelpUnknownErrors(t *testing.T) {
	b := NewBase("app", Config{})
	b.AddCommand(NewCommand("real", "d", "real", nil))
	// A prefix with no match returns the token unchanged -> undefined command.
	_, err := b.CommandHelp("ghost")
	assertErr(t, err, KindUndefinedCommand,
		`Could not find command "ghost" in "app" namespace.`)
}

func TestCommandHelpAmbiguousErrors(t *testing.T) {
	b := NewBase("app", Config{})
	b.AddCommand(NewCommand("start", "d", "start", nil))
	b.AddCommand(NewCommand("stop", "d", "stop", nil))
	_, err := b.CommandHelp("st")
	assertErr(t, err, KindAmbiguousCommand, "Ambiguous command st matches [start, stop]")
}

func TestClassOptionsHelpEmpty(t *testing.T) {
	if out := classOptionsHelp(nil, 80); out != nil {
		t.Fatalf("out=%v", out)
	}
}

func TestPrintOptionsAllHidden(t *testing.T) {
	hidden := mo(t, "x", Option{Typ: Boolean, Hide: true})
	if out := printOptions([]*Option{hidden}, "", 80); out != nil {
		t.Fatalf("out=%v", out)
	}
}

func TestPrintOptionsEmpty(t *testing.T) {
	if out := printOptions(nil, "", 80); out != nil {
		t.Fatalf("out=%v", out)
	}
}

func TestCollapseWhitespace(t *testing.T) {
	cases := map[string]string{
		"a  b\tc\nd": "a b c d",
		" lead":      " lead",
		"trail ":     "trail ",
		"":           "",
		"plain":      "plain",
	}
	for in, want := range cases {
		if got := collapseWhitespace(in); got != want {
			t.Errorf("collapseWhitespace(%q)=%q want %q", in, got, want)
		}
	}
}

func TestHelpMultiWordDescCollapsed(t *testing.T) {
	b := NewBase("x", Config{TerminalWidth: 200})
	b.AddCommand(NewCommand("go", "multi\n\tline   desc", "go", nil))
	got := b.Help()
	if !strings.Contains(got, "# multi line desc") {
		t.Fatalf("desc not collapsed: %q", got)
	}
}
