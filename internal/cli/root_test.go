package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	var out bytes.Buffer

	root := NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got := out.String(); !strings.HasPrefix(got, "asgard-cli ") {
		t.Errorf("version output = %q, want prefix %q", got, "asgard-cli ")
	}
}

func TestVersionCommandJSON(t *testing.T) {
	var out bytes.Buffer

	root := NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var got struct {
		Version  string `json:"version"`
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", out.String(), err)
	}
	if got.Version == "" || got.Platform == "" {
		t.Errorf("incomplete version info: %+v", got)
	}
}

func TestUnknownCommandFails(t *testing.T) {
	var out bytes.Buffer

	root := NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"definitely-not-a-command"})

	if err := root.Execute(); err == nil {
		t.Error("expected error for unknown command, got nil")
	}
}
