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

// **The Secret the platform guarantees is not a dangling reference.**
//
// A chart that reads `preset-agent-hub` / `api_key` needs no value obtained and
// no key declared - which is the whole point, since the value has no documented
// route out of the namespace. Verified against a deployed namespace before this
// was relaxed: admission accepted it, apply succeeded, the SourceSet came up
// Ready and its Syncer was provisioned.
func TestPlatformOwnedSecretIsNotDangling(t *testing.T) {
	doc := Doc{Kind: "SourceSet", Name: "ss-rma", Spec: map[string]any{
		"apiKey": map[string]any{"valueFrom": map[string]any{
			"secretKeyRef": map[string]any{"name": "preset-agent-hub", "key": "api_key"},
		}},
	}}
	opts := Options{Release: "cs-dev", DeclaredKeys: map[string]map[string]bool{
		"secretKeyRef": {}, "configMapKeyRef": {},
	}}
	if got := Placeholder(CredentialRefs([]Doc{doc}, opts)); got != "" {
		t.Errorf("the platform's own Secret was reported:\n%s", got)
	}

	// **The common error still fails.** A name that is neither the release's
	// own object nor the platform's resolves to nothing, and says so only at
	// runtime.
	doc.Spec["apiKey"].(map[string]any)["valueFrom"].(map[string]any)["secretKeyRef"].(map[string]any)["name"] = "some-other-secret"
	if got := Placeholder(CredentialRefs([]Doc{doc}, opts)); got == "" {
		t.Error("a Secret that is nobody's was not reported")
	}
}

// **A key that is declared, set, and read by nothing is the direction nothing
// reported.** `variables list` marks a value with no declaration ORPHAN, and
// the run reports `vars/orphan`; the reverse - a declaration with no reader -
// is created and injected on every run into a Secret no CR names, so it shows
// as a value that is present and working. It arrives when a shape changes and
// the template that read the key is replaced, which leaves a live credential
// declared for a system the chart no longer talks to.
func TestCredentialRefsReportsADeclaredKeyNothingReads(t *testing.T) {
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
		"secretKeyRef":    {"erp_db_password": true, "old_api_key": true},
		"configMapKeyRef": {"retired_base_url": true},
	}}

	got := Placeholder(CredentialRefs([]Doc{doc}, opts))
	for _, want := range []string{`"old_api_key"`, `"retired_base_url"`} {
		if !strings.Contains(got, want) {
			t.Errorf("a declared key nothing reads went unreported (%s):\n%s", want, got)
		}
	}
	// The one that IS read must not be reported in either direction.
	if strings.Contains(got, `"erp_db_password"`) {
		t.Errorf("a declared key that the render reads was reported:\n%s", got)
	}

	// **Nil is "the declaration was not read".** Same rule as the other
	// direction: with nothing to compare against, say nothing.
	opts.DeclaredKeys = nil
	if got := Placeholder(CredentialRefs([]Doc{doc}, opts)); strings.Contains(got, "nothing in this render reads it") {
		t.Errorf("reported an unread declaration with no declaration to compare against:\n%s", got)
	}
}

// **A key read only through the platform's own Secret is not a reader of the
// release's declaration.** `preset-agent-hub` carries `api_key`, which the
// platform mints and nobody declares; a release that also declares its own
// `api_key` has one nothing reads, and the exemption must not hide that.
func TestPlatformOwnedReadDoesNotSatisfyADeclaration(t *testing.T) {
	doc := Doc{Kind: "SourceSet", Name: "ss-rma", Spec: map[string]any{
		"apiKey": map[string]any{"valueFrom": map[string]any{
			"secretKeyRef": map[string]any{"name": "preset-agent-hub", "key": "api_key"},
		}},
	}}
	opts := Options{Release: "cs-dev", DeclaredKeys: map[string]map[string]bool{
		"secretKeyRef": {"api_key": true}, "configMapKeyRef": {},
	}}
	if got := Placeholder(CredentialRefs([]Doc{doc}, opts)); !strings.Contains(got, "nothing in this render reads it") {
		t.Errorf("a declaration satisfied only by the platform's own Secret passed:\n%s", got)
	}
}
