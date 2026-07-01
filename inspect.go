// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"strconv"
	"strings"
)

// rubyInspect renders a string the way Ruby's String#inspect does for the
// characters Thor's messages exercise: wrap in double quotes and escape the
// common control characters and the quote/backslash.
func rubyInspect(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\v':
			b.WriteString(`\v`)
		case 0:
			b.WriteString(`\0`)
		case 0x1b:
			b.WriteString(`\e`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// valueInspect renders a parsed value with Ruby #inspect semantics, used in
// enum/malformatted error messages (a bare String is quoted; numbers plain).
func valueInspect(v any) string {
	switch x := v.(type) {
	case string:
		return rubyInspect(x)
	case bool:
		return strconv.FormatBool(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return formatFloat(x)
	case nil:
		return "nil"
	}
	return valueToDisplay(v)
}

// valueToDisplay renders a value the way Ruby String() / to_s would, used for
// "# Default:" lines where the raw value is printed without inspection quoting.
func valueToDisplay(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case int:
		return strconv.Itoa(x)
	case float64:
		return formatFloat(x)
	case nil:
		return ""
	}
	return ""
}

// formatFloat renders a float64 the way Ruby's Float#to_s does for the values
// Thor produces (always a decimal point, e.g. "3.0").
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}
