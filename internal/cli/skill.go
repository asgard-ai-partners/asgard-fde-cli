package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/skills"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Fetch the reference material for the Asgard server this repository deploys to",
		Long: `Fetch the reference material for the Asgard server this repository deploys to.

An agent writing an Asgard CR needs facts that move with the platform: which
Workflow processors exist, what each config key is called, which are required,
what they default to. Those are not in this binary and must not be.

    asgard-cli skill status    what is here, and what the platform has
    asgard-cli skill update    write what the platform has

The version is a number the platform declares and a person increments. It is
not a digest of the material, deliberately: three upstreams feed it - a
cluster's CRDs, the runtime's own constants, the written documents - and any of
them can move for a cosmetic reason. Deriving the version from them would tell
every repository in the world that it was behind because a map iterated
differently. So the number is a claim somebody makes, and it only goes up.

The consequence worth knowing: the material can change without the number
moving. ` + "`skill status`" + ` says so when it does, and ` + "`skill update`" + ` writes what the
platform has now whatever the number says.

**The reason it is fetched is on-prem.** Asgard runs as a hosted platform and,
for some customers, on their own hardware. A customer's server can be several
versions behind this CLI or ahead of it, and a document saying what
` + "`llm-completion`" + ` takes is only true of one of them. Compiled into the binary,
it would be pinned to whichever release somebody installed - which has nothing
to do with the server their runs deploy to - and there would be no way to
correct it without shipping a new binary. Fetched, it describes the server that
will accept or reject the CR.

The files land in ` + "`.agents/skills/`" + ` (or ` + "`.claude/skills/`" + ` where a repository
already uses that), and they are meant to be committed: whoever clones the
repository, and whatever agent works in it, then has them without a fetch.

They are generated and say so. **Do not edit them** - the next update overwrites
them, and an edit is a claim about the server that the server did not make. An
edited file is reported rather than silently replaced.

What they do not contain is judgement. They are shape - names, types, defaults -
not which mistakes apply cleanly and fail at runtime. That is what the
hand-written skills beside them are for.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newSkillStatusCmd(), newSkillUpdateCmd())
	return cmd
}

// skillRoot resolves where the material goes, and says so, so that a command
// can print a path rather than leave somebody guessing which of two
// directories was written.
func skillRoot(cmd *cobra.Command, override string) (root, repoRoot string, err error) {
	repoRoot, _ = locateRepo(cmd.Context())
	if repoRoot == "" {
		// Outside a checkout the working directory is the only honest answer.
		// It is also usually a mistake, which `status` says out loud.
		repoRoot, err = os.Getwd()
		if err != nil {
			return "", "", err
		}
	}
	return skills.Root(repoRoot, override), repoRoot, nil
}

func newSkillStatusCmd() *cobra.Command {
	var (
		profile string
		dir     string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report which version of the reference material is here, and which the platform has",
		Long: `Report which version of the reference material is in this repository, and
which the platform has.

It writes nothing and always exits 0. The question it answers is whether an
agent working here is reading facts about the server it deploys to, or facts
about some other version of Asgard.

Two versions that differ is not a warning about the past - the material was
right when it was written. It means the server has moved since, and what an
agent reads here now describes something else.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := checkFormat(format); err != nil {
				return err
			}
			root, repoRoot, err := skillRoot(cmd, dir)
			if err != nil {
				return err
			}
			stamp, err := skills.ReadStamp(root)
			if err != nil {
				return err
			}

			pc, err := resolveContext(cmd, contextOptions{Profile: profile})
			if err != nil {
				return err
			}
			remote, err := pc.Client.DocsVersionOnly(cmd.Context())
			if err != nil {
				return err
			}

			local := ""
			fetchedAt := ""
			if stamp != nil {
				local = stamp.Version
				fetchedAt = stamp.FetchedAt
			}

			out := cmd.OutOrStdout()
			if format == formatJSON {
				return writeJSON(out, map[string]any{
					"directory":  root,
					"local":      local,
					"fetched_at": fetchedAt,
					"platform":   remote.Version,
					"profile":    pc.Session.Profile.Name,
					"api":        pc.Session.Profile.API,
					"sources":    remote.Sources,
					"current":    local != "" && local == remote.Version,
				})
			}

			rel := root
			if r, relErr := filepath.Rel(repoRoot, root); relErr == nil {
				rel = r
			}
			fmt.Fprintf(out, "%-11s %s\n", "profile", pc.Session.Profile.Name)
			fmt.Fprintf(out, "%-11s %s\n", "directory", rel)
			fmt.Fprintf(out, "%-11s %s\n", "platform", remote.Version)
			for _, s := range remote.Sources {
				fmt.Fprintf(out, "%-11s   %-12s %s (%d)\n", "", s.Name, s.Digest, s.Records)
			}
			if local == "" {
				fmt.Fprintf(out, "%-11s none\n", "here")
				fmt.Fprintf(out, "\nNothing has been fetched into this repository, so an agent working here\n"+
					"has no statement of what this server accepts.\n\n    asgard-cli skill update\n")
				return nil
			}
			fmt.Fprintf(out, "%-11s %s", "here", local)
			if fetchedAt != "" {
				fmt.Fprintf(out, "  (fetched %s)", fetchedAt)
			}
			fmt.Fprintln(out)

			if local != remote.Version {
				if skills.Behind(local, remote.Version) {
					fmt.Fprintf(out, "\nBehind: the platform has published a newer version of the material.\n\n    asgard-cli skill update\n")
				} else {
					fmt.Fprintf(out, "\nDifferent: this was not written from %s.\n"+
						"A repository can hold material fetched from another platform - check the api line above.\n\n"+
						"    asgard-cli skill update\n", pc.Session.Profile.Name)
				}
				return nil
			}

			// Same version and different content is a state the declared
			// version makes possible on purpose: somebody publishes a change
			// and does not consider it worth telling everybody about. It still
			// changes what an author reads, so it is worth saying here.
			if moved := skills.MovedSources(stamp, sourceDigests(remote.Sources)); len(moved) > 0 {
				fmt.Fprintf(out, "\nSame version, different material: %s moved since this was fetched.\n"+
					"The version is declared rather than derived, so a change reaches you when you ask.\n\n"+
					"    asgard-cli skill update\n", strings.Join(moved, ", "))
				return nil
			}
			fmt.Fprintf(out, "\nCurrent.\n")
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&dir, "dir", "", "skills directory, relative to the repository root; defaults to .agents/skills, or .claude/skills where that exists")
	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	return cmd
}

func newSkillUpdateCmd() *cobra.Command {
	var (
		profile   string
		dir       string
		check     bool
		force     bool
		formatOut string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Write the platform's reference material into this repository",
		Long: `Write the platform's reference material into this repository.

    asgard-cli skill update
    asgard-cli skill update --check    write nothing; exit 1 if it would

Every file is written, not only the changed ones: a partial write leaves a
repository holding two versions at once, and the record beside them would then
be true of neither.

**Commit what it writes.** The point of the files being on disk is that whoever
clones this repository has them without a fetch, and that a change to them shows
up in a diff - which is the changelog. A config key that becomes required is one
changed line in a review.

A file that was edited by hand since it was last written is reported and left
alone. ` + "`--force`" + ` overwrites it. These files are generated and say so at the
top, so an edit is usually somebody correcting what they believed was a mistake
in the material - which belongs upstream, in the server, not in a file the next
update replaces.

--check is for a gate. It exits 1 when the repository is not holding what the
server has, which is the state where an agent's next CR is written against facts
that no longer hold.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := checkFormat(formatOut); err != nil {
				return err
			}
			root, repoRoot, err := skillRoot(cmd, dir)
			if err != nil {
				return err
			}
			stamp, err := skills.ReadStamp(root)
			if err != nil {
				return err
			}

			pc, err := resolveContext(cmd, contextOptions{Profile: profile})
			if err != nil {
				return err
			}
			bundle, err := pc.Client.DocsSkills(cmd.Context())
			if err != nil {
				return err
			}
			if len(bundle.Files) == 0 {
				return fmt.Errorf("the platform sent no files, so there is nothing to write; this is a platform fault, not a local one")
			}

			contents := make([]skills.Content, 0, len(bundle.Files))
			for _, f := range bundle.Files {
				contents = append(contents, skills.Content{Path: f.Path, Content: f.Content})
			}
			plan, err := skills.Plan(root, contents, stamp)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			rel := root
			if r, relErr := filepath.Rel(repoRoot, root); relErr == nil {
				rel = r
			}

			var edited []string
			changes := 0
			for _, f := range plan {
				if f.Status == skills.Edited {
					edited = append(edited, f.Path)
				}
				if f.Status != skills.Unchanged {
					changes++
				}
			}

			if formatOut == formatJSON {
				rows := make([]map[string]string, 0, len(plan))
				for _, f := range plan {
					rows = append(rows, map[string]string{"path": f.Path, "status": string(f.Status)})
				}
				if err := writeJSON(out, map[string]any{
					"directory": rel,
					"version":   bundle.Version,
					"applied":   !check && (force || len(edited) == 0),
					"files":     rows,
				}); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(out, "%s  (%s)\n", bundle.Version, pc.Session.Profile.Name)
				for _, f := range plan {
					fmt.Fprintf(out, "  %-10s %s\n", f.Status, filepath.Join(rel, f.Path))
				}
			}

			if check {
				if changes == 0 {
					if formatOut != formatJSON {
						fmt.Fprintf(out, "\nCurrent.\n")
					}
					return nil
				}
				return fmt.Errorf("%d file(s) are not what the platform has; run `asgard-cli skill update`", changes)
			}

			if len(edited) > 0 && !force {
				return fmt.Errorf("%d generated file(s) were edited here since they were written (%s).\n"+
					"They say not to edit them, so this is probably a correction that belongs in the server rather than\n"+
					"in a file the next update replaces. `--force` overwrites them",
					len(edited), edited[0])
			}

			if err := skills.Apply(root, contents, skills.Stamp{
				Version:   bundle.Version,
				Platform:  pc.Session.Profile.API,
				FetchedAt: time.Now().UTC().Format(time.RFC3339),
				Sources:   sourceDigests(bundle.Sources),
			}); err != nil {
				return err
			}

			if formatOut != formatJSON {
				if changes == 0 {
					fmt.Fprintf(out, "\nAlready current; the record beside them was refreshed.\n")
				} else {
					fmt.Fprintf(out, "\nWrote %d file(s) into %s. Commit them.\n", len(contents), rel)
				}
			}
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&dir, "dir", "", "skills directory, relative to the repository root; defaults to .agents/skills, or .claude/skills where that exists")
	cmd.Flags().BoolVar(&check, "check", false, "write nothing, and exit 1 if this repository is not holding what the platform has")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite a generated file that was edited here")
	cmd.Flags().StringVar(&formatOut, formatFlag, formatText, formatUsage)
	return cmd
}

// sourceDigests flattens the bundle's per-upstream digests for the stamp and
// for the comparison.
func sourceDigests(sources []platform.DocsSource) map[string]string {
	if len(sources) == 0 {
		return nil
	}
	out := make(map[string]string, len(sources))
	for _, s := range sources {
		out[s.Name] = s.Digest
	}
	return out
}

// warnIfBehind says so when the version the platform reported during this
// command differs from what this repository holds.
//
// **It is the mechanism, not a nicety.** An agent finds out its reference
// material is stale only if something it already runs tells it - it is not
// going to remember to poll - so this rides on the header every `/v1/iac`
// response carries and prints at the end of whatever was run.
//
// It is silent when there is nothing to compare: no call was made, or no
// material has ever been fetched here. The one exception is a repository that
// declares a pipeline and holds no material at all, which is an agent writing
// CRs with no statement of what the server accepts.
func warnIfBehind(cmd *cobra.Command) {
	remote := platform.LastDocsVersion()
	if remote == "" {
		return
	}
	// `skill` has already said it, in more detail and with the remedy in
	// context. Repeating it under the command whose job this is reads like a
	// second, different problem.
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "skill" {
			return
		}
	}
	root, repoRoot, err := skillRoot(cmd, "")
	if err != nil {
		return
	}
	stamp, err := skills.ReadStamp(root)
	if err != nil {
		return
	}

	errOut := cmd.ErrOrStderr()
	if stamp == nil {
		if !declaresPipeline(repoRoot) {
			return
		}
		fmt.Fprintf(errOut, "\nthis repository declares a pipeline and holds no reference material for %s\n"+
			"    asgard-cli skill update\n", remote)
		return
	}
	if stamp.Version == remote {
		return
	}
	if skills.Behind(stamp.Version, remote) {
		fmt.Fprintf(errOut, "\nthe platform has published version %s of the reference material; this repository holds %s\n"+
			"    asgard-cli skill update\n", remote, stamp.Version)
		return
	}
	fmt.Fprintf(errOut, "\nthis repository holds version %s of the reference material and this platform serves %s\n"+
		"    asgard-cli skill update\n", stamp.Version, remote)
}

// declaresPipeline reports whether this repository declares one. A repository
// with no declaration is not an IaC repository, and telling it to fetch
// CR-authoring material would be noise.
func declaresPipeline(repoRoot string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, pipelineconfig.FileName))
	return err == nil
}
