package gate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/render"
)

// CredentialRefs checks that every credential reference points at an object
// that will exist.
//
// A Release owns exactly two: `iac-<release>-app-secret` and
// `iac-<release>-app-config`, both created and maintained by the Platform,
// whose names arrive as `.Values.asgard.appSecretName` and
// `.Values.asgard.appConfigMapName`. A chart may not declare a Secret or a
// ConfigMap of its own - the runner's kind whitelist rejects anything outside
// the asgard-ai.com group - so those two names are the only ones a
// `secretKeyRef` or `configMapKeyRef` can legitimately carry.
//
// Nothing else sees this. A name is a valid string, so the CR is valid, the
// CRD's own rules pass, the server dry run passes, apply succeeds, and the
// reference resolves to nothing at runtime with no error on the CR. It shipped
// exactly once and cost a deployment its database password: a chart helper read
// `.Values.appSecretName` instead of `.Values.asgard.appSecretName` and fell
// back to a literal `app-secret` - a name that is real in the
// Terraform-provisioned demo namespaces and absent from every Release
// namespace, which is why it read as correct all the way through.
//
// The local render supplies the real names, so what this catches is a template
// that wrote a name instead of reading one. With no release name - a stream
// rendered elsewhere - there is nothing to compare against and the check says
// so rather than passing quietly.
func CredentialRefs(docs []Doc, opts Options) Result {
	if len(docs) == 0 {
		return Result{Summary: "nothing rendered yet"}
	}
	// With no release name the expected names are unknown. That is reported
	// only when there is in fact something it would have checked: a `Result`
	// carries no "skipped", so a summary saying so prints under `ok` and reads
	// as a pass - and silence on a stream with no credential references at all
	// would be noise, not a finding.
	want := map[string]string{}
	if opts.Release != "" {
		want["secretKeyRef"] = render.AppSecretName(opts.Release)
		want["configMapKeyRef"] = render.AppConfigMapName(opts.Release)
	} else {
		want["secretKeyRef"] = ""
		want["configMapKeyRef"] = ""
	}

	var warnings []string
	checked := 0
	seen := map[string]bool{}

	var walk func(v any, kind, name string)
	walk = func(v any, kind, name string) {
		switch t := v.(type) {
		case map[string]any:
			for ref, expect := range want {
				block, ok := t[ref].(map[string]any)
				if !ok {
					continue
				}
				got, _ := block["name"].(string)
				if got == "" {
					continue
				}
				checked++
				if expect == "" || got == expect {
					continue
				}
				msg := fmt.Sprintf("%s/%s: %s.name is %q, and the only %s this release has is %q",
					kind, name, ref, got, refObject(ref), expect)
				if !seen[msg] {
					seen[msg] = true
					warnings = append(warnings, msg)
				}
			}
			for _, child := range t {
				walk(child, kind, name)
			}
		case []any:
			for _, child := range t {
				walk(child, kind, name)
			}
		}
	}

	for _, d := range docs {
		walk(d.Spec, d.Kind, d.Name)
	}

	if len(warnings) > 0 {
		sort.Strings(warnings)
		for i, w := range warnings {
			warnings[i] = w + ". The Platform injects the name on every run, so the fix is to read it in the " +
				"template - `{{ include \"<chart>.appSecretName\" . }}` - rather than to create the object or " +
				"rename anything. Nothing downstream reports this: the CR is valid and the failure is at runtime"
		}
		return Result{Warnings: warnings, Summary: fmt.Sprintf("%d credential reference(s), %d not resolvable", checked, len(warnings))}
	}
	if checked == 0 {
		return Result{Summary: "no credential references"}
	}
	if opts.Release == "" {
		return Result{
			Warnings: []string{fmt.Sprintf(
				"%d credential reference(s) went unchecked: with no release name the injected Secret and "+
					"ConfigMap names are unknown, so a reference to an object that will not exist cannot be "+
					"told apart from a correct one. Render through `asgard-cli verify <release>` to check them",
				checked)},
			Summary: fmt.Sprintf("%d credential reference(s), none checked", checked),
		}
	}
	return Result{Summary: fmt.Sprintf("%d credential reference(s), all to this release's own objects", checked)}
}

func refObject(ref string) string {
	if strings.HasPrefix(ref, "config") {
		return "ConfigMap"
	}
	return "Secret"
}
