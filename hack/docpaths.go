package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
)

func init() {
	register("doc-paths", check{
		Needs: "this repository",
		What:  "every path and package-qualified Go symbol this repository's own documents name",
		Run:   runDocPaths,
	})
}

// The root documents, plus this directory's own - `hack/README.md` and the
// scripts, which tell somebody what to run and had the last stale reference.
//
// **This file is not in the list.** It names the deleted symbols it was written
// to catch, in the comments explaining why it resolves a symbol inside its own
// package, so checking itself would report its own subject. Same reason
// `internal/cli/self.go` parses Go source instead of grepping it: a note
// recording that something was removed must not read as naming it.
var docs = []string{
	"Goal.md", "README.md", "README.zh-TW.md", "AGENTS.md", "STRUCTURE.md",
	"APPROACH.md", "TASK.md", "CLAUDE.md",
	"hack/README.md", "hack/verify-references.sh",
	".env.example", ".agents/skills/consistency-checks/SKILL.md",
}

// A path inside this repository: a directory we own, then a file or directory
// under it.
//
// **Backticks are not required.** The gate section writes its commands inside a
// fenced block, where a path carries none, so the match is on the shape of the
// path - anchored to a directory this repository owns - wherever it appears.
var pathRe = regexp.MustCompile(
	`(?:^|[^A-Za-z0-9_./` + "`" + `-])((?:source|hack|internal|cmd|prompts|projects|docs|assets|plugins|\.github)/[A-Za-z0-9_./*<>-]*)`)

// A Go symbol written `pkg.Symbol`. The documents cite these as the place a
// rule is implemented, and a deleted one sends a reader looking for it.
var symbolRe = regexp.MustCompile("`([a-z][a-z0-9]*)\\.([A-Z][A-Za-z0-9_]*)`")

var declRe = "func|type|var|const"

// skip drops placeholders the scaffold expands at write time, globs, and paths
// inside a customer's repository rather than ours. None is a path on disk here
// and all are correct in prose.
func skipPath(p string) bool {
	return strings.ContainsAny(p, "<*") || strings.Contains(p, "__") ||
		strings.HasPrefix(p, "projects/") || strings.HasPrefix(p, "docs/") ||
		strings.HasPrefix(p, "assets/")
}

func goFiles(root string) []string {
	var out []string
	_ = filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".go") {
			return nil
		}
		if strings.Contains(p, string(filepath.Separator)+".out"+string(filepath.Separator)) {
			return nil
		}
		out = append(out, p)
		return nil
	})
	return out
}

// goPackages returns the package names a `pkg.Symbol` reference could
// legitimately name: every directory under internal/, plus the last segment of
// every import path in the tree, plus the module root's own package.
//
// A reference to a package outside that set names nothing - which is how
// `config.Find` survived the deletion of the whole `config` package.
func goPackages(root string) map[string]bool {
	names := map[string]bool{"selfsrc": true}
	entries, _ := os.ReadDir(filepath.Join(root, "internal"))
	for _, e := range entries {
		if e.IsDir() {
			names[e.Name()] = true
		}
	}
	for _, f := range goFiles(root) {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(strings.Trim(strings.TrimSpace(line), "_ "))
			if strings.HasPrefix(line, `"`) && strings.HasSuffix(line, `"`) {
				p := strings.Trim(line, `"`)
				names[p[strings.LastIndex(p, "/")+1:]] = true
			}
		}
	}
	return names
}

func runDocPaths(_ []string) error {
	root, err := src.Root()
	if err != nil {
		return err
	}
	pkgs := goPackages(root)
	internal := map[string]string{"selfsrc": root}
	entries, _ := os.ReadDir(filepath.Join(root, "internal"))
	for _, e := range entries {
		if e.IsDir() {
			internal[e.Name()] = filepath.Join(root, "internal", e.Name())
		}
	}

	bad, checked := 0, 0
	for _, name := range docs {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			fmt.Printf("missing document  %s\n", name)
			bad++
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			for _, m := range pathRe.FindAllStringSubmatch(line, -1) {
				p := m[1]
				if skipPath(p) {
					continue
				}
				checked++
				if _, err := os.Stat(filepath.Join(root, strings.TrimRight(p, "/"))); err != nil {
					fmt.Printf("gone  %s:%d -> `%s`\n", name, i+1, p)
					bad++
				}
			}
			for _, m := range symbolRe.FindAllStringSubmatch(line, -1) {
				pkg, sym := m[1], m[2]
				checked++
				if !pkgs[pkg] {
					fmt.Printf("gone  %s:%d -> `%s.%s` (no such package)\n", name, i+1, pkg, sym)
					bad++
					continue
				}
				dir, ours := internal[pkg]
				if !ours {
					// Somebody else's package. Whether it has that symbol is
					// their business and their version's.
					continue
				}
				// **Resolved inside the package that owns it.** A search of the
				// whole tree cannot tell `usecase.Index` from `strings.Index`,
				// and passed the first for months after it was deleted.
				if !declares(dir, sym) {
					fmt.Printf("gone  %s:%d -> `%s.%s`\n", name, i+1, pkg, sym)
					bad++
				}
			}
		}
	}
	fmt.Printf("\n%d path(s) and symbol(s) named, %d that are not there.\n", checked, bad)
	if bad > 0 {
		fmt.Println("\nA document that names a file or a symbol this repository does not have")
		fmt.Println("sends a reader to look for it. Fix it, or delete the sentence.")
		return errFailed
	}
	return nil
}

func declares(dir, sym string) bool {
	paths, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	var b strings.Builder
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			b.Write(data)
			b.WriteString("\n")
		}
	}
	own := b.String()
	decl := regexp.MustCompile(`\b(?:` + declRe + `)\s+(?:\([^)]*\)\s*)?` + regexp.QuoteMeta(sym) + `\b`)
	field := regexp.MustCompile(`(?m)^\t` + regexp.QuoteMeta(sym) + `\s`)
	return decl.MatchString(own) || field.MatchString(own)
}
