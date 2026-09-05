package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/binding"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
)

// loadRepo finds the customer repository from the working directory.
//
// **The declaration is what makes a directory one.** It used to be
// `.asgard-config.json`, this tool's own scaffold record - so a repository that
// had everything the platform needs and none of this tool's bookkeeping was not
// a repository as far as every command was concerned. The declaration is the
// file the platform reads on every run, which is the honest test.
func loadRepo() (root string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}
	declPath, _, err := binding.Locate(dir)
	if err != nil {
		return "", err
	}
	if declPath == "" {
		return "", fmt.Errorf("no %s at or above %s, so this is not a repository this tool deploys from.\n"+
			"`asgard-cli scaffold` writes one, along with the rest of the skeleton",
			pipelineconfig.FileName, dir)
	}
	return filepath.Dir(declPath), nil
}

// today is the date the records are stamped with. It is a variable so a test can
// pin it: a golden file that changes at midnight is not a test.
var today = func() string { return time.Now().Format("2006-01-02") }

// loadState reads the repository and every record it keeps.
//
// Four commands list what the repo contains, and they used to be one command
// that listed all of it and then said which guidance the shape of it made
// relevant. The listing survived that; the inference did not. Each command now
// reads the whole state and prints only its own part, because the parts are
// separate claims and an agent asking what is unanswered should not have to
// read past what a chart declares to find out.
func loadState() (stage.State, error) {
	root, err := loadRepo()
	if err != nil {
		return stage.State{}, err
	}
	return stage.Inspect(root)
}

// errNotInRepo is the one wording for "you are not in a repository this tool
// deploys from", so the answer reads the same whichever command asked.
func errNotInRepo() error {
	dir, _ := os.Getwd()
	return fmt.Errorf("no %s at or above %s, so this is not a repository this tool deploys from.\n"+
		"`asgard-cli scaffold` writes one, along with the rest of the skeleton",
		pipelineconfig.FileName, dir)
}

// workspaceForTemplates resolves the customer's name for the scaffolded
// documents, from the platform.
//
// **Nothing stores it.** The name belongs to the workspace on the platform, and
// a copy on disk is a copy that is wrong the day somebody renames it. Offline,
// or before a workspace is bound, the id or a placeholder is written instead -
// which is at least true, and re-running `scaffold` fills it in.
//
// NeedWorkspace is set: without it the resolution stops at the session and
// leaves the id empty, so every scaffolded document said "<workspace>" whether
// or not the checkout was bound. Every way this can fail - not logged in,
// nothing recorded, no network - is a placeholder rather than an error, because
// none of them is a reason to refuse to write a skeleton.
func workspaceForTemplates(cmd *cobra.Command) (repo.Workspace, error) {
	pc, err := resolveContext(cmd, contextOptions{NeedWorkspace: true})
	if err != nil || pc.Workspace == "" {
		return repo.Workspace{Name: "<workspace>"}, nil
	}
	return workspaceNamed(cmd, pc.Session, pc.Workspace), nil
}

// workspaceNamed turns a workspace id into the display name the templates
// print, falling back to the id - which is at least true - when the platform
// cannot be asked or does not know it.
func workspaceNamed(cmd *cobra.Command, session *auth.Session, id string) repo.Workspace {
	workspaces, err := platform.New(session, "").ListWorkspaces(cmd.Context())
	if err != nil {
		return repo.Workspace{Name: id}
	}
	for _, w := range workspaces {
		if w.ID == id {
			return repo.Workspace{Name: w.Name}
		}
	}
	return repo.Workspace{Name: id}
}
