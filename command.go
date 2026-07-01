// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"sort"
	"strings"
)

// Command models a registered command (Thor::Command): its name, description,
// usage line, and the options declared for it. Invoking the command body is the
// host seam; this type carries only the metadata used for dispatch and help.
type Command struct {
	// Name is the command's normalized method name (underscores).
	Name string
	// Description is the one-line `desc` text.
	Description string
	// LongDescription is the `long_desc` text ("" if none).
	LongDescription string
	// WrapLongDescription toggles word-wrapping of LongDescription (long_desc
	// wrap: option; default true).
	WrapLongDescription bool
	// Usage is the raw usage line (`desc "usage", ...`).
	Usage string
	// Options are the command's options in declaration order.
	Options []*Option
	// AncestorName, when set, prefixes the usage (subcommand ancestor).
	AncestorName string
	// Hidden hides the command from help listings.
	Hidden bool
	// Relations holds this command's exclusive / at-least-one option groups.
	Relations Relations
}

// NewCommand builds a Command with WrapLongDescription defaulting to true.
func NewCommand(name, description, usage string, options []*Option) *Command {
	return &Command{
		Name:                name,
		Description:         description,
		Usage:               usage,
		Options:             options,
		WrapLongDescription: true,
	}
}

// requiredOptions renders the required options appended to a usage line: each
// required option's Usage(0), sorted, space-joined (Thor::Command#required_options).
func (c *Command) requiredOptions() string {
	var parts []string
	for _, o := range c.Options {
		if o.Required {
			parts = append(parts, o.Usage(0))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

// FormattedUsage renders the command's usage line as Thor::Command#formatted_usage
// does: an ancestor / namespace / subcommand prefix, the usage with required
// class arguments injected, then required options. namespace is the class's
// namespace; args are the class's declared positional arguments (may be nil).
func (c *Command) FormattedUsage(namespace string, args []*Option, useNamespace, subcommand bool) string {
	var formatted string
	switch {
	case c.AncestorName != "":
		formatted = c.AncestorName + " "
	case useNamespace:
		ns := strings.Replace(namespace, "default", "", 1)
		if strings.HasPrefix(namespace, "default") {
			ns = namespace[len("default"):]
		}
		formatted = ns + ":"
	case subcommand:
		segs := strings.Split(namespace, ":")
		formatted = segs[len(segs)-1] + " "
	}

	usage := c.requiredArgumentsFor(args)
	usage += " " + c.requiredOptions()
	return strings.TrimSpace(formatted + usage)
}

// requiredArgumentsFor injects the class's positional argument usages right
// after the command name in the usage string (Thor::Command#required_arguments_for).
func (c *Command) requiredArgumentsFor(args []*Option) string {
	if len(args) == 0 {
		return c.Usage
	}
	usages := make([]string, 0, len(args))
	for _, a := range args {
		if u := argumentUsage(a); u != "" {
			usages = append(usages, u)
		}
	}
	inject := " " + strings.Join(usages, " ")
	// gsub(/^#{name}/) — replace a leading command name occurrence.
	if strings.HasPrefix(c.Usage, c.Name) {
		return c.Name + inject + c.Usage[len(c.Name):]
	}
	return c.Usage
}

// argumentUsage renders a positional argument's usage (Thor::Argument#usage):
// the banner, bracketed when optional.
func argumentUsage(a *Option) string {
	if a.Required {
		return a.Banner
	}
	return "[" + a.Banner + "]"
}
