// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// runOracle invokes scratchpad/oracle.rb with the given mode and returns its
// stdout, or skips the test when ruby / the thor gem is unavailable (the CI
// no-ruby and Windows lanes). The deterministic golden tests keep coverage at
// 100% on those lanes; this test adds a differential check against MRI Thor.
func runOracle(t *testing.T, mode string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("no ruby oracle on windows")
	}
	if _, err := exec.LookPath("ruby"); err != nil {
		t.Skip("ruby not installed")
	}
	// Confirm the thor gem is present; otherwise skip rather than fail.
	if err := exec.Command("ruby", "-e", "require 'thor'").Run(); err != nil {
		t.Skip("thor gem not installed")
	}
	out, err := exec.Command("ruby", "scratchpad/oracle.rb", mode).Output()
	if err != nil {
		t.Fatalf("oracle.rb %s: %v", mode, err)
	}
	return string(out)
}

// section extracts the block that follows a "=====<marker>=====" line up to the
// next "=====" marker (or EOF).
func section(dump, marker string) string {
	head := "=====" + marker + "=====\n"
	i := strings.Index(dump, head)
	if i < 0 {
		return ""
	}
	rest := dump[i+len(head):]
	if j := strings.Index(rest, "====="); j >= 0 {
		return rest[:j]
	}
	return rest
}

// TestOracleCommandHelpDiff builds the same CLI as scratchpad/oracle.rb and
// checks our CommandHelp("greet") output is byte-identical to real Thor's.
func TestOracleCommandHelpDiff(t *testing.T) {
	dump := runOracle(t, "help")
	want := section(dump, "CMDHELP greet")
	if want == "" {
		t.Fatal("oracle produced no CMDHELP greet block")
	}
	b := buildCheckCLI(t)
	got, err := b.CommandHelp("greet")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("command help mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}
