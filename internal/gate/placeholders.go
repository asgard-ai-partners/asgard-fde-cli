package gate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// placeholder is what `asgard-cli add` leaves where a person has to answer.
//
// **It is deliberately a bare word.** The generator writes "TODO" and nothing
// around it, because a field that looks decided but never was is worse than an
// empty one - so the same word is what survives into the render when nobody
// answered.
var placeholder = regexp.MustCompile(`\bTODO\b`)

// reachesAReader names the fields a placeholder is worst in, and says who meets
// it. Everything else is reported as a count; these are named one by one.
//
// **A chart full of TODOs renders, lints, applies and deploys green.** That is
// by design at every other step - the apiserver has no opinion about the string
// "TODO" - so this is the only place between `add` and a tag where anybody says
// they are still there.
var reachesAReader = map[string]map[string]string{
	"Agent": {
		"managed.description":     "the orchestrator reads this to decide whether to delegate here",
		"managed.sampleQuestions": "the customer sees these in Sindri, before anything the agent says",
		"managed.prompt.persona":  "the agent answers in the register its prompt is written in",
		"managed.prompt.task":     "the agent answers in the register its prompt is written in",
	},
	"SemanticLayer": {
		"instruction": "the model reads this to decide what to query",
	},
	"SkillSet": {
		"searchPaths": "a path that is not a skill directory resolves to no skills at all, and nothing else reports it",
	},
}

// Placeholders reports the generator's TODOs that are still in the render.
//
// **Warnings, never failures.** A chart is full of them through the whole
// middle of an onboarding, and failing there would leave the gate red for days
// - which trains people to ignore it, the same reasoning Deployability is built
// on. What this is for is the last read before a tag.
func Placeholders(docs []Doc, opts Options) Result {
	if len(docs) == 0 {
		return Result{Summary: "nothing rendered yet"}
	}

	var warnings []string
	total := 0
	byKind := map[string]int{}

	for _, d := range docs {
		n := countPlaceholders(d.Spec) + countPlaceholders(map[string]any{
			"annotations": toAny(d.Annotations),
		})
		if n == 0 {
			continue
		}
		total += n
		byKind[d.Kind] += n

		named := reachesAReader[d.Kind]
		var hit []string
		for path, why := range named {
			if placeholder.MatchString(fmt.Sprint(digAny(d.Spec, strings.Split(path, ".")...))) {
				hit = append(hit, fmt.Sprintf("%s (%s)", path, why))
			}
		}
		sort.Strings(hit)
		for _, h := range hit {
			warnings = append(warnings, fmt.Sprintf("%s/%s: %s is still TODO - %s", d.Kind, d.Name, strings.SplitN(h, " (", 2)[0], strings.TrimSuffix(strings.SplitN(h, " (", 2)[1], ")")))
		}
	}

	if total == 0 {
		return Result{Summary: "no generator TODO left in the render"}
	}
	kinds := make([]string, 0, len(byKind))
	for _, k := range sortedKindKeys(byKind) {
		kinds = append(kinds, fmt.Sprintf("%s=%d", k, byKind[k]))
	}
	warnings = append(warnings, fmt.Sprintf(
		"%d TODO(s) remain across the render (%s). `asgard-cli add` writes them where somebody has to answer, "+
			"and nothing between here and a tag mentions them again: helm renders the word, the apiserver accepts it, "+
			"and the run succeeds", total, strings.Join(kinds, ", ")))
	return Result{Warnings: warnings, Summary: fmt.Sprintf("%d TODO(s)", total)}
}

func sortedKindKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func toAny(m map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

// countPlaceholders walks any decoded YAML and counts the strings carrying one.
func countPlaceholders(node any) int {
	switch v := node.(type) {
	case string:
		return len(placeholder.FindAllString(v, -1))
	case map[string]any:
		n := 0
		for _, e := range v {
			n += countPlaceholders(e)
		}
		return n
	case []any:
		n := 0
		for _, e := range v {
			n += countPlaceholders(e)
		}
		return n
	}
	return 0
}

// digAny follows a dotted path and returns whatever is there, list or scalar.
func digAny(m map[string]any, path ...string) any {
	var cur any = m
	for _, p := range path {
		node, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = node[p]
	}
	return cur
}
