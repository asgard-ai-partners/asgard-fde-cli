// Package cli assembles the asgard-cli command tree.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

// NewRootCmd builds the root command. Every call returns a fresh tree so tests
// cannot interfere with one another.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asgard-cli",
		Short: "Command line tool for Asgard FDE",
		Long: `asgard-cli is what an agent asks about integrating with Asgard, so that
an FDE can walk into a customer's room with it.

BEFORE A MEETING, run the one for what you are about to do:

    asgard-cli brief customer-meeting    the five things said wrong to a customer
    asgard-cli brief write-chart         before touching a chart
    asgard-cli brief handover            before telling anyone it is live
    asgard-cli guide requirements        the interview, and what to ask for

**The riskiest thing in an engagement leaves no trace in a repository.** Talking
to a customer changes no file, so nothing derived from what the repo contains can
prepare anybody for it - and every entry in "brief" is there because somebody
has actually got it wrong, not because it is important.

It does two things after that, and they are reached differently.

ASKING - what the platform has, which CR a UI name maps to, how one shape is
assembled field by field, and where each has been got wrong before:

    asgard-cli find <terms>

That searches all four parts of the material at once - the platform wiki, the
deployment extracts, the guidance for each decision, and the skills the agent in
a customer repo loads - and hands over the counterpart of whatever it finds. Ask
in Chinese if that is the language the question was asked in; the glossary
carries the translation. **This works with no repository**, which is the point:
the question gets asked in a meeting, before there is a directory.

    asgard-cli wiki <page>     the platform
    asgard-cli usecase <name>  one deployment shape, field by field
    asgard-cli guide <name>    one decision, and how it has been got wrong
    asgard-cli brief <what>    the thing you are about to do

BUILDING - a chart of Asgard custom resources per project, each deployed to its
own namespace:

    asgard-cli init          onboard a repository: skeleton, binding, material
    asgard-cli project       every chart: its shape, what it declares, what it lacks
    asgard-cli question      what nobody has answered yet, and who each is with
    asgard-cli request       what the customer asked for and is not done
    asgard-cli task          the task specs that are open
    asgard-cli add <kind>    a CR skeleton, wired to what the chart declares
    asgard-cli check         the structure; "verify" is the rendered chart

Each of those four reads a file in the customer's repository back to you, and
each takes ` + "`--format json`" + `. **None of them says where the engagement is.** There
is no such command and there was: it derived one position from the earliest
missing CR kind, and an onboarding is not linear - three of the most expensive
decisions in the engagement this was built from were made, built and reversed.
What replaced it is the records themselves, and guidance reached by subject
through "find" or by name through "guide", without arriving anywhere to be
handed it.

Work arrives as a request: one thing the customer wants that the agent cannot do
today. "asgard-cli request add" opens one, and every status the engagement keeps
lives in the customer's repository, never in this tool, so the agent that opens
that repo next can read where the work stands.

With no repository yet, "asgard-cli guide init" says how one begins and
"asgard-cli init" does it: the skeleton, the binding to a workspace and a
pipeline on the platform, and the reference material describing that platform.
Which workspace and which pipeline are the only two facts a repository cannot
supply about itself, so they are the only two it records - and neither is ever
guessed, not even from a list of one.

Run "asgard-cli <command> --help" for details on an individual command.`,
		Version: version.Get().String(),

		// On failure print just the error, not a full page of usage; main is the
		// single place that prints it and picks the exit code.
		SilenceUsage:  true,
		SilenceErrors: true,

		// With no subcommand, show help rather than silently succeeding.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

	// --template-dir is persistent because a prompt is read by `guide`, `find`
	// and `audit-material` alike, and an override that applied to only one of
	// them would make the three disagree about what the material says.
	var templateDir string
	cmd.PersistentFlags().StringVar(&templateDir, "template-dir", "",
		"read stage prompts from this directory instead of the embedded copies, per file; for iterating on prompt text")
	cmd.PersistentPreRunE = func(c *cobra.Command, _ []string) error {
		if templateDir == "" {
			stage.SetOverrideDir("")
			return nil
		}
		info, err := os.Stat(templateDir)
		if err != nil {
			return fmt.Errorf("--template-dir %s: %w", templateDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("--template-dir %s is not a directory", templateDir)
		}
		stage.SetOverrideDir(templateDir)
		// Say it on every run. A prompt that is not the released one is the
		// first thing to suspect when guidance looks wrong, and an override
		// nobody can see is worse than no override.
		fmt.Fprintf(c.ErrOrStderr(), "reading stage prompts from %s where they exist, embedded otherwise\n", templateDir)
		return nil
	}

	// Said at the end of whatever was run, because it is the only moment an
	// agent is looking. Its reference material going stale has no symptom - the
	// CR it writes is wrong in a way that reads fine - so the report has to
	// ride on something already being run rather than wait to be asked for.
	//
	// PersistentPostRun rather than a call in each command: every command that
	// reaches the platform is a chance to say it, and one that has to be added
	// per command is one that gets forgotten on the command somebody adds next.
	cmd.PersistentPostRun = func(c *cobra.Command, _ []string) {
		warnIfBehind(c)
	}

	// The command tree, grouped. **The grouping is the source, not a rendering
	// of it**: a command is added by naming the group it belongs to, so adding
	// one without deciding where it goes does not compile. A flat list of
	// thirty commands is what this help used to be, and it is what made four
	// commands that compose look like four commands that compete.
	// Insertion order, not alphabetical. Within a group the first entry is
	// where somebody starts - `gate` before the three checks it composes,
	// `login` before what needs a session - and alphabetical ordering put
	// `check` and `doctor` in front of `gate`, which is the opposite of what
	// the grouping was for.
	cobra.EnableCommandSorting = false

	cmd.AddGroup(
		&cobra.Group{ID: groupAsk, Title: "Ask - what the platform is, and how a shape is built:"},
		&cobra.Group{ID: groupBuild, Title: "Build - write the repository and the CRs in it:"},
		&cobra.Group{ID: groupCheck, Title: "Check - everything this machine can check:"},
		&cobra.Group{ID: groupDeploy, Title: "Deploy - the platform, and what it knows:"},
	)

	addTo(cmd, groupAsk,
		newFindCmd(),
		newWikiCmd(),
		newUsecaseCmd(),
		newBriefCmd(),
		newGuideCmd(),
		newSizeCmd(),
		newReadingCmd(),
		newIssueCmd(),
	)
	addTo(cmd, groupBuild,
		newInitCmd(),
		newScaffoldCmd(),
		newProjectCmd(),
		newAddCmd(),
		newQuestionCmd(),
		newRequestCmd(),
		newTaskCmd(),
		newDecisionCmd(),
		newReferenceCmd(),
	)
	addTo(cmd, groupCheck,
		newGateCmd(),
		newCheckCmd(),
		newRenderCmd(),
		newVerifyCmd(),
		newDoctorCmd(),
	)
	addTo(cmd, groupDeploy,
		newLoginCmd(),
		newLogoutCmd(),
		newWhoamiCmd(),
		newProfileCmd(),
		newWorkspaceCmd(),
		newPipelineCmd(),
		newSkillCmd(),
	)

	// Ungrouped, and they belong there. `version` answers a question about the
	// binary rather than about an engagement, and `audit-material` is hidden -
	// its reader edits this material, and the help belongs to whoever is
	// onboarding a customer.
	cmd.AddCommand(
		newVersionCmd(),
		newAuditCmd(),
	)

	// What this build answers to, handed to `check` so it can report a command
	// name in a customer repository that no longer exists. The list has to come
	// from the tree rather than from a constant: a constant is a second copy
	// that goes stale exactly when a command is renamed, which is the failure
	// this exists to catch.
	check.SetKnownCommands(commandNames(cmd))
	check.SetReplacements(replacements)

	return cmd
}

// The four groups the top-level help is organised into.
//
// They are the four questions somebody arrives with, in the order they arrive:
// what is this platform, how do I write the repository, is what I wrote sound,
// and get it deployed. A command that fits none of them is a command whose
// place in the tool has not been decided.
const (
	groupAsk    = "ask"
	groupBuild  = "build"
	groupCheck  = "check"
	groupDeploy = "deploy"
)

// addTo registers commands into a group.
//
// It exists so the grouping cannot be forgotten: cobra panics on a GroupID
// naming a group the parent does not have, and a command added through
// AddCommand instead of this lands in "Additional Commands" where the next
// reader will see it and ask why.
func addTo(parent *cobra.Command, group string, children ...*cobra.Command) {
	for _, c := range children {
		c.GroupID = group
		parent.AddCommand(c)
	}
}

// replacements says what to type instead of a command this tool removed.
//
// **Add a row here when you rename or remove one.** Every repository already
// scaffolded carries the old name in files `scaffold` will never overwrite, and
// `asgard-cli check` reports them - but a reader told only that a command is
// gone has to find out what replaced it, and the first one to hit this had to
// ask a maintainer. That is the answer living in a conversation instead of in
// the binary.
var replacements = map[string]string{
	"project shape": "Gone with `.asgard-config.json`. It recorded what a chart was being built to be, which is a claim about intent that nothing can verify - " +
		"say it in the chart, next to whatever makes the project unusual, where the next reader is already looking. `asgard-cli size` still lists the shapes",
	"next": "It derived a position from the earliest missing CR kind and there is no replacement for that, deliberately - an onboarding is not linear. " +
		"Where it meant \"what is still open\", `asgard-cli question`; where it meant \"what does each chart declare\", `asgard-cli project`; " +
		"`next --stage <name>` is `asgard-cli guide <name>`, and `next --list` is `asgard-cli guide` with no argument",
	"status": "Where it meant \"what is still open\", `asgard-cli question`, `asgard-cli request` and `asgard-cli task`; " +
		"where it meant \"what does each chart declare and still lack\", `asgard-cli project`. It also named the guidance the " +
		"repository's shape made relevant, and nothing replaces that: read one with `asgard-cli guide <name>` or reach it by subject with `asgard-cli find`",
}

// commandNames returns every name and alias in the tree, one level deep.
//
// One level is deliberate. `asgard-cli request add` names the command
// `request`, and whether `add` is one of its subcommands is a different
// question - a wrong subcommand is a typo, a wrong command is a rename nobody
// was told about.
func commandNames(root *cobra.Command) []string {
	var out []string
	for _, c := range root.Commands() {
		out = append(out, c.Name())
		out = append(out, c.Aliases...)
	}
	return out
}

// recallHere notes a page as opened, when the command was run inside an
// engagement. Both `wiki` and `usecase` work with no repository at all - that
// is deliberate, they are reference material - so this finds one if there is
// one and does nothing if there is not.
func recallHere(kind, name string) {
	root := repo.Root(".")
	if root == "" {
		return
	}
	work.Recall(root, kind, name)
}
