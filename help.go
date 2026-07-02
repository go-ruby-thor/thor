// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"sort"
	"strings"
)

func (b *Base) basename() string {
	if b.config.Basename == "" {
		return "thor"
	}
	return b.config.Basename
}

func (b *Base) width() int {
	env := b.config.Getenv
	if env == nil {
		env = func(string) string { return "" }
	}
	return terminalWidth(b.config.TerminalWidth, env)
}

// banner renders "<basename> <formatted usage>" for a command, joining the
// per-usage lines (Thor::banner). subcommand toggles the subcommand prefix.
func (b *Base) banner(c *Command, subcommand bool) string {
	fu := c.FormattedUsage(b.Namespace, b.Arguments, false, subcommand)
	lines := strings.Split(fu, "\n")
	for i, l := range lines {
		lines[i] = b.basename() + " " + l
	}
	return strings.Join(lines, "\n")
}

// CommandHelp renders the per-command help block (Thor::command_help) for the
// named command, byte-faithful to the gem, or a [KindUndefinedCommand] error.
func (b *Base) CommandHelp(commandName string) (string, error) {
	name, err := b.NormalizeCommandName(commandName)
	if err != nil {
		return "", err
	}
	command := b.byName[name]
	if command == nil {
		return "", newError(KindUndefinedCommand,
			"Could not find command "+rubyInspect(commandName)+" in "+rubyInspect(b.Namespace)+" namespace.")
	}

	var out lineWriter
	out.say("Usage:")
	bannerLines := strings.Split(b.banner(command, false), "\n")
	out.say("  " + strings.Join(bannerLines, "\n  "))
	out.blank()
	// class_options_help(shell, nil => command.options.values): the default
	// group is seeded with this command's options, then class options are
	// appended by group.
	out.append(classOptionsHelpSeeded(command.Options, b.ClassOptions, b.width()))

	if command.LongDescription != "" {
		out.say("Description:")
		if command.WrapLongDescription {
			out.append(printWrapped(command.LongDescription, 2, b.width()))
		} else {
			out.say(command.LongDescription)
		}
	} else {
		out.say(command.Description)
	}
	return out.String(), nil
}

// Help renders the class-level help (Thor::help): the command listing followed
// by the class options table, byte-faithful to the gem.
func (b *Base) Help() string {
	list := b.printableCommands(true, false)
	sort.SliceStable(list, func(i, j int) bool { return list[i][0] < list[j][0] })

	var out lineWriter
	if b.config.PackageName != "" {
		out.say(b.config.PackageName + " commands:")
	} else {
		out.say("Commands:")
	}
	out.append(printTable(list, 2, b.width()))
	out.blank()
	out.append(classOptionsHelp(b.ClassOptions, b.width()))
	return out.String()
}

// printableCommands returns the [banner, "# desc"] rows for the command listing
// (Thor::printable_commands), skipping hidden commands.
func (b *Base) printableCommands(all, subcommand bool) [][]string {
	var list [][]string
	for _, c := range b.commands {
		if c.Hidden {
			continue
		}
		desc := ""
		if c.Description != "" {
			desc = "# " + collapseWhitespace(c.Description)
		}
		list = append(list, []string{b.banner(c, subcommand), desc})
	}
	return list
}

// collapseWhitespace mirrors gsub(/\s+/m, ' '): every maximal run of whitespace
// (including a leading or trailing run) becomes a single space.
func collapseWhitespace(s string) string {
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if isSpace(r) {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
			continue
		}
		inSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
}

// classOptionsHelp groups class options by group and renders each block, the
// default (empty) group first (Thor::Base#class_options_help).
func classOptionsHelp(options []*Option, width int) []string {
	var order []string
	groups := map[string][]*Option{}
	for _, o := range options {
		if _, ok := groups[o.Group]; !ok {
			order = append(order, o.Group)
		}
		groups[o.Group] = append(groups[o.Group], o)
	}
	var out []string
	out = append(out, printOptions(groups[""], "", width)...)
	for _, g := range order {
		if g == "" {
			continue
		}
		out = append(out, printOptions(groups[g], g, width)...)
	}
	return out
}

// classOptionsHelpSeeded renders the option help with the default (nil) group
// pre-seeded by a command's own options, then class options appended by group
// (Thor::Base#class_options_help called with {nil => command.options}).
func classOptionsHelpSeeded(seed, classOptions []*Option, width int) []string {
	var order []string
	groups := map[string][]*Option{}
	// The seed occupies the default group first.
	order = append(order, "")
	groups[""] = append(groups[""], seed...)
	for _, o := range classOptions {
		if _, ok := groups[o.Group]; !ok {
			order = append(order, o.Group)
		}
		groups[o.Group] = append(groups[o.Group], o)
	}
	var out []string
	out = append(out, printOptions(groups[""], "", width)...)
	for _, g := range order {
		if g == "" {
			continue
		}
		out = append(out, printOptions(groups[g], g, width)...)
	}
	return out
}

// printOptions renders one option table block (Thor::Base#print_options): a
// "<Group> options:" / "Options:" heading, the aligned table, and a blank line.
func printOptions(options []*Option, groupName string, width int) []string {
	if len(options) == 0 {
		return nil
	}

	padding := 0
	for _, o := range options {
		if l := len(o.aliasesForUsage()); l > padding {
			padding = l
		}
	}

	var rows [][]string
	for _, o := range options {
		if o.Hide {
			continue
		}
		desc := ""
		if o.Desc != "" {
			desc = "# " + o.Desc
		}
		rows = append(rows, []string{o.Usage(padding), desc})
		if o.ShowDefault() {
			rows = append(rows, []string{"", "# Default: " + o.PrintDefault()})
		}
		if len(o.Enum) > 0 {
			rows = append(rows, []string{"", "# Possible values: " + o.EnumToS()})
		}
	}
	if len(rows) == 0 {
		return nil
	}

	var out lineWriter
	if groupName != "" {
		out.say(groupName + " options:")
	} else {
		out.say("Options:")
	}
	out.append(printTable(rows, 2, 0))
	out.say("")
	return out.lines
}

// lineWriter accumulates output lines the way Thor::Shell#say does (each say
// emits one line; a bare say emits an empty line).
type lineWriter struct{ lines []string }

func (w *lineWriter) say(s string)       { w.lines = append(w.lines, s) }
func (w *lineWriter) blank()             { w.lines = append(w.lines, "") }
func (w *lineWriter) append(ls []string) { w.lines = append(w.lines, ls...) }
func (w *lineWriter) String() string     { return strings.Join(w.lines, "\n") + "\n" }
