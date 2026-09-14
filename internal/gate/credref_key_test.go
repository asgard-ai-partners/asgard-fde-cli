package gate

import (
	"strings"
	"testing"
)

// **A reference to the right Secret with a key nothing declares is injected by
// nothing.** The declaration says so about itself: an undeclared key is stored
// and never injected, `variables list` marks it ORPHAN, the run reports
// `vars/orphan`, and lint, render and the dry run all stay green while the CR
// resolves to nothing at runtime.
//
// `asgard-cli add dataconnector` generates the reference and leaves the
// declaring to somebody. Across the nine DataConnector classes that is fourteen
// keys, and before this nothing said they were undeclared - the credential
// check looked at the object's name and passed.
func TestCredentialRefsChecksTheKeyNotOnlyTheObject(t *testing.T) {
	doc := Doc{Kind: "DataConnector", Name: "dc-erp", Spec: map[string]any{
		"postgres": map[string]any{
			"password": map[string]any{
				"valueFrom": map[string]any{
					"secretKeyRef": map[string]any{
						"name": "iac-erp-dev-app-secret",
						"key":  "erp_db_password",
					},
				},
			},
		},
	}}
	opts := Options{Release: "erp-dev", DeclaredKeys: map[string]map[string]bool{
		"secretKeyRef": {}, "configMapKeyRef": {},
	}}

	got := Placeholder(CredentialRefs([]Doc{doc}, opts))
	if !strings.Contains(got, "declares it for no release") {
		t.Errorf("an undeclared key passed:\n%s", got)
	}

	opts.DeclaredKeys["secretKeyRef"]["erp_db_password"] = true
	if got := Placeholder(CredentialRefs([]Doc{doc}, opts)); strings.Contains(got, "declares it for no release") {
		t.Errorf("a declared key was still reported:\n%s", got)
	}

	// **Nil is "the declaration was not read", not "it declares none."** A
	// check that cannot see the declaration must say nothing rather than
	// report every key in the chart.
	opts.DeclaredKeys = nil
	if got := Placeholder(CredentialRefs([]Doc{doc}, opts)); strings.Contains(got, "declares it for no release") {
		t.Errorf("reported an undeclared key with no declaration to compare against:\n%s", got)
	}
}

// Placeholder joins a Result's warnings and problems for assertion.
func Placeholder(r Result) string {
	return strings.Join(append(append([]string{}, r.Problems...), r.Warnings...), "\n")
}
