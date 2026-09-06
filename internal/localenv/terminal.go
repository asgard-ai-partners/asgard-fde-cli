package localenv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Terminal fills the file in from a terminal, for a machine with no browser and
// no way to reach one - a container, a locked-down server, an SSH session
// where forwarding a port is not on offer.
//
// **It does not echo a secret**, which is the only reason this needs a
// dependency at all: turning off terminal echo means talking to the tty, and
// improvising that is not something to do with a customer's password.
func Terminal(opts Options, in *os.File, out io.Writer) (Result, error) {
	if !term.IsTerminal(int(in.Fd())) {
		return Result{}, fmt.Errorf(
			"--terminal needs a terminal to type into, and this is not one\n" +
				"  a script filling this in should write the .env directly")
	}

	before := snapshot(opts.File)
	focus := map[string]bool{}
	for _, k := range opts.Focus {
		focus[k] = true
	}

	entries := opts.File.Entries()
	if len(entries) == 0 {
		return Result{}, fmt.Errorf("%s has no keys yet, and --terminal cannot add them\n"+
			"  the agent writes the keys it needs first; then this fills them in", opts.File.Path)
	}

	fmt.Fprintf(out, "%s - %d keys\n\n", opts.File.Path, len(entries))
	fmt.Fprintln(out, "Enter keeps what is there. A single - empties a key.")
	fmt.Fprintln(out, "Secrets are not echoed.")
	fmt.Fprintln(out)

	reader := bufio.NewReader(in)
	for _, e := range entries {
		label := e.Key
		switch {
		case focus[e.Key]:
			label += "  [asked for]"
		case e.Value == "":
			label += "  [to fill in]"
		}
		fmt.Fprintf(out, "%s\n", label)
		if e.Note != "" {
			fmt.Fprintf(out, "  %s\n", e.Note)
		}
		if e.Value != "" {
			if e.Secret {
				fmt.Fprintf(out, "  currently set (%d characters)\n", len(e.Value))
			} else {
				fmt.Fprintf(out, "  currently %s\n", e.Value)
			}
		}

		var typed string
		var err error
		if e.Secret {
			fmt.Fprint(out, "  > ")
			var b []byte
			b, err = term.ReadPassword(int(in.Fd()))
			fmt.Fprintln(out)
			typed = string(b)
		} else {
			fmt.Fprint(out, "  > ")
			typed, err = reader.ReadString('\n')
		}
		if err != nil && err != io.EOF {
			return Result{}, fmt.Errorf("read %s: %w", e.Key, err)
		}
		typed = strings.TrimRight(typed, "\r\n")

		switch typed {
		case "":
			// Left alone.
		case "-":
			opts.File.Set(e.Key, "")
		default:
			opts.File.Set(e.Key, typed)
		}
		fmt.Fprintln(out)
		if err == io.EOF {
			break
		}
	}

	if err := opts.File.Save(); err != nil {
		return Result{}, err
	}
	return diff(before, opts.File), nil
}
