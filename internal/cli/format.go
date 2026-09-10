package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrSilent fails a command without a message.
//
// A `--format json` run has already emitted its whole answer, and the answer
// says whether it passed. Printing "Error: 3 problem(s)" after it puts a line
// on stderr that no caller parsing the JSON can use, and a caller reading both
// has two sources for one fact - which is how they come to disagree.
var ErrSilent = errors.New("")

// formatFlag is the name of the output-format flag, and formatJSON its one
// non-default value.
//
// Column-aligned output is for the FDE and is not an interface: an agent acting
// on open questions should not have to recover them from a `%-9s`. The commands
// an agent acts on carry this; the ones a person reads for prose do not, and
// `wiki`, `usecase` and `brief` are already whole documents rather than
// records.
const (
	formatFlag  = "format"
	formatText  = "text"
	formatJSON  = "json"
	formatUsage = "output format: text for a reader, json for an agent"
)

// checkFormat rejects an unknown format rather than silently printing text,
// because a caller that asked for json and got prose will parse the prose.
func checkFormat(format string) error {
	switch format {
	case formatText, formatJSON:
		return nil
	}
	return fmt.Errorf("unknown --%s %q; one of %s, %s", formatFlag, format, formatText, formatJSON)
}

// writeJSON emits v indented, so a person can also read what an agent gets.
func writeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// wrapAt and truncate came from internal/cli/usecase.go, which was deleted with
// the reader commands. They were never that command's: three other files
// already used them, and text-shaping does not belong in a command file.

// wrapAt breaks a provenance line so it stays readable in a terminal, indenting
// continuations to line up under the first.
func wrapAt(s string, width, indent int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	line := 0
	for i, w := range words {
		if i > 0 && line+1+len([]rune(w)) > width {
			b.WriteString("\n" + strings.Repeat(" ", indent))
			line = 0
		} else if i > 0 {
			b.WriteString(" ")
			line++
		}
		b.WriteString(w)
		line += len([]rune(w))
	}
	return b.String()
}

// truncate cuts to n runes, not n bytes. Slicing a string by byte splits a
// multi-byte character in half and prints a replacement glyph, which the wiki
// pages hit on every line because they are written in Chinese.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "..."
}
