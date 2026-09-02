package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/size"
)

func newSizeCmd() *cobra.Command {
	var in size.Inputs

	cmd := &cobra.Command{
		Use:   "size [shape]",
		Short: "What one capability is made of, before it is written",
		Long: `Estimate a capability's shape and size from what the interview established.

"How many agents, how many projects" is the first question a proposal is asked
and the basis of a quote, and until this existed the answer was worked out by
hand, differently each time. The stage prompts say how the split is decided and
the extracts say what one shape contains; nothing added them up.

    asgard-cli size                       the shapes, and what each costs empty
    asgard-cli size flow-agent-single --databases 2 --queries 4 --writes 1 --knowledge 1

**One capability at a time.** A request covering two audiences is two requests
and two estimates - they share no entry point and no read path, so adding their
CRs together describes nothing that will be built.

The counts come from deployments in production rather than from reasoning, which
matters most where the intuitive answer is wrong: **the flow-agent shapes
contain no Agent CR at all.**

Two outputs, and the second is not decoration. The CR table is for the estimate;
the plain reading is what may go in front of the customer, because CR kinds are
the first thing a proposal deck forbids. Handing over only the table means
somebody translates it under time pressure, and reaches for the word in front of
them.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if len(args) == 0 {
				fmt.Fprintf(out, "Shapes an engagement chooses between, by who is on the other end:\n\n")
				for _, s := range size.Shapes {
					fmt.Fprintf(out, "  %s\n", s)
					fmt.Fprintf(out, "  %-24s base: %s\n", "", s.BaseSummary())
					fmt.Fprintf(out, "  %-24s seen in %s\n\n", "", s.SeenIn)
				}
				fmt.Fprintf(out, "One of them: `asgard-cli size <shape> --databases N --queries N ...`\n"+
					"The audience decides which - `asgard-cli next --stage requirements`, question 2.\n")
				return nil
			}

			shape, ok := size.Find(args[0])
			if !ok {
				return fmt.Errorf("no shape named %q; one of: %s", args[0], strings.Join(size.Names(), ", "))
			}

			e := size.Of(shape, in)

			fmt.Fprintf(out, "%s\n%s\n\n", shape.Name, shape.Audience)
			if shape.Note != "" {
				fmt.Fprintf(out, "%s\n\n", shape.Note)
			}

			fmt.Fprintf(out, "One project. A second audience is a second project and a second estimate.\n\n")
			fmt.Fprintf(out, "CRs, for the estimate - not for a customer's screen:\n\n")
			for _, k := range e.Sorted() {
				fmt.Fprintf(out, "  %-18s %d\n", k, e.CRs[k])
			}
			fmt.Fprintf(out, "  %-18s %d\n", "TOTAL", e.Total)
			if e.CRs["Agent"] == 0 {
				fmt.Fprintf(out, "\n  Agents: 0. That is the shape, not an omission.\n")
			}

			fmt.Fprintf(out, "\nThe same thing, said the way a customer can check:\n\n")
			for _, line := range e.Plain() {
				fmt.Fprintf(out, "  - %s\n", line)
			}

			if len(e.Warnings) > 0 {
				fmt.Fprintf(out, "\nWhat makes this number conditional:\n\n")
				for _, w := range e.Warnings {
					fmt.Fprintf(out, "  %s\n\n", strings.ReplaceAll(w, "\n", "\n  "))
				}
			}

			fmt.Fprintf(out, "A number stated where an open question could double it is a guess with a\n"+
				"decimal point. `asgard-cli next` prints what is still open.\n")
			return nil
		},
	}

	f := cmd.Flags()
	f.IntVar(&in.Databases, "databases", 0, "systems read through a semantic layer")
	f.IntVar(&in.APIs, "apis", 0, "systems reached over HTTP")
	f.IntVar(&in.Queries, "queries", 0, "fixed query tools, where a layer is the wrong surface")
	f.IntVar(&in.Writes, "writes", 0, "actions with a side effect, each gated")
	f.IntVar(&in.Knowledge, "knowledge", 0, "document sources - manuals, FAQs, a site")
	f.IntVar(&in.Agents, "agents", 0, "specialisms, for the shapes that carry Agent CRs")
	f.IntVar(&in.Consoles, "consoles", 0, "systems with no database and no API")
	f.IntVar(&in.Schedules, "schedules", 0, "scheduled runs")
	return cmd
}
