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
// is correct during an onboarding and fatal once someone tags. A project has no
// Syncer until skills are added, and platformMainEnvironmentId does not exist
// until tf-asgard has created the namespace. Failing on either would leave the
// gate red through the entire middle of the work, which trains people to ignore
// it - and saying nothing means finding out from CD.
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
	// this is not obscure: a project with a read path, an entry point and no
	// skills is exactly the shape a fresh chart has, and the one shape CD
	// refuses.
	if syncers == 0 {
		warnf("no Syncer. Whether that fails your CD depends on one `if` in its workflow: some poll for CronJobs labelled syncer-name and exit 1 after 180s when none appear, even after a successful helm upgrade, and some count what the chart declares first and skip the step at zero. A production chart runs today with no Syncer under the second kind. "+
			"Check with `grep -n syncer-name -A15 .github/workflows/*.y*ml` before the first tag. If it waits unconditionally, "+
			"`asgard-cli add skillset base --project %s --repo <git url>` creates one, as does "+
			"`asgard-cli add knowledgedrive <name> --project %s`. "+
			"If it skips, nothing in the pipeline is checking this project at all - green means helm returned",
			opts.ProjectOr(), opts.ProjectOr())
	}

	// platformMainEnvironmentId is per project per environment and the platform
	// only issues it after tf-asgard creates the namespace. When it is empty the
	// chart deliberately renders no label at all, which is why this is invisible
	// to every other check.
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
		warnf("platformMainEnvironmentId is empty, so %s render without a project-environment-id label. "+
			"On a cluster they work - a Trigger fires on schedule - while their editors open as a blank canvas, which is why this is worth clearing before anyone looks. "+
			"The platform issues the id after tf-asgard creates the namespace; put it in projects/%s/chart/values-<env>.yaml before tagging",
			strings.Join(unlabelled, ", "), opts.ProjectOr())
	}

	summary := fmt.Sprintf("%d syncer(s)", syncers)
	if syncers > 0 && len(unlabelled) == 0 {
		summary += ", environment id set"
	}
	return Result{Warnings: warnings, Summary: summary}
}
