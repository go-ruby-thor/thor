// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"fmt"
	"strconv"
	"strings"
)

// defaultTerminalWidth is Thor::Shell::Terminal::DEFAULT_TERMINAL_WIDTH.
const defaultTerminalWidth = 80

// terminalWidth resolves the layout width the way Thor::Shell::Terminal does:
// an explicit width wins; otherwise THOR_COLUMNS; otherwise the 80-column
// default. A width below 10 falls back to the default. width==0 means unset.
func terminalWidth(width int, env func(string) string) int {
	result := width
	if result == 0 {
		if c := env("THOR_COLUMNS"); c != "" {
			result = atoiRuby(c)
		} else {
			result = defaultTerminalWidth
		}
	}
	if result < 10 {
		return defaultTerminalWidth
	}
	return result
}

// atoiRuby mimics Ruby String#to_i: leading integer, else 0.
func atoiRuby(s string) int {
	s = strings.TrimSpace(s)
	i, sign := 0, 1
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if start == i {
		return 0
	}
	n, _ := strconv.Atoi(s[start:i])
	return n * sign
}

// printTable renders a Thor::Shell::TablePrinter table (no borders) to lines.
// indent is the left indent; truncateWidth>0 truncates each line to that width
// (Thor passes the terminal width when truncate: true, else 0 for no truncate).
func printTable(rows [][]string, indent, truncateWidth int) []string {
	if len(rows) == 0 {
		return nil
	}
	// colcount = size of the widest row.
	colcount := 0
	for _, r := range rows {
		if len(r) > colcount {
			colcount = len(r)
		}
	}
	// maxima per column and per-column format string.
	maximas := make([]int, colcount)
	formats := make([]string, colcount)
	for index := 0; index < colcount; index++ {
		maxima := 0
		for _, r := range rows {
			if index < len(r) {
				if l := len(r[index]); l > maxima {
					maxima = l
				}
			}
		}
		maximas[index] = maxima
		if index == colcount-1 {
			formats[index] = "%-s"
		} else {
			formats[index] = "%-" + strconv.Itoa(maxima+2) + "s"
		}
	}

	var out []string
	for _, row := range rows {
		var sentence strings.Builder
		for index := 0; index < len(row); index++ {
			sentence.WriteString(fmt.Sprintf(formats[index], row[index]))
		}
		line := sentence.String()
		if truncateWidth > 0 {
			line = truncate(line, truncateWidth, indent)
		}
		out = append(out, strings.Repeat(" ", indent)+line)
	}
	return out
}

// truncate mirrors TablePrinter#truncate: keep the string if it fits, else cut
// to (width-3-indent) chars and append "...".
func truncate(s string, width, indent int) string {
	chars := []rune(s)
	if len(chars) <= width {
		return s
	}
	n := width - 3 - indent
	if n < 0 {
		n = 0
	}
	return string(chars[:n]) + "..."
}

// printWrapped renders a Thor::Shell::WrappedPrinter block: word-wrap each
// blank-line-separated paragraph to (width-indent), indent every line, and put
// a blank line between paragraphs. width is the resolved terminal width.
func printWrapped(message string, indent, width int) []string {
	wrapWidth := width - indent
	paras := strings.Split(message, "\n\n")
	wrapped := make([]string, 0, len(paras))
	for _, unwrapped := range paras {
		words := strings.Fields(unwrapped)
		if len(words) == 0 {
			continue
		}
		memo := words[0]
		counter := len(words[0])
		for _, word := range words[1:] {
			if counter+len(word)+1 < wrapWidth {
				memo += " " + word
				counter += len(word) + 1
			} else {
				memo += "\n" + word
				counter = len(word)
			}
		}
		wrapped = append(wrapped, memo)
	}

	var out []string
	for pi, para := range wrapped {
		for _, line := range strings.Split(para, "\n") {
			out = append(out, strings.Repeat(" ", indent)+line)
		}
		if pi != len(wrapped)-1 {
			out = append(out, "")
		}
	}
	return out
}
