package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strings"

	selfsrc "github.com/asgard-ai-partners/asgard-fde-cli"
)

// goStrings returns every string literal in this repository's Go source, keyed
// by file.
//
// **A printed string makes the same claim a document does**: that this build
// answers to what it names. The material and the scaffold templates are
// checked, and the tool's own output was not - so a renamed command left dead
// names in the help, in `check`'s findings and in the generator's warnings,
// and the only thing that caught one was the check written into a customer's
// repository, an engagement later. The source is embedded at the module root;
// see that package for why.
//
// **Parsed rather than grepped, so that comments are excluded.** A comment
// recording that a command *was* removed must not read as naming it - the
// package doc above says `asgard-cli add` and must not become a claim that
// `add` prints it - and a regular expression over the source cannot tell a
// comment from a message.
//
// **What it reads inside a string is a code segment**, the same as in the
// material: a backticked command or an indented block. `asgard-cli x y` in
// bare prose is not checked, and widening to that produces false positives -
// "a newer asgard-cli than the one you are running" resolves `than` as a
// subcommand. The convention is the check: write a command the way the
// material does.
func goStrings() (map[string]string, error) {
	out := map[string]string{}
	entries, err := fs.Glob(selfsrc.Go, "internal/*/*.go")
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	for _, name := range entries {
		src, err := selfsrc.Go.ReadFile(name)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, src, 0)
		if err != nil {
			// A file this build cannot parse is this build's problem, not the
			// material's, and failing the audit for it would be the worst of
			// both.
			continue
		}
		var b strings.Builder
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			b.WriteString(literalText(lit.Value))
			b.WriteString("\n")
			return true
		})
		if b.Len() > 0 {
			out[name] = b.String()
		}
	}
	return out, nil
}

// literalText unquotes a Go string literal well enough to read commands out of
// it. `strconv.Unquote` refuses a raw string containing a backquote, which is
// most of the help text here, so this does the two things that matter: drop
// the delimiters, and turn an escaped newline into a real one so a command
// split across concatenated lines is still one line.
func literalText(v string) string {
	if len(v) >= 2 {
		v = v[1 : len(v)-1]
	}
	return strings.ReplaceAll(v, `\n`, "\n")
}
