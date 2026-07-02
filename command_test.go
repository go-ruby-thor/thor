// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import "testing"

func TestNewCommandDefaults(t *testing.T) {
	c := NewCommand("go", "Go", "go", nil)
	if !c.WrapLongDescription {
		t.Fatal("WrapLongDescription should default true")
	}
}

func TestFormattedUsageUseNamespace(t *testing.T) {
	token := opt(t, "token", Option{Typ: String, Required: true})
	c := NewCommand("push", "Push", "push URL", []*Option{token})
	got := c.FormattedUsage("app", nil, true, false)
	if got != "app:push URL --token=TOKEN" {
		t.Fatalf("got %q", got)
	}
}

func TestFormattedUsageNoNamespace(t *testing.T) {
	token := opt(t, "token", Option{Typ: String, Required: true})
	c := NewCommand("push", "Push", "push URL", []*Option{token})
	if got := c.FormattedUsage("app", nil, false, false); got != "push URL --token=TOKEN" {
		t.Fatalf("got %q", got)
	}
}

func TestFormattedUsageSubcommand(t *testing.T) {
	c := NewCommand("push", "Push", "push URL", nil)
	if got := c.FormattedUsage("app:sub", nil, false, true); got != "sub push URL" {
		t.Fatalf("got %q", got)
	}
}

func TestFormattedUsageAncestor(t *testing.T) {
	c := NewCommand("greet", "Greet", "greet NAME", nil)
	c.AncestorName = "parent"
	if got := c.FormattedUsage("app", nil, true, false); got != "parent greet NAME" {
		t.Fatalf("got %q", got)
	}
}

func TestFormattedUsageDefaultNamespaceStripped(t *testing.T) {
	c := NewCommand("greet", "Greet", "greet NAME", nil)
	if got := c.FormattedUsage("default:sub", nil, true, false); got != ":sub:greet NAME" {
		t.Fatalf("got %q", got)
	}
}

func TestFormattedUsageSortsRequiredOptions(t *testing.T) {
	token := opt(t, "token", Option{Typ: String, Required: true})
	aaa := opt(t, "aaa", Option{Typ: String, Required: true})
	optional := opt(t, "verbose", Option{Typ: Boolean})
	c := NewCommand("push", "Push", "push", []*Option{token, aaa, optional})
	// Required options are sorted; the optional one is omitted.
	if got := c.FormattedUsage("app", nil, false, false); got != "push --aaa=AAA --token=TOKEN" {
		t.Fatalf("got %q", got)
	}
}

func TestRequiredArgumentsForInjection(t *testing.T) {
	name := opt(t, "name", Option{Typ: String, Required: true})
	title := opt(t, "title", Option{Typ: String})
	c := NewCommand("greet", "Greet", "greet MORE", nil)
	got := c.FormattedUsage("app", []*Option{name, title}, false, false)
	if got != "greet NAME [TITLE] MORE" {
		t.Fatalf("got %q", got)
	}
}

func TestRequiredArgumentsForNoLeadingName(t *testing.T) {
	// When the usage does not begin with the command name, injection is skipped
	// and the raw usage is returned.
	name := opt(t, "name", Option{Typ: String, Required: true})
	c := NewCommand("greet", "Greet", "hello world", nil)
	got := c.FormattedUsage("app", []*Option{name}, false, false)
	if got != "hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestArgumentUsageOptionalBracketed(t *testing.T) {
	req := opt(t, "name", Option{Typ: String, Required: true})
	optn := opt(t, "title", Option{Typ: String})
	if u := argumentUsage(req); u != "NAME" {
		t.Fatalf("req=%q", u)
	}
	if u := argumentUsage(optn); u != "[TITLE]" {
		t.Fatalf("opt=%q", u)
	}
}

func TestRequiredArgumentsForEmptyBannerSkipped(t *testing.T) {
	// An argument whose usage renders empty (explicit empty required banner) is
	// dropped from the injected list.
	blank := opt(t, "flag", Option{Typ: String, Required: true, Banner: "", BannerSet: true})
	c := NewCommand("go", "Go", "go", nil)
	got := c.FormattedUsage("app", []*Option{blank}, false, false)
	if got != "go" {
		t.Fatalf("got %q", got)
	}
}
