package gate

import (
	"strings"
	"testing"
)

// **An Agent's `sourceSetMounts` names a SourceSet and went unresolved.**
//
// The blueprint form has been checked since this file was written; the Agent
// form is a real YAML list where the blueprint's is a stringified JSON string,
// and it reads so differently that it looked like a different field. It is not.
// It fails the same way and worse: the CR is valid, the apiserver accepts it,
// `gate` was green, and at runtime the agent simply cannot see the files it was
// given - with nothing anywhere saying why.
func TestAgentSourceSetMountIsResolved(t *testing.T) {
	agent := Doc{Kind: "Agent", Name: "ag-x", Spec: map[string]any{
		"managed": map[string]any{
			"skillSetNames": []any{"sk-x"},
			"sourceSetMounts": []any{
				map[string]any{
					"sourceSetName": "ss-missing",
					"mountPath":     "/work/x",
					"readOnly":      true,
				},
			},
		},
	}}
	skillset := Doc{Kind: "SkillSet", Name: "sk-x", Spec: map[string]any{}}

	got := joined(Xref([]Doc{agent, skillset}, Options{}))
	if !strings.Contains(got, "sourceSetMounts") || !strings.Contains(got, "ss-missing") {
		t.Errorf("a mount naming a SourceSet that is not in the render went unreported:\n%s", got)
	}

	// **And it must not fire when the SourceSet is there**, which is the half
	// that decides whether the rule can stay: a check that reports a correct
	// chart is one somebody turns off.
	present := Doc{Kind: "SourceSet", Name: "ss-missing", Spec: map[string]any{}}
	if got := joined(Xref([]Doc{agent, skillset, present}, Options{})); strings.Contains(got, "sourceSetMounts") {
		t.Errorf("a mount naming a SourceSet the render declares was reported:\n%s", got)
	}
}

// joined flattens a Result for assertion.
func joined(r Result) string {
	return strings.Join(append(append([]string{}, r.Problems...), r.Warnings...), "\n")
}
