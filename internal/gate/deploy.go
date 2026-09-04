package gate

import (
	"fmt"
	"sort"
	"strings"
)

// Deployability reports what CD requires of a chart, as opposed to what the
// apiserver requires.
//
// The distinction matters because a chart can be entirely valid - every CR
// applies, every reference resolves, helm lint is green - and still fail the
// deploy. Nothing local was checking these, so the first anyone knew was a red
// tag.
//
// Everything here is a **warning**, and that is the whole design: each condition
// is correct during an onboarding and a problem once someone tags. A project has
// no Syncer until skills are added. Failing on that would leave the gate red
// through the entire middle of the work, which trains people to ignore it - and
// saying nothing means finding out from a run.
func Deployability(docs []Doc, opts Options) Result {
	if len(docs) == 0 {
		return Result{Summary: "nothing rendered yet"}
	}

	ix := newIndex(docs)

	var warnings []string
	warnf := func(format string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	}

	syncers := ix.countsByKind["Syncer"]

	// The CD workflow polls for CronJobs labelled asgard-ai.com/syncer-name and
	// exits 1 after 180 seconds if it finds none, even when helm upgrade
	// succeeded. Its own comment calls zero Syncers a configuration error.
	//
	// A Syncer comes from a SkillSet (the skill set, its own SourceSet and the
	// Syncer that feeds it are generated together) or from a knowledge drive. So
	// this is not obscure: a chart with a read path, an entry point and no
	// skills is exactly the shape a fresh one has.
	if syncers == 0 {
		warnf("no Syncer. The rollout's apply step fires the Syncers this release deploys that carry "+
			"asgard-ai.com/auto-fire-on-rollout and waits for them, so with none there is nothing after "+
			"the dry run that proves the platform accepted any of it - a succeeded run means helm returned. "+
			"`asgard-cli add skillset base --project %s --repo <git url>` creates one, as does "+
			"`asgard-cli add knowledgedrive <name> --project %s`",
			opts.ProjectOr(), opts.ProjectOr())
	}

	// The project environment id is injected as .Values.asgard.projectEnvironmentId
	// on every run, so a CR without the label is a chart that does not read it
	// rather than a value nobody has fetched. When the value is absent the chart
	// deliberately renders no label at all, which is why this is invisible to
	// every other check - and a local render supplies a placeholder, so what this
	// catches is a template that never mentions it.
	var unlabelled []string
	for _, d := range docs {
		switch d.Kind {
		case "Workflow", "Trigger", "CompletionModel":
			if d.Labels[annotationPrefix+"project-environment-id"] == "" {
				unlabelled = append(unlabelled, d.Kind+"/"+d.Name)
			}
		}
	}
	if len(unlabelled) > 0 {
		sort.Strings(unlabelled)
		warnf("%s render without a project-environment-id label. "+
			"On a cluster they work - a Trigger fires on schedule - while their editors open as a blank canvas, which is why this is worth clearing before anyone looks. "+
			"The platform injects the id as .Values.asgard.projectEnvironmentId on every run, so the fix is in the template rather than in a value: "+
			"read it there the way the other CRs of this kind do",
			strings.Join(unlabelled, ", "))
	}

	summary := fmt.Sprintf("%d syncer(s)", syncers)
	if syncers > 0 && len(unlabelled) == 0 {
		summary += ", environment id set"
	}
	return Result{Warnings: warnings, Summary: summary}
}
