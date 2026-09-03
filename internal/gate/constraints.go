package gate

import (
	"fmt"
	"regexp"
	"sort"
)

// Field constraints the CRDs enforce beyond enums, from the kubebuilder
// markers in asgard-kube `pkg/apis/asgard/v1alpha1/types.go` at 15ded0f on
// 2026-09-03: Pattern, MinLength, MaxLength, Minimum, Maximum, MinItems and
// MaxItems, keyed by json field name.
//
// **`baseAgentName` is absent although the Go types give it a pattern.** It is
// not a CRD property at all: `SandboxBlueprint.spec.agents` is a **string**
// holding JSON, so the whole subagent shape rides inside it where no schema and
// no CEL rule can see it. C1 walks the decoded document and would never reach
// it. `gate/xref.go` parses that string for the reference check, and
// `asgard-cli brief flow-agent` says why it is the one blueprint worth reading
// by hand.
//
// **`name`, `aliasName` and `key` are dropped.** The first two carry a
// different constraint in different Asgard structs - six sets for `name` alone.
// `key` is worse: it collides with **core Kubernetes**, where
// `secretKeyRef.key` is a Secret's key name and freely contains hyphens, while
// Asgard's `key` is an identifier that may not. Matching on a json field name
// cannot tell those apart, and the first run of this check reported two
// production Syncers for it.
//
// For the same reason the walk skips everything under `valueFrom`: that subtree
// is Kubernetes' own, and none of the constraints here describe it.
//
// Unlike an enum, none of these can go stale in the direction that matters. A
// pattern the apiserver enforces today it enforced yesterday, and loosening one
// would only make this warn where the platform has stopped caring. So **C1 is
// still a warning**, for consistency with E1 and W1, but it is the most
// trustworthy of the three.
const constraintsRead = "2026-09-03, asgard-kube 15ded0f"

type fieldConstraint struct {
	pattern   string
	minLength int
	maxLength int
	minimum   *float64
	maximum   *float64
	minItems  int
	maxItems  int
}

func f64(v float64) *float64 { return &v }

var crdConstraints = map[string]fieldConstraint{
	"aheadSeconds":         {minimum: f64(0)},
	"batchSize":            {minimum: f64(1)},
	"channelMaxIdleMs":     {minimum: f64(1)},
	"chunkSize":            {minimum: f64(1)},
	"columns":              {minItems: 1},
	"configs":              {maxItems: 100},
	"cube":                 {minLength: 1},
	"dataSubPath":          {pattern: "^[^/].*[^/]$|^[^/]$"},
	"delayMs":              {minimum: f64(0)},
	"description":          {minLength: 1},
	"destinationPath":      {minLength: 1, maxLength: 1024},
	"dimensions":           {minItems: 1},
	"drillMembers":         {minItems: 1},
	"executionHistory":     {maxItems: 10},
	"extraDirectories":     {maxItems: 10},
	"fields":               {minItems: 1},
	"folderId":             {minLength: 1},
	"folderPath":           {minLength: 1},
	"host":                 {minLength: 1},
	"index":                {minimum: f64(0)},
	"limitPerUrl":          {minimum: f64(1), maximum: f64(10000)},
	"maxDepth":             {minimum: f64(0), maximum: f64(10)},
	"maxUnsupervisedSteps": {minimum: f64(1)},
	"minLength":            {minimum: f64(0)},
	"mountPath":            {pattern: "^/.+"},
	"oAuthCredentialName":  {pattern: "^[a-z0-9][a-z0-9\\-]*$"},
	"port":                 {minimum: f64(1), maximum: f64(65535)},
	"processors":           {maxItems: 100},
	"query":                {minLength: 1},
	"relationships":        {maxItems: 1000},
	"remotePath":           {minLength: 1},
	"repoUrl":              {minLength: 1},
	"revision":             {minLength: 1},
	"schedule":             {pattern: "^(\\*|([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])|\\*\\/([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])) (\\*|([0-9]|1[0-9]|2[0-3])|\\*\\/([0-9]|1[0-9]|2[0-3])) (\\*|([1-9]|1[0-9]|2[0-9]|3[0-1])|\\*\\/([1-9]|1[0-9]|2[0-9]|3[0-1])) (\\*|([1-9]|1[0-2])|\\*\\/([1-9]|1[0-2])) (\\*|[0-6]|\\*\\/[0-6])$"},
	"siteMapUrl":           {minLength: 1},
	"sourceSetName":        {pattern: "^[a-z0-9][a-z0-9\\-]*$"},
	"sql":                  {minLength: 1},
	"statePath":            {minLength: 1, maxLength: 1024},
	"timeoutMs":            {minimum: f64(1)},
	"urls":                 {minItems: 1},
	"variables":            {maxItems: 1000},
	"waitForMs":            {minimum: f64(0)},
	"workingDirectory":     {pattern: "^/.+"},
	"workingDirectoryPath": {pattern: "^[^/].*/$", maxLength: 1024},
}

// Constraints walks every rendered CR and reports a field outside what the CRD
// allows.
//
//	C1  a string outside its Pattern or length bounds, a number outside its
//	    Minimum/Maximum, or a list outside its MinItems/MaxItems.
//
// The most useful of these in practice is `schedule`: a Trigger's cron
// expression has a pattern that rejects several forms a person would expect to
// work - ranges like `1-5`, lists like `1,15`, and `@daily` - and getting it
// wrong is refused by the apiserver at apply time rather than by helm.
//
// Immutability is deliberately absent. Forty of the 79 CEL rules are
// `self == oldSelf`, which compares a proposed object against the one already
// on the cluster; a render is a single object with no history, so nothing here
// can see it. `botProviderClass` is the one that bites - see
// `asgard-cli usecase chat-channel`.
func Constraints(docs []Doc, opts Options) Result {
	var warnings []string
	checked := 0

	var walk func(path string, v any, kind, name string)
	walk = func(path string, v any, kind, name string) {
		switch t := v.(type) {
		case map[string]any:
			for k, e := range t {
				// Below valueFrom the schema is Kubernetes', not Asgard's.
				if k == "valueFrom" {
					continue
				}
				child := k
				if path != "" {
					child = path + "." + k
				}
				if c, has := crdConstraints[k]; has {
					checked++
					if msg := c.check(e); msg != "" {
						warnings = append(warnings, fmt.Sprintf("C1 %s/%s: %s %s (read %s)",
							kind, name, child, msg, constraintsRead))
					}
				}
				walk(child, e, kind, name)
			}
		case []any:
			for _, e := range t {
				walk(path, e, kind, name)
			}
		}
	}
	for _, d := range docs {
		walk("spec", d.Spec, d.Kind, d.Name)
	}

	sort.Strings(warnings)
	return Result{Warnings: warnings, Summary: fmt.Sprintf("%d constrained field(s) checked", checked)}
}

// check returns why v breaks the constraint, or "" if it does not.
func (c fieldConstraint) check(v any) string {
	switch t := v.(type) {
	case string:
		if c.minLength > 0 && len(t) < c.minLength {
			return fmt.Sprintf("is empty, and the CRD requires at least %d character(s)", c.minLength)
		}
		if c.maxLength > 0 && len(t) > c.maxLength {
			return fmt.Sprintf("is %d characters, and the CRD allows at most %d", len(t), c.maxLength)
		}
		if c.pattern != "" {
			re, err := regexp.Compile(c.pattern)
			// A pattern this build cannot compile is this build's problem, not
			// the chart's, and failing the chart for it would be the worst of
			// both.
			if err == nil && !re.MatchString(t) {
				return fmt.Sprintf("is %q, which does not match the pattern the CRD enforces: %s", t, c.pattern)
			}
		}
	case []any:
		if c.minItems > 0 && len(t) < c.minItems {
			return fmt.Sprintf("has %d entries, and the CRD requires at least %d", len(t), c.minItems)
		}
		if c.maxItems > 0 && len(t) > c.maxItems {
			return fmt.Sprintf("has %d entries, and the CRD allows at most %d", len(t), c.maxItems)
		}
	case int:
		return c.checkNumber(float64(t))
	case float64:
		return c.checkNumber(t)
	}
	return ""
}

func (c fieldConstraint) checkNumber(n float64) string {
	if c.minimum != nil && n < *c.minimum {
		return fmt.Sprintf("is %g, and the CRD's minimum is %g", n, *c.minimum)
	}
	if c.maximum != nil && n > *c.maximum {
		return fmt.Sprintf("is %g, and the CRD's maximum is %g", n, *c.maximum)
	}
	return ""
}
