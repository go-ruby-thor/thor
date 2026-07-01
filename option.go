// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"strings"
)

// Type is a Thor option/argument value type (Thor::Option::VALID_TYPES).
type Type string

// The Thor option value types.
const (
	String  Type = "string"
	Numeric Type = "numeric"
	Boolean Type = "boolean"
	Array   Type = "array"
	Hash    Type = "hash"
)

// Option models a single declared option (Thor::Option), i.e. a switch given
// via `option`/`method_option`/`method_options`. The zero value is not valid;
// build one with [NewOption].
type Option struct {
	// Name is the raw declared name (may use "_" or "-"). It drives SwitchName
	// and HumanName below.
	Name string
	// Desc is the one-line description shown in the option table.
	Desc string
	// Required marks the option as required (defaults to false for options).
	Required bool
	// Type is the value type; empty means String.
	Typ Type
	// Default is the default value (any of the mapped Go value types), or nil.
	Default any
	// LazyDefault is Thor's :lazy_default — the value used when the switch is
	// given without an inline value.
	LazyDefault any
	// Aliases are the short/long alias switches (e.g. "-f", "--flag"),
	// normalized to carry a leading dash.
	Aliases []string
	// Banner overrides the value placeholder shown in usage; empty uses the
	// per-type default banner.
	Banner string
	// BannerSet records whether Banner was explicitly provided (so an explicit
	// empty banner suppresses the placeholder).
	BannerSet bool
	// Enum, when non-empty, restricts accepted values.
	Enum []string
	// Group is the capitalized option group name ("" is the default group).
	Group string
	// Hide hides the option from help output.
	Hide bool
	// Repeatable accumulates repeated switches into an array (or merged hash).
	Repeatable bool
}

// NewOption returns an Option with the Thor defaults applied: Type defaults to
// String and, when unset, the banner defaults per type. It normalizes aliases
// to carry a leading dash. It returns an error for the invalid declarations
// Thor rejects (boolean+required).
func NewOption(name string, o Option) (*Option, error) {
	o.Name = name
	if o.Typ == "" {
		o.Typ = String
	}
	if o.Boolean() && o.Required {
		return nil, newError(KindArgument, "An option cannot be boolean and required.")
	}
	if !o.BannerSet {
		o.Banner = o.defaultBanner()
	}
	o.Aliases = normalizeAliases(o.Aliases)
	return &o, nil
}

func normalizeAliases(aliases []string) []string {
	out := make([]string, 0, len(aliases))
	for _, a := range aliases {
		if !strings.HasPrefix(a, "-") {
			a = "-" + a
		}
		out = append(out, a)
	}
	return out
}

// Boolean reports whether the option is of boolean type.
func (o *Option) Boolean() bool { return o.Typ == Boolean }

// StringType reports whether the option is of string type.
func (o *Option) StringType() bool { return o.Typ == String }

// dasherized reports whether Name already begins with a dash.
func (o *Option) dasherized() bool { return strings.HasPrefix(o.Name, "-") }

func undasherize(s string) string {
	return strings.TrimLeft(s, "-")
}

func dasherize(s string) string {
	prefix := "-"
	if len(s) > 1 {
		prefix = "--"
	}
	return prefix + strings.ReplaceAll(s, "_", "-")
}

// SwitchName is the canonical switch (e.g. "--flag") used as the map key.
func (o *Option) SwitchName() string {
	if o.dasherized() {
		return o.Name
	}
	return dasherize(o.Name)
}

// HumanName is the option's programmatic name (the key in the options hash).
func (o *Option) HumanName() string {
	if o.dasherized() {
		return undasherize(o.Name)
	}
	return o.Name
}

func (o *Option) defaultBanner() string {
	switch o.Typ {
	case Boolean:
		return ""
	case String:
		return strings.ToUpper(o.HumanName())
	case Numeric:
		return "N"
	case Hash:
		return "key:value"
	case Array:
		return "one two three"
	}
	return ""
}

// aliasesForUsage renders the "-a, -b, " prefix shown before a switch in the
// option table, or "" when there are no aliases.
func (o *Option) aliasesForUsage() string {
	if len(o.Aliases) == 0 {
		return ""
	}
	return strings.Join(o.Aliases, ", ") + ", "
}

// Usage renders the option's line in the help table, padded so the switch
// columns align (Thor::Option#usage). padding is the max aliases-prefix width.
func (o *Option) Usage(padding int) string {
	var sample string
	if o.BannerSet && o.Banner == "" {
		sample = o.SwitchName()
	} else if o.Banner != "" {
		sample = o.SwitchName() + "=" + o.Banner
	} else {
		sample = o.SwitchName()
	}
	if !o.Required {
		sample = "[" + sample + "]"
	}
	if o.Boolean() && o.Name != "force" && !hasNoSkipPrefix(o.Name) {
		sample += ", [" + dasherize("no-"+o.HumanName()) + "], [" + dasherize("skip-"+o.HumanName()) + "]"
	}
	return ljust(o.aliasesForUsage(), padding) + sample
}

func hasNoSkipPrefix(name string) bool {
	for _, p := range []string{"no-", "no_", "skip-", "skip_"} {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// ljust left-justifies s to at least width using spaces (Ruby String#ljust).
func ljust(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// EnumToS renders the enum for messages/help (Thor::Argument#enum_to_s).
func (o *Option) EnumToS() string {
	return strings.Join(o.Enum, ", ")
}

// ShowDefault reports whether a "# Default:" line is shown for this option.
func (o *Option) ShowDefault() bool {
	switch d := o.Default.(type) {
	case bool:
		return true
	case nil:
		return false
	case string:
		return d != ""
	case []string:
		return len(d) != 0
	case *OrderedMap:
		return d.Len() != 0
	default:
		return d != nil
	}
}

// PrintDefault renders the default value as it appears after "# Default:".
func (o *Option) PrintDefault() string {
	if o.Typ == Array {
		if arr, ok := o.Default.([]string); ok {
			parts := make([]string, len(arr))
			for i, s := range arr {
				parts[i] = rubyInspect(s)
			}
			return strings.Join(parts, " ")
		}
	}
	return valueToDisplay(o.Default)
}
