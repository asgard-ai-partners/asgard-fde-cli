package cli

import (
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
)

// Every command that files something carries the same table, so an agent that
// reads only one command's help still learns the other two places.
func TestFilingCommandsCarryTheTrackersTable(t *testing.T) {
	t.Setenv(auth.EnvSandboxMode, "")
	for _, args := range [][]string{{"workbench"}, {"issue-report"}} {
		out, _, err := runCLI(t, "", append(args, "--help")...)
		if err != nil || !strings.Contains(out, "WHERE A THING IS FILED") {
			t.Errorf("%v --help lacks the table: %v", args, err)
		}
	}
	// The rest point at it rather than repeat it.
	for _, args := range [][]string{{"workbench", "create"}, {"question"}, {"request"}, {"task"}, {}} {
		out, _, err := runCLI(t, "", append(args, "--help")...)
		if err != nil || !strings.Contains(out, "asgard-cli workbench --help") {
			t.Errorf("%v --help does not point at the table: %v", args, err)
		}
	}
}

func TestIssueReportFilesNothingFromTheSandbox(t *testing.T) {
	sandboxEnv(t, "https://platform-api.example.test")
	out, _, err := runCLI(t, "", "issue-report")
	if err != nil || !strings.Contains(out, "Nothing is filed upstream") || strings.Contains(out, "gh issue create") {
		t.Errorf("issue-report in the sandbox: %q %v", out, err)
	}
	out, _, err = runCLI(t, "", "issue-report", "--new")
	if err == nil || out != "" {
		t.Errorf("issue-report --new in the sandbox wrote %q, err %v", out, err)
	}

	t.Setenv(auth.EnvSandboxMode, "")
	if out, _, err := runCLI(t, "", "issue-report"); err != nil || !strings.Contains(out, "gh issue create") {
		t.Errorf("issue-report outside the sandbox: %q %v", out, err)
	}
}
