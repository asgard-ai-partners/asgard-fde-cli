package tool

import (
	"strings"
	"testing"
)

// TestLinuxHintsNameARepositoryWhereThereIsNoPackage is the finding that shapes
// this whole file: neither helm nor kubectl is in the Debian or Ubuntu default
// repositories, so an `apt install helm` would fail on an unmet dependency. A
// hint that does not work is worse than no hint.
func TestLinuxHintsNameARepositoryWhereThereIsNoPackage(t *testing.T) {
	for _, tl := range []Tool{Helm, Kubectl} {
		hint := tl.linuxHint("debian")
		if hint == "" {
			t.Fatalf("%s has no Debian hint", tl.Name)
		}
		if strings.Contains(hint, "apt install "+tl.Name) {
			t.Errorf("%s Debian hint suggests an apt install that cannot work:\n%s", tl.Name, hint)
		}
		// Either the direct download or the repository to add first.
		if !strings.Contains(hint, "curl") {
			t.Errorf("%s Debian hint should give a way that works without a repo:\n%s", tl.Name, hint)
		}
	}
}

func TestLinuxHintsPerFamily(t *testing.T) {
	tests := []struct {
		family, tool, want string
	}{
		{"alpine", Helm.Name, "apk add helm"},
		{"alpine", Kubectl.Name, "apk add kubectl"},
		{"arch", Helm.Name, "pacman -S helm"},
		{"rhel", Kubectl.Name, "kubernetes-client"},
		{"rhel", Python.Name, "dnf install python3"},
		{"debian", Python.Name, "apt install python3"},
	}
	for _, tt := range tests {
		t.Run(tt.family+"/"+tt.tool, func(t *testing.T) {
			var tl Tool
			for _, candidate := range All {
				if candidate.Name == tt.tool {
					tl = candidate
				}
			}
			if got := tl.linuxHint(tt.family); !strings.Contains(got, tt.want) {
				t.Errorf("linuxHint(%q) = %q, want it to mention %q", tt.family, got, tt.want)
			}
		})
	}

	// An unrecognised distribution gets no advice rather than wrong advice.
	if got := Helm.linuxHint("plan9"); got != "" {
		t.Errorf("unknown family should produce no hint, got %q", got)
	}
}

// TestErrMissingCarriesTheInstallLine is the point of the package: the bare
// exec.LookPath error tells an FDE nothing they can act on.
func TestErrMissingCarriesTheInstallLine(t *testing.T) {
	err := &ErrMissing{Tool: Helm}
	msg := err.Error()

	if !strings.Contains(msg, "helm is not on PATH") {
		t.Errorf("error = %q", msg)
	}
	if !strings.Contains(msg, Helm.Purpose) {
		t.Errorf("error should say what it is needed for: %q", msg)
	}
	// On any platform this test runs on, there is a hint to give.
	if !strings.Contains(msg, "Install it with") {
		t.Errorf("error should carry an install line: %q", msg)
	}
}

func TestPathReportsMissingWithTheRightError(t *testing.T) {
	absent := Tool{Name: "a-command-that-does-not-exist", Purpose: "nothing"}
	if _, err := absent.Path(); err == nil {
		t.Fatal("a command that does not exist should not resolve")
	} else if _, ok := err.(*ErrMissing); !ok {
		t.Errorf("error is %T, want *ErrMissing so callers can distinguish it", err)
	}
}

func TestMajor(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"v3.18.4+gd80839c", 3},
		{"v4.2.4", 4},
		{"3.18.4", 3},
		{" v12.0.0 ", 12},
		{"", 0},
		{"unknown", 0},
	}
	for _, tt := range tests {
		if got := Major(tt.in); got != tt.want {
			t.Errorf("Major(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestKubectlVersionReadsStructuredOutput(t *testing.T) {
	// kubectl has had no --short since 1.28, so structured output is the only
	// stable form; the first line of the YAML form is just "clientVersion:".
	got := kubectlVersion(`{"clientVersion":{"gitVersion":"v1.35.1"}}`)
	if got != "v1.35.1" {
		t.Errorf("kubectlVersion = %q, want v1.35.1", got)
	}
	// Anything unparseable falls back rather than reporting nothing.
	if got := kubectlVersion("something else\nsecond line"); got != "something else" {
		t.Errorf("fallback = %q", got)
	}
}
