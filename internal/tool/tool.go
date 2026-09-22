// Package tool resolves the external commands asgard-cli shells out to, and
// says how to install one when it is missing.
//
// The point is the error message. `exec: "helm": executable file not found in
// $PATH` tells an FDE nothing they can act on, and on Windows the failure is
// worse than that: `python3` resolves to an App Execution Alias that opens the
// Microsoft Store instead of running anything, so the command appears to do
// something and does not. Every lookup here produces the install line for the
// machine it is running on.
package tool

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Tool is an external command.
type Tool struct {
	// Name as it appears on PATH. Never suffixed with .exe: exec.LookPath reads
	// PATHEXT on Windows, and hardcoding the suffix breaks the shims that Git
	// Bash and scoop install.
	Name string

	// Alternatives are tried in order when Name is not found. Windows Python is
	// the reason this exists: python.org's installer creates python.exe and no
	// python3.exe.
	Alternatives []string

	// Purpose is what asgard-cli needs it for, shown by doctor.
	//
	// **A bare infinitive, because it is read in two positions.** `doctor`
	// prints it as a column, where any form reads; `NotFoundError` puts it
	// after "needs it to", where anything else does not. Two of the three were
	// noun phrases and rendered as "needs it to reading a cluster by hand" and
	// "needs it to the db-query skill" - correct in the column nobody
	// complained about, wrong in the sentence a customer meets when the tool
	// is missing.
	Purpose string

	// Optional marks a tool the acceptance gate does not need.
	Optional bool

	// VersionArgs prints a version and exits zero.
	VersionArgs []string

	// version extracts the version from that output. It defaults to the first
	// line, which is wrong for kubectl: it has no --short since 1.28, so the
	// only stable form is structured output.
	version func(string) string

	// brew, winget and scoop package names, where they differ from Name.
	brew, winget, scoop string
}

// The tools asgard-cli uses, in the order doctor reports them.
var (
	Helm = Tool{
		Name:        "helm",
		Purpose:     "render a chart (asgard-cli render) and lint it",
		VersionArgs: []string{"version", "--short"},
		brew:        "helm",
		winget:      "Helm.Helm",
		scoop:       "helm",
	}

	Kubectl = Tool{
		Name: "kubectl",
		// **No check in this tool needs it.** The server-side dry run and the
		// unknown-field check moved to the platform's plan when the Pipeline
		// landed, and no client is given cluster credentials - so kubectl is
		// something an FDE may want by hand and the gate never asks for.
		Purpose:     "read a cluster by hand; no check here needs it",
		Optional:    true,
		VersionArgs: []string{"version", "--client", "--output=json"},
		version:     kubectlVersion,
		brew:        "kubernetes-cli",
		winget:      "Kubernetes.kubectl",
		scoop:       "kubectl",
	}

	Python = Tool{
		Name:         "python3",
		Alternatives: []string{"python"},
		Purpose:      "run the db-query skill, which reads a customer's source systems",
		Optional:     true,
		VersionArgs:  []string{"--version"},
		brew:         "python",
		winget:       "Python.Python.3.13",
		scoop:        "python",
	}
)

// All lists every tool doctor reports on.
var All = []Tool{Helm, Kubectl, Python}

// ErrMissing reports that a tool is not on PATH, and carries the install line
// for this machine.
type ErrMissing struct {
	Tool Tool
}

func (e *ErrMissing) Error() string {
	hint := e.Tool.InstallHint()
	msg := fmt.Sprintf("%s is not on PATH, and %s needs it to %s",
		e.Tool.Name, "asgard-cli", e.Tool.Purpose)
	if hint == "" {
		return msg
	}
	return msg + "\n\nInstall it with:\n\n    " + strings.ReplaceAll(hint, "\n", "\n    ")
}

// Path resolves the tool on PATH.
func (t Tool) Path() (string, error) {
	for _, name := range append([]string{t.Name}, t.Alternatives...) {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		// On Windows a name can resolve to an App Execution Alias: a zero-byte
		// stub that opens the Microsoft Store. It is on PATH, it is executable,
		// and it runs nothing. Reject it here rather than let a caller wait on
		// a store window.
		if runtime.GOOS == "windows" && isStoreStub(path) {
			continue
		}
		return path, nil
	}
	return "", &ErrMissing{Tool: t}
}

// isStoreStub reports whether path is a Windows App Execution Alias. The alias
// is a reparse point that reads as a zero-length file.
func isStoreStub(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() == 0
}

// Command builds a command for the tool, resolved through Path so a missing
// tool is reported before anything runs.
//
// It never goes through a shell: there is no sh on Windows, and passing
// arguments as a slice avoids every quoting question.
func (t Tool) Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	path, err := t.Path()
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, path, args...), nil
}

// Version returns the tool's reported version, trimmed to one line.
func (t Tool) Version(ctx context.Context) (string, error) {
	cmd, err := t.Command(ctx, t.VersionArgs...)
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", t.Name, strings.Join(t.VersionArgs, " "), err)
	}
	if t.version != nil {
		return t.version(string(out)), nil
	}
	return firstLine(string(out)), nil
}

// kubectlVersion pulls the client version out of `kubectl version -o json`.
func kubectlVersion(out string) string {
	var parsed struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil || parsed.ClientVersion.GitVersion == "" {
		return firstLine(out)
	}
	return parsed.ClientVersion.GitVersion
}

// Major returns the leading major version number of a reported version string,
// or 0 when there is not one to read.
func Major(version string) int {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	digits := 0
	for digits < len(version) && version[digits] >= '0' && version[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0
	}
	n, err := strconv.Atoi(version[:digits])
	if err != nil {
		return 0
	}
	return n
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// InstallHint is the command that installs this tool on the current machine.
//
// It is per-platform because the answer genuinely differs, and on Linux it is
// per-distribution: neither helm nor kubectl is in the Debian or Ubuntu default
// repositories, so the honest answer there is a repository to add first, not an
// apt install that will fail.
func (t Tool) InstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		if t.brew == "" {
			return ""
		}
		return "brew install " + t.brew

	case "windows":
		var lines []string
		if t.winget != "" {
			lines = append(lines, "winget install -e --id "+t.winget)
		}
		if t.scoop != "" {
			lines = append(lines, "scoop install "+t.scoop+"        (if you use scoop instead)")
		}
		return strings.Join(lines, "\n")

	case "linux":
		return t.linuxHint(distro())

	default:
		return ""
	}
}

// linuxHint answers per distribution family. The Debian and RHEL cases name the
// repository because the package is not in the default one, and an install
// command that fails on an unmet dependency is worse than no advice.
func (t Tool) linuxHint(family string) string {
	switch family {
	case "alpine":
		switch t.Name {
		case Helm.Name, Kubectl.Name:
			return "apk add " + t.Name + "        (community repository)"
		case Python.Name:
			return "apk add python3"
		}

	case "arch":
		switch t.Name {
		case Helm.Name, Kubectl.Name:
			return "pacman -S " + t.Name
		case Python.Name:
			return "pacman -S python"
		}

	case "rhel":
		switch t.Name {
		case Helm.Name:
			return "dnf install helm        (Fedora; on RHEL/CentOS use the script below)\n" +
				"curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
		case Kubectl.Name:
			return "dnf install kubernetes-client        (Fedora)\n" +
				"# RHEL/CentOS: add the pkgs.k8s.io yum repository first, see\n" +
				"# https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/"
		case Python.Name:
			return "dnf install python3"
		}

	case "debian":
		switch t.Name {
		case Helm.Name:
			// Not in the Debian or Ubuntu repositories. The upstream script is
			// one line and needs no repository, so it is the honest answer.
			return "curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash\n" +
				"# or add the Helm apt repository: https://helm.sh/docs/intro/install/"
		case Kubectl.Name:
			// Likewise: pkgs.k8s.io has to be added, and the direct download
			// avoids that entirely for a single machine.
			return "curl -fsSLO \"https://dl.k8s.io/release/$(curl -fsSL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl\"\n" +
				"install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl\n" +
				"# or add the pkgs.k8s.io apt repository: https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/"
		case Python.Name:
			return "apt install python3"
		}
	}
	return ""
}

// distro reports the distribution family from /etc/os-release, so the hint is
// the one that works on this machine rather than a list of every possibility.
func distro() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer f.Close()

	fields := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		fields[key] = strings.Trim(value, `"'`)
	}
	// A truncated os-release is not worth an error: the hint degrades to none,
	// which is the same outcome as an unrecognised distribution.
	if scanner.Err() != nil {
		return ""
	}

	// ID_LIKE is what makes derivatives work: Mint reports debian, Rocky
	// reports rhel, and neither needs its own case.
	for id := range strings.FieldsSeq(fields["ID"] + " " + fields["ID_LIKE"]) {
		switch id {
		case "debian", "ubuntu":
			return "debian"
		case "rhel", "fedora", "centos":
			return "rhel"
		case "alpine":
			return "alpine"
		case "arch":
			return "arch"
		}
	}
	return ""
}
