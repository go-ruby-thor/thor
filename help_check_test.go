// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"os"
	"testing"
)

// buildCheckCLI builds the CLI mirrored by scratchpad/oracle.rb.
func buildCheckCLI(t *testing.T) *Base {
	t.Helper()
	mustOpt := func(name string, o Option) *Option {
		op, err := NewOption(name, o)
		if err != nil {
			t.Fatal(err)
		}
		return op
	}
	b := NewBase("oracle.rb", Config{Basename: "oracle.rb", TerminalWidth: 80})
	verbose := mustOpt("verbose", Option{Typ: Boolean, Desc: "Be verbose", Aliases: []string{"-v"}})
	b.ClassOptions = []*Option{verbose}

	greet := NewCommand("greet", "Greet someone by NAME", "greet NAME", []*Option{
		mustOpt("greeting", Option{Typ: String, Default: "Hello", Desc: "The greeting to use", Aliases: []string{"-g"}}),
		mustOpt("loud", Option{Typ: Boolean, Desc: "Shout it"}),
		mustOpt("count", Option{Typ: Numeric, Desc: "Times", Enum: []string{"1", "2", "3"}}),
	})
	greet.LongDescription = "This command greets the person named NAME with an optional greeting."
	b.AddCommand(greet)
	b.AddCommand(NewCommand("list", "List things", "list", nil))
	return b
}

func TestCheckDumpHelp(t *testing.T) {
	if os.Getenv("THOR_DUMP") == "" {
		t.Skip("dump only")
	}
	b := buildCheckCLI(t)
	os.Stdout.WriteString("=====CLASSHELP=====\n")
	os.Stdout.WriteString(b.Help())
	os.Stdout.WriteString("=====CMDHELP greet=====\n")
	h, _ := b.CommandHelp("greet")
	os.Stdout.WriteString(h)
}
