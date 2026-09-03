package gate

import (
	"fmt"
	"sort"
)

// The processor contract, extracted from `ProcessorDefinitions` in asgard-core
// `internal/constants.go` with a go/ast walk on 2026-09-03, against
// asgard-kube 15ded0f. `asgard-cli wiki processors` is the prose version and
// carries the same table.
//
// It is a pinned copy, and it is also **incomplete**: `await` is documented on
// `stream-llm-completion-message`, set in five production deployments, and
// declared in neither this file's source nor the CRD. So treat it as the set of
// keys the definitions know about, never as the set a chart may use.
const processorDefsRead = "2026-09-03, asgard-kube 15ded0f"

// processorDef holds only what a rule reads. It carried five more fields -
// optional keys, defaults, whether extra keys are allowed, and the declared
// outputs - and they were removed on 2026-09-03 after the outputs proved
// **wrong**: `ProcessorDefinitions` lists no relationships for `listen-message`
// and Success only for `http-request`, while the documentation gives
// `http-request` a Failure branch producing `prevError` and four production
// charts route one. A table that is partly wrong invites the next rule to be
// built on the wrong part, which is exactly what happened here.
type processorDef struct {
	// mustSet is required with no default. Leaving one out is broken against
	// every version of the platform, which is why W2 fails on it.
	mustSet []string
}

var processorDefs = map[string]processorDef{
	"execute-script":                {mustSet: []string{"engine", "script"}},
	"generate-embedding":            {mustSet: []string{"embeddingModel", "input", "resultField"}},
	"http-request":                  {mustSet: []string{"url", "method"}},
	"listen-message":                {mustSet: nil},
	"llm-completion":                {mustSet: []string{"completionModel", "prompt", "outputSchema"}},
	"llm-query-database":            {mustSet: []string{"semanticLayer", "query", "resultField", "completionModel", "maxTokens"}},
	"push-message":                  {mustSet: nil},
	"query-database":                {mustSet: []string{"dataConnector", "sql", "resultField"}},
	"retrieve-knowledge":            {mustSet: []string{"knowledgeBases", "textQuery", "similarityThreshold", "resultField"}},
	"router":                        {mustSet: nil},
	"stream-llm-completion-message": {mustSet: []string{"completionModel"}},
	"update-context":                {mustSet: nil},
	"validate-payload":              {mustSet: nil},
}

// Processors checks every Workflow's processors against the contract above.
//
//	W1  the type is one of the thirteen the CRD enum allows. A warning: the
//	    enum can gain a value, and a chart using a new one is right while this
//	    copy is stale.
//	W2  every required key with no default is set. **This fails.** A required
//	    key with no fallback is broken against every version of the platform,
//	    and the scaffold itself shipped a Workflow whose only LLM processor had
//	    no completionModel - it rendered, it verified, and it could not run.
//
// **There is no W3.** It was written - flag a key the contract does not declare
// on a processor that takes no arbitrary keys - and it fired on five of five
// production charts, every time for `await`. `await` is documented with real
// semantics on `stream-llm-completion-message` and set in five separate
// deployments, and it is in neither `ProcessorDefinitions` nor the CRD. So
// **the definitions are a subset of what the runtime accepts, not the config
// contract**, and a rule built on treating them as complete calls correct
// charts wrong. It was deleted rather than tuned: a checker that cries wolf
// teaches people to change what it can see rather than what is wrong, which
// this material has already caused once.
//
// The counterpart to W2 is what it deliberately does not check: a required key
// that *does* have a default is left alone, because omitting it is legal. That
// is where `semanticLayer.allowWrite` lives, defaulting to true, and no gate can
// tell a deliberate omission from a forgotten one - which is why the templates
// write it out explicitly instead.
func Processors(docs []Doc, opts Options) Result {
	var problems, warnings []string
	errf := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}
	warnf := func(format string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	}

	seen := 0
	for _, d := range docs {
		if d.Kind != "Workflow" {
			continue
		}
		for _, raw := range digList(d.Spec, "processors") {
			p := mapOf(raw)
			pname := digStr(p, "name")
			ptype := digStr(p, "type")
			if ptype == "" {
				continue
			}
			seen++

			def, known := processorDefs[ptype]
			if !known {
				warnf("W1 Workflow/%s processor %q: type %q is not one of the thirteen this build knows (read %s). Either it is a typo, or the platform has added one and this copy is stale",
					d.Name, pname, ptype, processorDefsRead)
				continue
			}

			set := map[string]bool{}
			for _, c := range digList(p, "configs") {
				if n := digStr(mapOf(c), "name"); n != "" {
					set[n] = true
				}
			}

			for _, key := range def.mustSet {
				if !set[key] {
					errf("W2 Workflow/%s processor %q (%s): %s is required and has no default, so the processor cannot run without it",
						d.Name, pname, ptype, key)
				}
			}

		}
	}

	sort.Strings(problems)
	sort.Strings(warnings)
	return Result{Problems: problems, Warnings: warnings, Summary: fmt.Sprintf("%d processor(s)", seen)}
}

// Tool is one entry exposed to a model as a callable tool.
type Tool struct {
	Workflow    string
	Entry       string
	Name        string
	Description string
}

// Tools returns every tool a render exposes, in workflow order.
//
// There is no rule attached to this and there should not be. `tooling.description`
// is the single field where a wrong value makes a model pick the wrong tool, and
// what makes one wrong is that it does not distinguish itself from the tool
// beside it - which is a property of the set, not of any entry. No check can
// read that; a person reading all of them at once can, and until now there was
// nowhere that put them together.
func Tools(docs []Doc) []Tool {
	var out []Tool
	for _, d := range docs {
		if d.Kind != "Workflow" {
			continue
		}
		for _, raw := range digList(d.Spec, "entries") {
			e := mapOf(raw)
			t := mapOf(e["tooling"])
			if len(t) == 0 {
				continue
			}
			out = append(out, Tool{
				Workflow:    d.Name,
				Entry:       digStr(e, "name"),
				Name:        digStr(t, "name"),
				Description: digStr(t, "description"),
			})
		}
	}
	return out
}
