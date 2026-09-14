package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// immutableFields returns every property carrying `self == oldSelf`, by kind.
//
// **The rule count and the property count are not the same number**, because
// one kind can carry the rule at two paths - which is why `wiki/crd-rules.md`
// states 41 rules and 40 properties and says so.
func immutableFields(crds []crd) map[string][]string {
	out := map[string][]string{}
	for _, c := range crds {
		seen := map[string]bool{}
		walkProps(c.spec(), "", func(path string, schema map[string]any) {
			rules, ok := schema["x-kubernetes-validations"].([]any)
			if !ok {
				return
			}
			for _, r := range rules {
				m, ok := r.(map[string]any)
				if !ok {
					continue
				}
				if strings.TrimSpace(fmt.Sprint(m["rule"])) == "self == oldSelf" {
					seen[path] = true
				}
			}
		})
		if len(seen) > 0 {
			out[c.Spec.Names.Kind] = sortedKeys(seen)
		}
	}
	return out
}

// checkImmutable holds `wiki/crd-rules.md`'s account of what is chosen once
// against the CRDs.
//
// Nothing offline can say a chart will be refused at apply - that is the
// apiserver's, evaluated on write - but **which fields carry the rule decides a
// plan before anything is applied**, and a field nobody wrote down is one an
// FDE meets after the tag is pushed.
func checkImmutable(root string, crds []crd) []string {
	page, err := os.ReadFile(filepath.Join(root, "internal/corpus/wiki/crd-rules.md"))
	if err != nil {
		return []string{err.Error()}
	}
	text := string(page)
	byKind := immutableFields(crds)

	// **Pairs, not distinct paths.** `bot.botProviderName` is immutable on the
	// Loader and on the Syncer, and those are two fields somebody can be
	// refused on - counting the path once says 33 where the answer is 40.
	pairs, classes := 0, map[string]bool{}
	for _, fields := range byKind {
		pairs += len(fields)
		for _, f := range fields {
			if strings.HasSuffix(f, "Class") {
				classes[f] = true
			}
		}
	}

	var out []string
	for _, want := range []struct {
		re    *regexp.Regexp
		value int
		what  string
	}{
		{regexp.MustCompile(`(\d+) properties across\s*\n?\s*twelve kinds`), pairs, "immutable properties"},
		{regexp.MustCompile(`(?i)the Syncer is where this costs the most: (\d+) of the \d+`), len(byKind["Syncer"]), "Syncer immutable fields"},
		{regexp.MustCompile(`costs the most: \d+ of the (\d+)`), pairs, "immutable properties"},
	} {
		m := want.re.FindStringSubmatch(text)
		if m == nil {
			out = append(out, fmt.Sprintf(
				"internal/corpus/wiki/crd-rules.md states no %s the way this check reads it, so %d is going unchecked",
				want.what, want.value))
			continue
		}
		if atoi(m[1]) != want.value {
			out = append(out, fmt.Sprintf("internal/corpus/wiki/crd-rules.md says %s %s, and the CRDs have %d",
				m[1], want.what, want.value))
		}
	}

	var unnamed []string
	for _, c := range sortedKeys(classes) {
		if !strings.Contains(text, "`"+c+"`") {
			unnamed = append(unnamed, c)
		}
	}
	if len(unnamed) > 0 {
		out = append(out, "crd-rules.md does not name every immutable class field: missing "+strings.Join(unnamed, ", "))
	}
	if len(byKind) != 12 {
		out = append(out, fmt.Sprintf("%d kinds carry an immutable field and the page says twelve", len(byKind)))
	}

	// The Syncer's list is written out in full, so every one of its fields has
	// to be somewhere in the corpus - that is what an FDE reads before
	// believing a Syncer can be edited.
	corpus, err := shipped(root)
	if err != nil {
		return append(out, err.Error())
	}
	for _, f := range byKind["Syncer"] {
		if !named(corpus, f) && !named(corpus, leaf(f)) {
			out = append(out, fmt.Sprintf("Syncer.%s is immutable and no page in the corpus names it", f))
		}
	}
	return out
}

// checkRequiredBlocks reports a required field of a per-class block that no
// page names.
//
// **These are the fields an FDE asks a customer for**, and a block whose second
// field nobody wrote down sends somebody to a meeting with half the ask:
// `BotProvider.spec.telegram` requires `webhookSecretToken` beside `botToken`,
// no documentation page mentions it, and this material listed "the Bot Token".
//
// Only the classed blocks - a spec property named by the class enum - because
// those are what a class chooses between, and a reader has to be told which
// fields come with the class they picked.
func checkRequiredBlocks(root string, crds []crd) []string {
	corpus, err := shipped(root)
	if err != nil {
		return []string{err.Error()}
	}
	var out []string
	for _, c := range crds {
		spec := c.spec()
		props, _ := spec["properties"].(map[string]any)
		if props == nil {
			continue
		}
		var classEnum map[string]bool
		for name, raw := range props {
			if !strings.HasSuffix(name, "Class") {
				continue
			}
			schema, _ := raw.(map[string]any)
			values, _ := schema["enum"].([]any)
			classEnum = map[string]bool{}
			for _, v := range values {
				classEnum[fmt.Sprint(v)] = true
			}
			break
		}
		if classEnum == nil {
			continue
		}
		for _, name := range sortedKeys(props) {
			if !classEnum[name] {
				continue
			}
			block, _ := props[name].(map[string]any)
			if block == nil || fmt.Sprint(block["type"]) != "object" {
				continue
			}
			required, _ := block["required"].([]any)
			for _, r := range required {
				field := fmt.Sprint(r)
				// **On a word boundary.** A plain substring test passes on
				// `region` because some page says "regional", which is a false
				// pass in a check whose whole job is to notice an absence.
				if !named(corpus, field) {
					out = append(out, fmt.Sprintf("%s.spec.%s.%s is required and no page in the corpus names it",
						c.Spec.Names.Kind, name, field))
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

// checkCEL holds every CEL-rule count this repository states.
//
// **Two numbers that are easy to write for each other**: 79 is the
// `XValidation` markers in asgard-kube's Go types, 231 is what the generator
// emits from them, because one marker on a struct several kinds embed lands in
// every CRD that embeds it. This material had the marker count written down as
// the CRDs' own, in the one page whose subject is what the CRDs enforce.
func checkCEL(root string, crds []crd, crdDir string) []string {
	types := filepath.Join(filepath.Dir(crdDir), "pkg/apis/asgard/v1alpha1/types.go")
	data, err := os.ReadFile(types)
	if err != nil {
		return []string{fmt.Sprintf("  (the CEL counts need the Go types beside %s; skipped)", crdDir)}
	}
	src := string(data)

	counts := map[string]int{
		"markers":         len(regexp.MustCompile(`XValidation:rule=`).FindAllString(src, -1)),
		"oldself_markers": len(regexp.MustCompile("XValidation:rule=`?\"?self == oldSelf").FindAllString(src, -1)),
	}
	distinct := map[string]bool{}
	for _, c := range crds {
		// **Every node, not only the named properties.** A rule can sit on the
		// object itself rather than on one of its fields - ToolsetSpec's three
		// are the shape - and counting only properties said 219 where the
		// generator emits 231.
		walkAll(c.Spec.Versions[0].Schema.OpenAPIV3Schema, func(node map[string]any) {
			rules, ok := node["x-kubernetes-validations"].([]any)
			if !ok {
				return
			}
			for _, r := range rules {
				m, ok := r.(map[string]any)
				if !ok {
					continue
				}
				text := strings.TrimSpace(fmt.Sprint(m["rule"]))
				counts["rules"]++
				distinct[text] = true
				if text == "self == oldSelf" {
					counts["oldself_rules"]++
				}
			}
		})
	}
	counts["distinct"] = len(distinct)

	// **Whitespace-tolerant, because these documents are hard-wrapped.** A
	// pattern written with a literal space stops matching the moment a rewrap
	// puts a newline inside the phrase, and a claim that matches nothing is
	// reported only when NO claim anywhere matches - so one file drifting out
	// is silent. APPROACH.md was.
	ws := func(p string) *regexp.Regexp { return regexp.MustCompile(strings.ReplaceAll(p, " ", `\s+`)) }
	claims := []struct {
		re    *regexp.Regexp
		names []string
	}{
		{ws(`(\d+) CEL rules written and (\d+) enforced`), []string{"markers", "rules"}},
		{ws(`(\d+) of the CRDs' (\d+) enforced CEL rules`), []string{"oldself_rules", "rules"}},
		{ws(`(\d+) of the enforced rules are exactly ` + "`" + `self == oldSelf` + "`"), []string{"oldself_rules"}},
		{ws(`(\d+) of the (\d+) ` + "`" + `XValidation` + "`" + ` markers`), []string{"oldself_markers", "markers"}},
		{ws(`(\d+) rule instances, (\d+) of them distinct`), []string{"rules", "distinct"}},
	}

	bodies, err := celClaimFiles(root)
	if err != nil {
		return []string{err.Error()}
	}
	var out []string
	seen := 0
	for _, claim := range claims {
		for _, b := range bodies {
			for _, m := range claim.re.FindAllStringSubmatch(b.text, -1) {
				seen++
				for i, name := range claim.names {
					if atoi(m[i+1]) != counts[name] {
						out = append(out, fmt.Sprintf("%s says %s for %s, and asgard-kube has %d",
							b.name, m[i+1], name, counts[name]))
					}
				}
			}
		}
	}
	if seen == 0 {
		out = append(out, fmt.Sprintf(
			"no CEL-rule claim matches any pattern, so %d enforced rules and %d markers are going unchecked",
			counts["rules"], counts["markers"]))
	}
	return out
}

type body struct{ name, text string }

func celClaimFiles(root string) ([]body, error) {
	var out []body
	// **Everything that states one.** APPROACH.md was not in this list and
	// carried the marker count as the CRDs' own for as long as the page whose
	// subject it is did.
	for _, pattern := range []string{"internal/corpus/*/*.md", "internal/gate/*.go",
		"TASK.md", "AGENTS.md", "APPROACH.md"} {
		paths, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			return nil, err
		}
		sort.Strings(paths)
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			rel, _ := filepath.Rel(root, p)
			out = append(out, body{rel, string(data)})
		}
	}
	return out, nil
}

// shipped is everything that reaches a reader: the corpus, what the generator
// writes, the scaffolded templates, and the Go-held bodies.
//
// **Not the prose alone.** A per-class field can be taught by the generator
// that writes it - `athena.outputLocation` is in `internal/generate/dbclass.go`
// and in the db-query skill's connector reference - and reading the wiki alone
// reported ten fields as unnamed that a chart author meets by running `add`.
func shipped(root string) (string, error) {
	var b strings.Builder
	for _, pattern := range []string{
		"internal/corpus/*/*.md", "internal/corpus/*.md",
		"internal/generate/*.go", "internal/generate/templates/*.tmpl",
		"internal/scaffold/templates/**/*.md", "internal/scaffold/templates/*.md",
		"internal/needs/*.go", "internal/brief/*.go", "internal/stage/prompts/*.md",
	} {
		paths, _ := filepath.Glob(filepath.Join(root, pattern))
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				return "", err
			}
			b.Write(data)
			b.WriteString("\n")
		}
	}
	// The scaffolded skills nest deeper than one glob reaches.
	_ = filepath.Walk(filepath.Join(root, "internal/scaffold/templates"), func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		data, err := os.ReadFile(p)
		if err == nil {
			b.Write(data)
			b.WriteString("\n")
		}
		return nil
	})
	return b.String(), nil
}

var wordCache = map[string]*regexp.Regexp{}

// named reports whether the text names this field on a word boundary.
func named(text, field string) bool {
	re, ok := wordCache[field]
	if !ok {
		re = regexp.MustCompile(`\b` + regexp.QuoteMeta(field) + `\b`)
		wordCache[field] = re
	}
	return re.MatchString(text)
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	return n
}
