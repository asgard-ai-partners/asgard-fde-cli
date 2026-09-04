package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
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

// printSources lists the documentation links a document cites, or every one in
// the corpus when no document is named.
//
// It exists because an engagement did it by hand. Building a customer deck, an
// agent opened each page it had used, read the Sources block at the foot, copied
// nine URLs out and checked each one itself - and every part of that except the
// checking is something the material already knows. A deck spans several pages,
// so this takes no argument as well as one.
func printSources(out io.Writer, docs []kb.Doc, read func(string) string, none string) {
	total := 0
	for _, d := range docs {
		if len(d.Sources) == 0 {
			continue
		}
		fmt.Fprintf(out, "%s\n", read(d.Name))
		for _, u := range d.Sources {
			fmt.Fprintf(out, "  %s\n", u)
			total++
		}
		fmt.Fprintln(out)
	}
	if total == 0 {
		fmt.Fprint(out, none)
		return
	}
	fmt.Fprintf(out, "%d link(s). **These are what the page was written from, not a\n"+
		"reading list for a customer** - a link that answers the question a customer\n"+
		"asked is worth handing over, and the rest is our own provenance.\n\n"+
		"`asgard-cli audit-material --urls` fetches every one of them and fails on a\n"+
		"404; six were dead the first time it ran, four of them pages marked\n"+
		"`draft: true`, which exist in a checkout and are not published.\n", total)
}
