package gate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

const (
	publishedLabel = annotationPrefix + "agent-published"

	// minSampleQuestions is what a published Agent needs. Unpublished ones are
	// exempt: they are not going to be given work yet.
	minSampleQuestions = 2
)

// AgentSplit checks the "one system, one Agent" split.
//
// None of these break helm lint or apply. They make the orchestrator route
// wrongly, or let an agent's search space grow back:
//
//	R1   at most one semantic layer per Agent, and no layer bound twice. At
//	     most, not exactly: zero is legal, and its search space cannot grow.
//	R1b  but an Agent cannot have no capability at all - a layer, a Toolset or a
//	     SkillSet, at least one - so that deleting one by accident does not pass
//	     quietly. **SkillSet was missing from that list and the rule was wrong
//	     about a running deployment**: every subagent of a flow-agent supervisor
//	     mounts skills and nothing else, and nine of them were told they had no
//	     capability at all.
//	R4   no Agent sets allowedCubes, which keeps the standing decision that a
//	     bound layer is queryable in full.
//	R7   a published Agent has at least two sampleQuestions. Published is the
//	     agent-published label, and that label is the only gate on whether a
//	     caller includes the Agent in agent_hub.agent_names.
//	R10  no Agent references an OLAP-only layer.
//	R11  a SemanticLayer that no Agent binds is either recorded as OLAP-only or
//	     reported, so that "deliberately unbound" and "somebody forgot" stop
//	     looking identical. A warning, because both are legitimate mid-onboarding.
//	R12  prompt.task and prompt.format are byte-identical across every Agent in
//	     one render. An Agent CR has no include mechanism, so a shared section
//	     can only be copied; keeping the copies identical is what lets a later
//	     change be one substitution and be verified with a diff.
//
// Zero Agents is legal and passes: a pure Flow Agent project keeps its prompt in
// a Workflow and its capability in a SandboxBlueprint, so its chart has no Agent
// CR at all. The summary says "0 agent(s)" so that a project that lost its
// Agents by accident is visible to a reviewer.
// blueprintAgents names every Agent a SandboxBlueprint mounts as a subagent.
//
// `spec.agents` is a JSON string holding the array - that is how the CRD defines
// it - so a name is only visible after parsing the string. A parse failure
// yields nothing rather than an error: `gate.Xref` already reports invalid JSON
// there, and reporting it twice from two checks reads as two defects.
func blueprintAgents(docs []Doc) map[string]bool {
	out := map[string]bool{}
	for _, d := range docs {
		if d.Kind != "SandboxBlueprint" {
			continue
		}
		raw := digStr(d.Spec, "agents", "value")
		if raw == "" {
			continue
		}
		var agents []struct {
			BaseAgentName string `json:"baseAgentName"`
		}
		if json.Unmarshal([]byte(raw), &agents) != nil {
			continue
		}
		for _, a := range agents {
			if a.BaseAgentName != "" {
				out[a.BaseAgentName] = true
			}
		}
	}
	return out
}

func AgentSplit(docs []Doc, opts Options) Result {
	ix := newIndex(docs)
	agents := ix.of("Agent")
	subagents := blueprintAgents(docs)

	olapOnly := map[string]bool{}
	for _, name := range opts.OLAPOnlyLayers {
		olapOnly[name] = true
	}

	var problems []string
	errf := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}
	var warnings []string
	warnf := func(format string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	}

	boundBy := map[string]string{} // layer -> the first Agent that bound it
	published := 0
	var allLayers []string

	for _, a := range agents {
		managed := mapOf(a.Spec["managed"])
		layers := digList(managed, "semanticLayers")

		if len(layers) > 1 {
			names := make([]string, 0, len(layers))
			for _, l := range layers {
				names = append(names, digStr(mapOf(l), "name"))
			}
			errf("R1 %s: has %d semanticLayers, at most 1 is allowed (%s)",
				a.Name, len(layers), strings.Join(names, ", "))
		}

		// A SkillSet is a capability source too, and leaving it out made this
		// rule wrong about nine Agents in a deployment that is running: every
		// subagent of a flow-agent supervisor mounts skills and nothing else,
		// and the message said it had "no source of capability at all" while it
		// had one. Found by running the gate over the reference deployments
		// rather than over a chart this tool generated.
		if len(layers) == 0 &&
			len(strList(managed["toolsetNames"])) == 0 &&
			len(strList(managed["skillSetNames"])) == 0 {
			errf("R1b %s: has no semanticLayers, no toolsetNames and no skillSetNames, so this Agent has no source of capability at all", a.Name)
		}

		for _, item := range layers {
			layer := mapOf(item)
			name := digStr(layer, "name")
			if name == "" {
				name = "<unnamed>"
			}
			allLayers = append(allLayers, name)

			if first, ok := boundBy[name]; ok {
				errf("R1 %s: %s is already bound by %s, and a semantic layer should have exactly one Agent",
					a.Name, name, first)
			} else {
				boundBy[name] = a.Name
			}

			if len(digList(layer, "allowedCubes")) > 0 {
				errf("R4 %s: %s sets allowedCubes, against the standing decision that a bound layer is queryable in full",
					a.Name, name)
			}

			if olapOnly[name] {
				errf("R10 %s: references %s, which is a Data Insight OLAP store and should be bound to no Agent",
					a.Name, name)
			}
		}

		if strings.EqualFold(a.Labels[publishedLabel], "true") {
			published++
			if n := len(digList(managed, "sampleQuestions")); n < minSampleQuestions {
				errf("R7 %s: published but has %d sampleQuestions, and needs at least %d (set %s to \"false\" for an unverified agent rather than leaving the questions empty)",
					a.Name, n, minSampleQuestions, publishedLabel)
			}
		}
	}

	// R12 is checked across all Agents at once, since it is about them agreeing.
	// R12 is an agent-hub rule and applies to agent-hub Agents. A subagent of a
	// flow-agent supervisor is excluded, measured rather than reasoned: across
	// every reference deployment the five agent-hub Agents share **one**
	// prompt.task, and the seventeen blueprint subagents have thirteen distinct
	// ones - six of them empty, because their prompt lives on the Workflow's
	// processor instead. Three supervisor deployments out of three, so it is the
	// convention and not a mistake three engagements made. A subagent's task is
	// what makes it a specialist; requiring them all to match cancels the split
	// the shape exists for.
	hub := make([]Doc, 0, len(agents))
	for _, a := range agents {
		if !subagents[a.Name] {
			hub = append(hub, a)
		}
	}
	if len(hub) > 1 {
		for _, field := range []string{"task", "format"} {
			byValue := map[string][]string{}
			for _, a := range hub {
				value := digStr(mapOf(a.Spec["managed"]), "prompt", field)
				byValue[value] = append(byValue[value], a.Name)
			}
			if len(byValue) > 1 {
				var groups []string
				for value, names := range byValue {
					sort.Strings(names)
					groups = append(groups, fmt.Sprintf("%s (%d chars)", strings.Join(names, "+"), len(value)))
				}
				sort.Strings(groups)
				errf("R12 prompt.%s differs: %d distinct values - %s",
					field, len(byValue), strings.Join(groups, " | "))
			}
		}
	}

	// R11. A layer nobody binds is either a Data Insight store, which is
	// correct and permanent, or a read path somebody has not finished wiring,
	// which is temporary - and the rendered chart cannot tell them apart. The
	// danger is not the layer sitting there: it is the later reader who binds it
	// to an Agent as a tidy-up, and so gives an agent restricted to an API a
	// second path straight into the database. R10 catches that only for layers
	// already recorded as OLAP-only, so the recording has to happen while
	// somebody still knows which kind it is.
	//
	// A warning, not a failure: a layer added before its Agent is the normal
	// order of work, and failing here would make the gate red for doing the
	// steps in the order the stages ask for.
	for _, d := range ix.of("SemanticLayer") {
		if boundBy[d.Name] != "" || olapOnly[d.Name] || mentionedElsewhere(docs, d) {
			continue
		}
		warnf("R11 %s has no consumer: no Agent binds it and nothing else in the render mentions it. If that is deliberate - a Data Insight OLAP store, read through Mimir rather than by an agent - record it with `asgard-cli verify --olap-only-layer %s`, which writes it to %s and makes R10 refuse any later attempt to bind it to an Agent. If it is not deliberate, its read path is unfinished. Record it either way while you still know which it is: the render cannot tell them apart, and the reader who finds it later is the one who binds it as a tidy-up",
			d.Name, d.Name, config.FileName)
	}

	sort.Strings(problems)
	sort.Strings(warnings)
	sort.Strings(allLayers)

	summary := fmt.Sprintf("%d agent(s) (%d published)", len(agents), published)
	if len(allLayers) > 0 {
		summary += ", layers: " + strings.Join(allLayers, ", ")
	}
	return Result{Problems: problems, Warnings: warnings, Summary: summary}
}

// mentionedElsewhere reports whether any document other than the layer itself
// names it.
//
// An Agent binds a layer in a structured field, but a Flow Agent project has no
// Agent CR at all: its layer is mounted on an LLM processor, and a processor
// config's value is a string. In the charts that value is an expression holding
// a JSON array - `[{"name": "sl-x", "allowQuery": true, ...}]` - so there is no
// field to dig for, and asking "which field references a layer" gets a
// different answer for every processor.
//
// So the test is textual and deliberately broad: a layer that no other document
// mentions anywhere is one nothing can be using. Broad in the safe direction -
// it can only suppress the warning, never raise a false one, and a name that
// appears in an unrelated comment is a name somebody chose to write down.
func mentionedElsewhere(docs []Doc, layer Doc) bool {
	for _, d := range docs {
		if d.Kind == layer.Kind && d.Name == layer.Name {
			continue
		}
		if mentions(d.Spec, layer.Name) {
			return true
		}
		for _, m := range []map[string]string{d.Labels, d.Annotations} {
			for _, v := range m {
				if strings.Contains(v, layer.Name) {
					return true
				}
			}
		}
	}
	return false
}

// mentions walks a decoded document looking for the name in any string.
func mentions(v any, name string) bool {
	switch t := v.(type) {
	case string:
		return strings.Contains(t, name)
	case map[string]any:
		for _, e := range t {
			if mentions(e, name) {
				return true
			}
		}
	case []any:
		for _, e := range t {
			if mentions(e, name) {
				return true
			}
		}
	}
	return false
}
