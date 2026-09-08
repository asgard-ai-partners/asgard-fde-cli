package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/brief"
)

func newBriefCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brief [activity]",
		Short: "What you are about to do, and where it goes wrong",
		Long: `The known ways to get one activity wrong, before doing it.

` + "`asgard-cli question`" + ` and its three neighbours answer "what does this
repository record", from files. This answers "the thing I am about to do - where
will I get it wrong", which is a different question with more expensive answers:

  - **the riskiest activity leaves no trace.** Talking to a customer changes no
    file, so nothing derived from repository state can prepare anybody for it
  - **meetings happen at every stage.** A briefing reachable only from the
    interview guidance is unreachable to an engagement halfway through a
    chart with a meeting tomorrow
  - **what you need to know precedes where it is filed.** The material is
    arranged for reading at the point of use, which is right, and the interview
    is where that breaks: half of what gets said out loud is settled in stages
    nobody has reached yet

Nothing here is new knowledge. Every stage already names the decision that gets
answered wrong, every extract carries an Unchecked block, and platform-unknowns
is a page of nothing else. What this adds is a way in addressed by intent rather
than by position.

    asgard-cli brief                     the activities
    asgard-cli brief customer-meeting    before any customer conversation
    asgard-cli brief write-chart         before authoring CRs
    asgard-cli brief handover            before saying it is live`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if len(args) == 0 {
				fmt.Fprintf(out, "Before doing one of these, read what it gets wrong:\n\n")
				for _, a := range brief.Activities {
					fmt.Fprintf(out, "  %-18s %s\n", a.Name, a.When)
					fmt.Fprintf(out, "  %-18s %d known way(s) to get it wrong\n\n", "", len(a.Items))
				}
				fmt.Fprintf(out, "One of them: `asgard-cli brief <activity>`.\n\n"+
					"An activity earns a place here when somebody has actually got it wrong,\n"+
					"not when it is important.\n")
				return nil
			}

			a, ok := brief.Find(args[0])
			if !ok {
				return fmt.Errorf("no briefing for %q; one of: %s",
					args[0], strings.Join(brief.Names(), ", "))
			}
			// **This is the one the log most needs.** Every entry here exists
			// because somebody got it wrong, so whether a predecessor was
			// briefed before talking to a customer changes what the next
			// person does - and the log could not say, because these reads
			// were never recorded.
			recallHere("brief", a.Name)
			a.Render(out)
			return nil
		},
	}
	return cmd
}
