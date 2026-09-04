package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
)

// loadRepo finds the customer repository from the working directory and reads
// its config. Every command that writes into the repo starts here, so that the
// error for "you are not in one" is worded the same way each time.
func loadRepo() (root string, cfg *config.Config, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", nil, fmt.Errorf("get current directory: %w", err)
	}
	path, err := config.Find(dir)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return "", nil, fmt.Errorf("no %s found; run `asgard-cli init` first", config.FileName)
		}
		return "", nil, err
	}
	cfg, err = config.Load(path)
	if err != nil {
		return "", nil, err
	}
	return filepath.Dir(path), cfg, nil
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
	root, cfg, err := loadRepo()
	if err != nil {
		return stage.State{}, err
	}
	return stage.Inspect(root, cfg)
}
