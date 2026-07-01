// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"regexp"
	"strconv"
	"strings"
)

// Regexes ported verbatim from Thor::Options / Thor::Arguments.
var (
	numericRe   = regexp.MustCompile(`[-+]?(\d*\.\d+|\d+)`)
	longRe      = regexp.MustCompile(`^(--\w+(?:-\w+)*)$`)
	shortRe     = regexp.MustCompile(`^(-[a-zA-Z])$`)
	eqRe        = regexp.MustCompile(`^(--\w+(?:-\w+)*|-[a-zA-Z])=(.*)$`)
	shortSqRe   = regexp.MustCompile(`^-([a-zA-Z]{2,})$`)
	shortNumRe = regexp.MustCompile(`^(-[a-zA-Z])([-+]?(?:\d*\.\d+|\d+))$`)
	noOrSkipRe = regexp.MustCompile(`^--(no|skip)-([-\w]+)$`)
)

// looksLikeUnknownSwitch implements Thor's /^--?(?:(?!--).)*$/ (RE2 has no
// negative lookahead): a token that starts with one or two dashes and contains
// no "--" run after that lead. Used by check_unknown!.
func looksLikeUnknownSwitch(s string) bool {
	if !strings.HasPrefix(s, "-") {
		return false
	}
	// Consume one or two leading dashes (the "--?" that the regex anchors to
	// the first char), then ensure the remainder never contains "--".
	rest := s[1:]
	if strings.HasPrefix(rest, "-") {
		rest = rest[1:]
	}
	return !strings.Contains(rest, "--")
}

const optsEnd = "--"

// Relations groups options for the exclusive / at-least-one checks. Each entry
// is a slice of option human names.
type Relations struct {
	Exclusive  [][]string
	AtLeastOne [][]string
}

// parser is the runtime state of a single parse, mirroring the instance
// variables of Thor::Options.
type parser struct {
	pile                       []string
	extra                      []string
	assigns                    *ValueMap
	nonAssignedRequired        []*Option
	switches                   map[string]*Option // switch name -> option
	shorts                     map[string]string  // alias -> switch name
	switchOrder                []*Option          // declaration order
	stopOnUnknown              bool
	disableRequiredCheck       bool
	parsingOptions             bool
	isTreatedAsValue           bool
	stoppedParsingAfterExtraIx int // -1 = not set (Ruby nil)
	exclusives                 [][]string
	atLeastOnes                [][]string
}

// Options declares a set of options to parse against. Build one with
// [NewOptions]; call [Options.Parse] to parse an argv.
type Options struct {
	options              []*Option
	defaults             *ValueMap
	stopOnUnknown        bool
	disableRequiredCheck bool
	relations            Relations
}

// NewOptions builds an Options parser from declared options and a defaults map
// (option human name -> value; Thor's `defaults` seeded from class/method
// options). Pass a zero Relations for no exclusive/at-least-one groups.
func NewOptions(options []*Option, defaults *ValueMap, stopOnUnknown, disableRequiredCheck bool, relations Relations) *Options {
	return &Options{
		options:              options,
		defaults:             defaults,
		stopOnUnknown:        stopOnUnknown,
		disableRequiredCheck: disableRequiredCheck,
		relations:            selectNonEmpty(relations),
	}
}

func selectNonEmpty(r Relations) Relations {
	out := Relations{}
	for _, g := range r.Exclusive {
		if len(g) > 0 {
			out.Exclusive = append(out.Exclusive, g)
		}
	}
	for _, g := range r.AtLeastOne {
		if len(g) > 0 {
			out.AtLeastOne = append(out.AtLeastOne, g)
		}
	}
	return out
}

// Parse parses argv against the declared options and returns the result
// (parsed options + remaining args) or a gem-faithful [*Error].
func (o *Options) Parse(argv []string) (*Result, error) {
	p := &parser{
		assigns:                    NewValueMap(),
		switches:                   map[string]*Option{},
		shorts:                     map[string]string{},
		stopOnUnknown:              o.stopOnUnknown,
		disableRequiredCheck:       o.disableRequiredCheck,
		stoppedParsingAfterExtraIx: -1,
		exclusives:                 o.relations.Exclusive,
		atLeastOnes:                o.relations.AtLeastOne,
	}

	// Seed defaults/required from Argument#initialize semantics.
	for _, opt := range o.options {
		p.switchOrder = append(p.switchOrder, opt)
		if opt.Default != nil {
			p.assigns.Set(opt.HumanName(), opt.Default)
		} else if opt.Required {
			p.nonAssignedRequired = append(p.nonAssignedRequired, opt)
		}
	}
	// Explicit defaults hash overrides and clears required.
	if o.defaults != nil {
		for _, k := range o.defaults.Keys() {
			v, _ := o.defaults.Get(k)
			p.assigns.Set(k, v)
			p.removeRequiredByHuman(k)
		}
	}
	// Register switches and aliases.
	for _, opt := range o.options {
		p.switches[opt.SwitchName()] = opt
		for _, a := range opt.Aliases {
			if _, ok := p.shorts[a]; !ok {
				p.shorts[a] = opt.SwitchName()
			}
		}
	}

	if err := p.run(argv); err != nil {
		return nil, err
	}
	return &Result{Options: p.assigns, Args: p.extra}, nil
}

func (p *parser) removeRequiredByHuman(human string) {
	for i, opt := range p.nonAssignedRequired {
		if opt.HumanName() == human {
			p.nonAssignedRequired = append(p.nonAssignedRequired[:i], p.nonAssignedRequired[i+1:]...)
			return
		}
	}
}

func (p *parser) removeRequired(opt *Option) {
	for i, o := range p.nonAssignedRequired {
		if o == opt {
			p.nonAssignedRequired = append(p.nonAssignedRequired[:i], p.nonAssignedRequired[i+1:]...)
			return
		}
	}
}

// --- pile primitives (Thor::Arguments) ---

func (p *parser) shift() string {
	p.isTreatedAsValue = false
	if len(p.pile) == 0 {
		return ""
	}
	s := p.pile[0]
	p.pile = p.pile[1:]
	return s
}

func (p *parser) unshift(arg string, isValue bool) {
	p.isTreatedAsValue = isValue
	p.pile = append([]string{arg}, p.pile...)
}

func (p *parser) unshiftMany(args []string) {
	p.isTreatedAsValue = false
	p.pile = append(append([]string{}, args...), p.pile...)
}

func (p *parser) last() bool { return len(p.pile) == 0 }

// peek returns the current token, honoring the "--" terminator like
// Thor::Options#peek: on hitting "--" while parsing options it shifts it,
// flips parsingOptions off, records the extra index, and returns the next.
func (p *parser) peek() (string, bool) {
	if !p.parsingOptions {
		if len(p.pile) == 0 {
			return "", false
		}
		return p.pile[0], true
	}
	if len(p.pile) == 0 {
		return "", false
	}
	result := p.pile[0]
	if result == optsEnd {
		p.shift()
		p.parsingOptions = false
		if p.stoppedParsingAfterExtraIx < 0 {
			p.stoppedParsingAfterExtraIx = len(p.extra)
		}
		if len(p.pile) == 0 {
			return "", false
		}
		return p.pile[0], true
	}
	return result, true
}

func (p *parser) parsingOptionsQ() bool {
	p.peek()
	return p.parsingOptions
}

// --- main loop (Thor::Options#parse) ---

func (p *parser) run(args []string) error {
	p.pile = append([]string{}, args...)
	p.isTreatedAsValue = false
	p.parsingOptions = true

	for {
		if _, ok := p.peek(); !ok {
			break
		}
		if p.parsingOptionsQ() {
			match, isSwitch := p.currentIsSwitch()
			shifted := p.shift()

			if isSwitch {
				var switchStr string
				switch {
				case shortSqRe.MatchString(shifted):
					m := shortSqRe.FindStringSubmatch(shifted)
					var expanded []string
					for _, f := range m[1] {
						expanded = append(expanded, "-"+string(f))
					}
					p.unshiftMany(expanded)
					continue
				case eqRe.MatchString(shifted):
					m := eqRe.FindStringSubmatch(shifted)
					p.unshift(m[2], true)
					switchStr = m[1]
				case shortNumRe.MatchString(shifted):
					m := shortNumRe.FindStringSubmatch(shifted)
					p.unshift(m[2], false)
					switchStr = m[1]
				case longRe.MatchString(shifted):
					switchStr = longRe.FindStringSubmatch(shifted)[1]
				case shortRe.MatchString(shifted):
					switchStr = shortRe.FindStringSubmatch(shifted)[1]
				}
				switchStr = p.normalizeSwitch(switchStr)
				option := p.switchOption(switchStr)
				result, err := p.parsePeek(switchStr, option)
				if err != nil {
					return err
				}
				p.assignResult(option, result)
			} else if p.stopOnUnknown {
				p.parsingOptions = false
				p.extra = append(p.extra, shifted)
				if p.stoppedParsingAfterExtraIx < 0 {
					p.stoppedParsingAfterExtraIx = len(p.extra)
				}
				for {
					if _, ok := p.peek(); !ok {
						break
					}
					p.extra = append(p.extra, p.shift())
				}
				break
			} else if match {
				p.extra = append(p.extra, shifted)
				for {
					pk, ok := p.peek()
					if !ok || strings.HasPrefix(pk, "-") {
						break
					}
					p.extra = append(p.extra, p.shift())
				}
			} else {
				p.extra = append(p.extra, shifted)
			}
		} else {
			p.extra = append(p.extra, p.shift())
		}
	}

	if !p.disableRequiredCheck {
		if err := p.checkRequirement(); err != nil {
			return err
		}
	}
	if err := p.checkExclusive(); err != nil {
		return err
	}
	if err := p.checkAtLeastOne(); err != nil {
		return err
	}
	return nil
}

// --- switch classification ---

func (p *parser) currentIsSwitch() (match bool, isSwitch bool) {
	if p.isTreatedAsValue {
		return false, false
	}
	pk, ok := p.peek()
	if !ok {
		return false, false
	}
	switch {
	case longRe.MatchString(pk):
		return true, p.switchQ(longRe.FindStringSubmatch(pk)[1])
	case shortRe.MatchString(pk):
		return true, p.switchQ(shortRe.FindStringSubmatch(pk)[1])
	case eqRe.MatchString(pk):
		return true, p.switchQ(eqRe.FindStringSubmatch(pk)[1])
	case shortNumRe.MatchString(pk):
		return true, p.switchQ(shortNumRe.FindStringSubmatch(pk)[1])
	case shortSqRe.MatchString(pk):
		m := shortSqRe.FindStringSubmatch(pk)
		for _, f := range m[1] {
			if p.switchQ("-" + string(f)) {
				return true, true
			}
		}
		return true, false
	}
	return false, false
}

func (p *parser) currentIsSwitchFormatted() bool {
	if p.isTreatedAsValue {
		return false
	}
	pk, ok := p.peek()
	if !ok {
		return false
	}
	return longRe.MatchString(pk) || shortRe.MatchString(pk) || eqRe.MatchString(pk) ||
		shortNumRe.MatchString(pk) || shortSqRe.MatchString(pk)
}

func (p *parser) currentIsValue() bool {
	if p.isTreatedAsValue {
		return true
	}
	pk, ok := p.peek()
	if !ok {
		return false
	}
	if !p.parsingOptionsQ() {
		return true
	}
	// Arguments#current_is_value?: peek && peek !~ /^-{1,2}\S+/
	return !valueLooksLikeSwitch(pk)
}

// valueLooksLikeSwitch implements /^-{1,2}\S+/ (a dash-led non-space run).
func valueLooksLikeSwitch(s string) bool {
	if !strings.HasPrefix(s, "-") {
		return false
	}
	rest := strings.TrimLeft(s, "-")
	dashes := len(s) - len(strings.TrimPrefix(s, "-"))
	if dashes < 1 || dashes > 2 {
		// more than two leading dashes: /^-{1,2}\S+/ still matches the first
		// two dashes then \S+ so treat as switch-looking.
		return len(rest) > 0 || dashes >= 1
	}
	return len(rest) > 0
}

func (p *parser) switchQ(arg string) bool {
	return p.switchOption(p.normalizeSwitch(arg)) != nil
}

func (p *parser) switchOption(arg string) *Option {
	if m := noOrSkip(arg); m != "" {
		if opt, ok := p.switches[arg]; ok {
			return opt
		}
		return p.switches["--"+m]
	}
	return p.switches[arg]
}

func (p *parser) normalizeSwitch(arg string) string {
	if s, ok := p.shorts[arg]; ok {
		arg = s
	}
	return strings.ReplaceAll(arg, "_", "-")
}

// noOrSkip returns the name captured by /^--(no|skip)-([-\w]+)$/, or "".
func noOrSkip(arg string) string {
	m := noOrSkipRe.FindStringSubmatch(arg)
	if m == nil {
		return ""
	}
	return m[2]
}

// --- value assignment ---

func (p *parser) assignResult(option *Option, result any) {
	if option == nil {
		return
	}
	if option.Repeatable && option.Typ == Hash {
		cur, ok := p.assigns.Get(option.HumanName())
		var m *OrderedMap
		if ok {
			m, _ = cur.(*OrderedMap)
		}
		if m == nil {
			m = NewOrderedMap()
		}
		if rm, ok := result.(*OrderedMap); ok {
			m.Merge(rm)
		}
		p.assigns.Set(option.HumanName(), m)
	} else if option.Repeatable {
		cur, ok := p.assigns.Get(option.HumanName())
		var arr []any
		if ok {
			arr, _ = cur.([]any)
		}
		arr = append(arr, result)
		p.assigns.Set(option.HumanName(), arr)
	} else {
		p.assigns.Set(option.HumanName(), result)
	}
}

// parsePeek mirrors Thor::Options#parse_peek: decide whether the switch takes a
// value here, applying boolean / no-value / lazy-default / default rules.
func (p *parser) parsePeek(switchStr string, option *Option) (any, error) {
	if p.parsingOptionsQ() && (p.currentIsSwitchFormatted() || p.last()) {
		switch {
		case option.Boolean():
			// falls through to parse below
		case noOrSkip(switchStr) != "":
			return nil, nil
		case option.StringType() && !option.Required:
			if option.LazyDefault != nil {
				return option.LazyDefault, nil
			}
			if option.Default != nil {
				return option.Default, nil
			}
			return option.HumanName(), nil
		case option.LazyDefault != nil:
			return option.LazyDefault, nil
		default:
			return nil, newError(KindMalformattedArgument,
				"No value provided for option '"+switchStr+"'")
		}
	}
	p.removeRequired(option)
	return p.parseByType(option.Typ, switchStr, option)
}

func (p *parser) parseByType(t Type, name string, option *Option) (any, error) {
	switch t {
	case Boolean:
		return p.parseBoolean(name), nil
	case String:
		return p.parseString(name, option)
	case Numeric:
		return p.parseNumeric(name, option)
	case Array:
		return p.parseArray(name, option)
	case Hash:
		return p.parseHash(name)
	}
	return nil, nil
}

func (p *parser) parseBoolean(switchStr string) bool {
	if p.currentIsValue() {
		pk, _ := p.peek()
		switch pk {
		case "true", "TRUE", "t", "T":
			p.shift()
			return true
		case "false", "FALSE", "f", "F":
			p.shift()
			return false
		default:
			if _, ok := p.switches[switchStr]; ok {
				return true
			}
			return noOrSkip(switchStr) == ""
		}
	}
	if _, ok := p.switches[switchStr]; ok {
		return true
	}
	return noOrSkip(switchStr) == ""
}

func (p *parser) parseString(name string, option *Option) (any, error) {
	if noOrSkip(name) != "" {
		return nil, nil
	}
	value := p.shift()
	if err := validateEnum(option, name, value, "Expected '%s' to be one of %s; got %s"); err != nil {
		return nil, err
	}
	return value, nil
}

func (p *parser) parseNumeric(name string, option *Option) (any, error) {
	pk, _ := p.peek()
	loc := numericRe.FindString(pk)
	if loc != pk || loc == "" {
		return nil, newError(KindMalformattedArgument,
			"Expected numeric value for '"+name+"'; got "+rubyInspect(pk))
	}
	var value any
	if strings.Contains(loc, ".") {
		f, _ := strconv.ParseFloat(p.shift(), 64)
		value = f
	} else {
		n, _ := strconv.ParseInt(p.shift(), 10, 64)
		value = n
	}
	if err := validateEnumValue(option, name, value, "Expected '%s' to be one of %s; got %s"); err != nil {
		return nil, err
	}
	return value, nil
}

func (p *parser) parseArray(name string, option *Option) (any, error) {
	var arr []string
	for p.currentIsValue() {
		value := p.shift()
		if value != "" {
			if err := validateEnum(option, name, value,
				"Expected all values of '%s' to be one of %s; got %s"); err != nil {
				return nil, err
			}
		}
		arr = append(arr, value)
	}
	if arr == nil {
		arr = []string{}
	}
	return arr, nil
}

func (p *parser) parseHash(name string) (any, error) {
	h := NewOrderedMap()
	for p.currentIsValue() {
		pk, _ := p.peek()
		if !strings.Contains(pk, ":") {
			break
		}
		tok := p.shift()
		parts := strings.SplitN(tok, ":", 2)
		key, value := parts[0], parts[1]
		if h.Has(key) {
			prev, _ := h.Get(key)
			return nil, newError(KindMalformattedArgument,
				"You can't specify '"+key+"' more than once in option '"+name+"'; got "+
					key+":"+prev+" and "+key+":"+value)
		}
		h.Set(key, value)
	}
	return h, nil
}

func validateEnum(option *Option, name, value string, message string) error {
	if option == nil || len(option.Enum) == 0 {
		return nil
	}
	for _, e := range option.Enum {
		if e == value {
			return nil
		}
	}
	return newError(KindMalformattedArgument,
		sprintfMsg(message, name, option.EnumToS(), value))
}

func validateEnumValue(option *Option, name string, value any, message string) error {
	if option == nil || len(option.Enum) == 0 {
		return nil
	}
	got := valueToDisplay(value)
	for _, e := range option.Enum {
		if e == got {
			return nil
		}
	}
	return newError(KindMalformattedArgument,
		sprintfMsg(message, name, option.EnumToS(), got))
}

// sprintfMsg fills a Ruby `msg % [a, b, c]` template with exactly three %s.
func sprintfMsg(tmpl, a, b, c string) string {
	out := tmpl
	out = replaceFirst(out, "%s", a)
	out = replaceFirst(out, "%s", b)
	out = replaceFirst(out, "%s", c)
	return out
}

func replaceFirst(s, old, new string) string {
	i := strings.Index(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

// --- requirement checks ---

func (p *parser) checkRequirement() error {
	if len(p.nonAssignedRequired) == 0 {
		return nil
	}
	names := make([]string, len(p.nonAssignedRequired))
	for i, o := range p.nonAssignedRequired {
		names[i] = o.SwitchName()
	}
	return newError(KindRequiredArgumentMissing,
		"No value provided for required options '"+strings.Join(names, "', '")+"'")
}

func (p *parser) checkExclusive() error {
	opts := p.assigns.Keys()
	for _, ex := range p.exclusives {
		if len(subtract(ex, opts)) < len(ex)-1 {
			names := p.namesToSwitchNames(intersect(ex, opts))
			quoted := quoteAll(names)
			return newError(KindExclusiveArgument,
				"Found exclusive option "+strings.Join(quoted, ", "))
		}
	}
	return nil
}

func (p *parser) checkAtLeastOne() error {
	opts := p.assigns.Keys()
	for _, group := range p.atLeastOnes {
		none := true
		for _, o := range group {
			if contains(opts, o) {
				none = false
				break
			}
		}
		if none {
			names := p.namesToSwitchNames(group)
			quoted := quoteAll(names)
			return newError(KindAtLeastOneRequiredArgument,
				"Not found at least one of required option "+strings.Join(quoted, ", "))
		}
	}
	return nil
}

// namesToSwitchNames maps option human names to switch names, in switch
// declaration order (Thor iterates @switches).
func (p *parser) namesToSwitchNames(names []string) []string {
	var out []string
	for _, opt := range p.switchOrder {
		if contains(names, opt.HumanName()) {
			out = append(out, opt.SwitchName())
		}
	}
	return out
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = "'" + s + "'"
	}
	return out
}

func subtract(a, b []string) []string {
	var out []string
	for _, x := range a {
		if !contains(b, x) {
			out = append(out, x)
		}
	}
	return out
}

func intersect(a, b []string) []string {
	var out []string
	for _, x := range a {
		if contains(b, x) {
			out = append(out, x)
		}
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
