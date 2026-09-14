// Command hack is this repository's maintainer gate.
//
// **One binary with a subcommand each, rather than a directory of scripts.**
// These checks run mostly by hand, so an error in a branch nobody takes often
// survives for weeks; compiled, it does not survive `go build`. AGENTS.md has
// the argument and the five that got through.
//
//	go run ./hack <check> [flags]
//	go run ./hack list          every check, and what each one needs
//
// The checks that need no clone run in CI. The rest need somebody else's
// repository - see hack/internal/src for how they are found, and note that
// nothing here fetches: a script that pulled would turn "read at this commit"
// into "read at whatever was there when it ran".
package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

// errFailed means the check failed and has already said why.
var errFailed = errors.New("check failed")

// check is one entry in the gate.
type check struct {
	// Needs is what it cannot run without, and is the only part of the
	// grouping that is judgement rather than discovery.
	Needs string
	// What it answers, in one line, for `list`.
	What string
	Run  func(args []string) error
}

var checks = map[string]check{}

func register(name string, c check) { checks[name] = c }

func main() {
	if len(os.Args) < 2 || os.Args[1] == "list" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		list()
		return
	}
	c, ok := checks[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "no check called %q\n\n", os.Args[1])
		list()
		os.Exit(2)
	}
	if err := c.Run(os.Args[2:]); err != nil {
		// **A check that has already printed its findings says nothing more.**
		// It returns errFailed so the exit code carries the answer, because a
		// summary line after the findings is a second account of one fact - the
		// same reason no check here records a verdict in prose.
		if err != errFailed {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func list() {
	names := make([]string, 0, len(checks))
	width := 0
	for n := range checks {
		names = append(names, n)
		if len(n) > width {
			width = len(n)
		}
	}
	sort.Strings(names)
	fmt.Println("go run ./hack <check>")
	fmt.Println()
	for _, n := range names {
		fmt.Printf("  %-*s  %s\n", width, n, checks[n].What)
		if checks[n].Needs != "" {
			fmt.Printf("  %-*s  needs %s\n", width, "", checks[n].Needs)
		}
	}
}
