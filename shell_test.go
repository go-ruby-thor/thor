// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"reflect"
	"testing"
)

func TestTerminalWidth(t *testing.T) {
	noenv := func(string) string { return "" }
	if w := terminalWidth(0, noenv); w != 80 {
		t.Fatalf("default=%d", w)
	}
	if w := terminalWidth(120, noenv); w != 120 {
		t.Fatalf("explicit=%d", w)
	}
	// Below-10 explicit width falls back to the default.
	if w := terminalWidth(5, noenv); w != 80 {
		t.Fatalf("tiny=%d", w)
	}
	env := func(k string) string {
		if k == "THOR_COLUMNS" {
			return "100"
		}
		return ""
	}
	if w := terminalWidth(0, env); w != 100 {
		t.Fatalf("env=%d", w)
	}
	// THOR_COLUMNS below 10 also falls back.
	envSmall := func(string) string { return "3" }
	if w := terminalWidth(0, envSmall); w != 80 {
		t.Fatalf("envsmall=%d", w)
	}
}

func TestWidthNilGetenv(t *testing.T) {
	b := NewBase("x", Config{})
	if b.width() != 80 {
		t.Fatalf("width=%d", b.width())
	}
	b2 := NewBase("x", Config{Getenv: func(k string) string {
		if k == "THOR_COLUMNS" {
			return "40"
		}
		return ""
	}})
	if b2.width() != 40 {
		t.Fatalf("width=%d", b2.width())
	}
}

func TestAtoiRuby(t *testing.T) {
	cases := map[string]int{
		"42":     42,
		"  42  ": 42,
		"+7":     7,
		"-7":     -7,
		"12abc":  12,
		"abc":    0,
		"":       0,
		"-":      0,
		"+":      0,
	}
	for in, want := range cases {
		if got := atoiRuby(in); got != want {
			t.Errorf("atoiRuby(%q)=%d want %d", in, got, want)
		}
	}
}

func TestPrintTableEmpty(t *testing.T) {
	if out := printTable(nil, 2, 0); out != nil {
		t.Fatalf("out=%v", out)
	}
}

func TestPrintTableRagged(t *testing.T) {
	// Rows of differing column counts: colcount is the widest row.
	rows := [][]string{{"a", "b", "c"}, {"x"}}
	out := printTable(rows, 0, 0)
	if len(out) != 2 {
		t.Fatalf("out=%v", out)
	}
}

func TestPrintTableTruncate(t *testing.T) {
	rows := [][]string{{"a very long cell that exceeds the width limit here"}}
	out := printTable(rows, 2, 20)
	if len(out[0]) > 22 { // indent 2 + width 20
		t.Fatalf("not truncated: %q (len %d)", out[0], len(out[0]))
	}
	if out[0][len(out[0])-3:] != "..." {
		t.Fatalf("no ellipsis: %q", out[0])
	}
}

func TestTruncate(t *testing.T) {
	if truncate("short", 20, 0) != "short" {
		t.Fatal("short should be unchanged")
	}
	got := truncate("abcdefghij", 8, 0)
	if got != "abcde..." {
		t.Fatalf("got %q", got)
	}
	// n < 0 clamps to 0.
	if got := truncate("abcdef", 5, 10); got != "..." {
		t.Fatalf("clamp got %q", got)
	}
}

func TestPrintWrapped(t *testing.T) {
	// Two paragraphs; the second wraps.
	msg := "first para\n\none two three four five six seven eight"
	out := printWrapped(msg, 2, 20)
	// A blank line separates the paragraphs.
	if !reflect.DeepEqual(out[0], "  first para") {
		t.Fatalf("out[0]=%q", out[0])
	}
	foundBlank := false
	for _, l := range out {
		if l == "" {
			foundBlank = true
		}
	}
	if !foundBlank {
		t.Fatalf("no paragraph separator: %v", out)
	}
}

func TestPrintWrappedSkipsEmptyParagraph(t *testing.T) {
	// A run of only whitespace between paragraphs yields no words -> skipped.
	out := printWrapped("only", 0, 80)
	if !reflect.DeepEqual(out, []string{"only"}) {
		t.Fatalf("out=%v", out)
	}
	if out2 := printWrapped("\n\n", 0, 80); out2 != nil {
		t.Fatalf("out2=%v", out2)
	}
}
