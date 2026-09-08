package gate

import (
	"strings"
	"testing"
)

// The bug this exists for, and the two ways of not reporting it.
//
// A wrong Secret name is a valid string, so the CR is valid, the CRD's rules
// pass, the server dry run passes and apply succeeds. Nothing but this says
// anything, and what it says has to be specific enough to act on.
func TestCredentialRefs(t *testing.T) {
	doc := func(name string) Doc {
		return Doc{Kind: "DataConnector", Name: "dc-uof", Spec: map[string]any{
			"mssql": map[string]any{
				"password": map[string]any{
					"valueFrom": map[string]any{
						"secretKeyRef": map[string]any{"key": "uof_db_password", "name": name},
					},
				},
			},
		}}
	}

	t.Run("the name the platform injects passes", func(t *testing.T) {
		r := CredentialRefs([]Doc{doc("iac-app-dev-app-secret")}, Options{Release: "app-dev"})
		if len(r.Warnings) != 0 {
			t.Fatalf("warned on the correct name: %q", r.Warnings)
		}
	})

	// This is the literal the broken helper fell back to. It is a real Secret
	// in the Terraform-provisioned demo namespaces, which is why it read as
	// correct - and absent from every Release namespace.
	t.Run("a literal app-secret is reported with both names", func(t *testing.T) {
		r := CredentialRefs([]Doc{doc("app-secret")}, Options{Release: "app-dev"})
		if len(r.Warnings) != 1 {
			t.Fatalf("want 1 warning, got %q", r.Warnings)
		}
		for _, want := range []string{"app-secret", "iac-app-dev-app-secret", "dc-uof"} {
			if !strings.Contains(r.Warnings[0], want) {
				t.Errorf("warning does not name %q: %s", want, r.Warnings[0])
			}
		}
	})

	// Silence here would be false confidence about exactly this bug, and an
	// "ok" summary saying "skipped" would read as a pass - Result has no
	// skipped state.
	t.Run("no release name says the references went unchecked", func(t *testing.T) {
		r := CredentialRefs([]Doc{doc("anything")}, Options{})
		if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "unchecked") {
			t.Fatalf("want an unchecked warning, got %q / %q", r.Warnings, r.Summary)
		}
	})

	// Nothing to check is not a finding.
	t.Run("a stream with no references says nothing", func(t *testing.T) {
		r := CredentialRefs([]Doc{{Kind: "Workflow", Name: "wf", Spec: map[string]any{"a": "b"}}}, Options{})
		if len(r.Warnings) != 0 {
			t.Fatalf("warned with nothing to check: %q", r.Warnings)
		}
	})
}
