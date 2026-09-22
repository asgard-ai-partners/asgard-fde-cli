package cli

import "runtime"

// installCommand is the one line that replaces this binary with the newest
// release.
//
// **It is here rather than written out at each site** so that the command the
// gate prints, the one `version --check` prints, and the one the README leads
// with cannot drift into three different answers - which is the failure this
// repository removes everywhere else.
func installCommand() string {
	if runtime.GOOS == "darwin" {
		return "curl -fsSL https://raw.githubusercontent.com/asgard-ai-partners/asgard-fde-cli/main/install.sh | sh"
	}
	return "see https://github.com/asgard-ai-partners/asgard-fde-cli/releases/latest"
}
