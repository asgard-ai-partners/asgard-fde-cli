package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
