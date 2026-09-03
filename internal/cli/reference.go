package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newReferenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reference",
		Short: "File the customer's own material, with where it came from",
		Long: `File the customer's own material, with where it came from.

` + "`references/`" + ` is background for humans and spec-writing agents. It is not what
the running agent reads - domain knowledge the agent needs at run time belongs in
a skill, because a skill is synced into the platform and this directory is not.

Filing a document is a step every engagement takes and none has done the same
way: each invented its own provenance table, and one invented a directory name
that then read like a convention.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newReferenceAddCmd())
	return cmd
}

func newReferenceAddCmd() *cobra.Command {
	var ref work.Reference

	cmd := &cobra.Command{
		Use:   "add <file>",
		Short: "Copy a document into references/ and record its provenance",
		Long: `Copy a document into ` + "`references/`" + ` and record its provenance.

    asgard-cli reference add ~/Downloads/rma.xlsx \
      --what "the RMA status codes, as their support team uses them" \
      --from "their support team" --dated 2024-11 --as erp/rma-status.xlsx

**The document is copied byte-identical and never rewritten.** The provenance
goes in ` + "`" + work.ReferenceIndex + "`" + ` instead of a header pasted into their file, so that
when they send a second version you can diff it against the filed one. For the
same reason, filing over an existing name is refused: a customer's second version
is a different document, and the two together are how anyone sees what changed.

**--dated is the document's own date, not today.** It is the one that decides
whether the material is stale. A document carrying no date is worth recording as
carrying none - material a customer wrote for their own staff describes the
system they believe they have, and a stale page reads exactly like a current one.

**--what is the sentence the file name cannot carry.** "the RMA status codes, as
their support team uses them" is worth more than "rma.xlsx", and it is what a
reader six weeks later uses to decide whether to open it.

Every row starts ` + "`verified: no`" + ` and is meant to be edited by hand once you have
held a claim against the running system. A row marked no is worth more than a
plausible one, because the reader knows which to trust.

Filing material is not reading it. ` + "`asgard-cli check`" + ` says so: material in
` + "`references/`" + ` with no question and no request recorded against it is a warning,
because the interview is what turns a document into a requirement.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, _, err := loadRepo()
			if err != nil {
				return err
			}
			ref.Filed = today()

			out := cmd.OutOrStdout()
			filed, err := work.AddReference(root, args[0], ref, out)
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Filed %s\n", filepath.ToSlash(filepath.Join(work.ReferenceDir, filed.Path)))
			fmt.Fprintf(out, "Recorded in %s\n", filepath.ToSlash(work.ReferenceIndex))

			var missing []string
			if filed.What == "" {
				missing = append(missing, "--what")
			}
			if filed.From == "" {
				missing = append(missing, "--from")
			}
			if filed.Dated == "" {
				missing = append(missing, "--dated")
			}
			if len(missing) > 0 {
				fmt.Fprintf(out, "\nThe row is short of %v. Fill them in now rather than later:\nwhoever handed you this document is the only person who knows, and they\nstop being available at exactly the point somebody needs to know.\n", missing)
			}

			fmt.Fprintf(out, `
Filing is not reading. What turns this into a requirement is the interview:

    asgard-cli question add "<what it left open>" --ask "<who>"
    asgard-cli next --stage requirements
`)
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&ref.What, "what", "", "what the document is, in your words - not its file name")
	f.StringVar(&ref.From, "from", "", "who supplied it; a role outlives a person")
	f.StringVar(&ref.Dated, "dated", "", "the document's own date (YYYY-MM-DD or YYYY-MM), not today")
	f.StringVar(&ref.Path, "as", "", "name to file it under, relative to references/ (defaults to the file's own name)")
	return cmd
}
