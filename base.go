// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"sort"
	"strings"
)

// Config tunes dispatch and help rendering.
type Config struct {
	// Basename is the program name shown at the head of usage/banner lines
	// (Thor's basename, e.g. "myapp"). Defaults to "thor" when empty.
	Basename string
	// TerminalWidth overrides the layout width; 0 uses THOR_COLUMNS then 80.
	TerminalWidth int
	// PackageName, when set, heads the command listing with
	// "<PackageName> commands:" instead of "Commands:".
	PackageName string
	// Getenv resolves environment variables (for THOR_COLUMNS); nil means the
	// process environment is not consulted and only TerminalWidth/80 apply.
	Getenv func(string) string
}

// Base is a command registry: the Go analogue of a Thor subclass. It resolves
// an argv to a command, parses that command's options, and renders help — all
// deterministically, with command-body invocation left to the host.
type Base struct {
	// Namespace is the class namespace used in usage prefixes (e.g. "myapp").
	Namespace string
	// ClassOptions are options shared by every command (Thor `class_option`),
	// in declaration order.
	ClassOptions []*Option
	// Arguments are the class's declared positional arguments, in order.
	Arguments []*Option
	// DefaultCommand is invoked when argv names none; "" means "help".
	DefaultCommand string
	// Map holds `map` aliases (input token -> command name).
	Map map[string]string

	commands []*Command
	byName   map[string]*Command
	config   Config
}

// NewBase returns an empty registry with the given namespace and config.
func NewBase(namespace string, config Config) *Base {
	return &Base{
		Namespace: namespace,
		Map:       map[string]string{},
		byName:    map[string]*Command{},
		config:    config,
	}
}

// AddCommand registers a command (normalizing foo-bar to foo_bar as the key).
func (b *Base) AddCommand(c *Command) {
	key := strings.ReplaceAll(c.Name, "-", "_")
	c.Name = key
	if _, ok := b.byName[key]; !ok {
		b.commands = append(b.commands, c)
	}
	b.byName[key] = c
}

// Commands returns the registered commands in registration order.
func (b *Base) Commands() []*Command { return b.commands }

func (b *Base) defaultCommandName() string {
	if b.DefaultCommand == "" {
		return "help"
	}
	return b.DefaultCommand
}

// NormalizeCommandName resolves a user token to a command name the way
// Thor::normalize_command_name does: nil -> default; prefix completion; map
// aliases; ambiguity is an error. The final name uses underscores.
func (b *Base) NormalizeCommandName(meth string) (string, error) {
	if meth == "" {
		return strings.ReplaceAll(b.defaultCommandName(), "-", "_"), nil
	}
	possibilities := b.findCommandPossibilities(meth)
	if len(possibilities) > 1 {
		return "", newError(KindAmbiguousCommand,
			"Ambiguous command "+meth+" matches ["+strings.Join(possibilities, ", ")+"]")
	}
	switch {
	case len(possibilities) == 0:
		// meth unchanged (default already handled above).
	case b.Map[meth] != "":
		meth = b.Map[meth]
	default:
		meth = possibilities[0]
	}
	return strings.ReplaceAll(meth, "-", "_"), nil
}

// findCommandPossibilities returns the sorted prefix matches for meth over the
// visible command names and map keys (Thor::find_command_possibilities).
func (b *Base) findCommandPossibilities(meth string) []string {
	length := len(meth)
	set := map[string]bool{}
	for _, c := range b.commands {
		if c.Hidden {
			continue
		}
		set[c.Name] = true
	}
	for k := range b.Map {
		set[k] = true
	}
	var possibilities []string
	for n := range set {
		if length <= len(n) && meth == n[:length] {
			possibilities = append(possibilities, n)
		}
	}
	sort.Strings(possibilities)

	// unique_possibilities: map[k] || k, uniq (order-preserving).
	seen := map[string]bool{}
	var unique []string
	for _, k := range possibilities {
		v := k
		if m := b.Map[k]; m != "" {
			v = m
		}
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}

	if contains(possibilities, meth) {
		return []string{meth}
	}
	if len(unique) == 1 {
		return unique
	}
	return possibilities
}

// Dispatch resolves argv into a command, its parsed options, and the remaining
// positional arguments. It mirrors the deterministic half of Thor::dispatch:
// pick the command name, split argv at the first switch, and parse options.
// Invoking the resolved command body is the host seam.
func (b *Base) Dispatch(argv []string) (*Command, *Result, error) {
	given := append([]string{}, argv...)
	meth := b.retrieveCommandName(&given)

	name, err := b.NormalizeCommandName(meth)
	if err != nil {
		return nil, nil, err
	}
	command := b.byName[name]
	if command == nil {
		return nil, nil, newError(KindUndefinedCommand,
			"Could not find command "+rubyInspect(meth)+" in "+rubyInspect(b.Namespace)+" namespace.")
	}

	args, opts := Split(given)
	res, err := NewOptions(command.Options, nil, false, false, command.Relations).Parse(opts)
	if err != nil {
		return nil, nil, err
	}
	// Prepend the pre-switch args to the parsed remainder.
	res.Args = append(append([]string{}, args...), res.Args...)
	return command, res, nil
}

// retrieveCommandName pops and returns the command name from argv the way
// Thor::retrieve_command_name does: the first token is consumed when it is a
// map alias or does not start with "-".
func (b *Base) retrieveCommandName(args *[]string) string {
	if len(*args) == 0 {
		return ""
	}
	meth := (*args)[0]
	if b.Map[meth] != "" || !strings.HasPrefix(meth, "-") {
		*args = (*args)[1:]
		return meth
	}
	return ""
}

// Split partitions argv into leading positional arguments and the switch tail,
// the way Thor::Arguments.split does: stop at the first token starting with "-".
func Split(args []string) (arguments, rest []string) {
	i := 0
	for ; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
	}
	return args[:i], args[i:]
}
