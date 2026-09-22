#!/bin/sh
# One-command install for macOS.
#
#   curl -fsSL https://raw.githubusercontent.com/asgard-ai-partners/asgard-fde-cli/main/install.sh | sh
#
# **It verifies what it downloaded and it warms the first run.** Neither is
# decoration. The binaries are ad-hoc signed and not notarized, so macOS scans
# the first execution of a newly written one - which takes anything from no time
# to minutes, and has been seen to kill it outright. Getting that over with here
# means the customer's first real use is not the one that stalls, and a failure
# happens while they are still looking at an installer rather than in a meeting.
set -eu

repo=asgard-ai-partners/asgard-fde-cli
base="https://github.com/$repo/releases/latest/download"
asset=asgard-cli_darwin_all.tar.gz      # one binary, both architectures

[ "$(uname -s)" = Darwin ] || {
	echo "This installer is for macOS. On Linux take the .deb, .rpm or the linux tarball:" >&2
	echo "  https://github.com/$repo/releases/latest" >&2
	exit 1
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

printf 'Downloading the latest asgard-cli...\n'
curl -fsSL --retry 3 -o "$tmp/$asset" "$base/$asset"
curl -fsSL --retry 3 -o "$tmp/checksums.txt" "$base/checksums.txt"

# **Matched by hash rather than by name.** The release carries the same bytes
# twice - once with a version in the filename and once without, so that the URL
# above needs no version - and only the versioned name is in checksums.txt.
got=$(shasum -a 256 "$tmp/$asset" | cut -d' ' -f1)
grep -q "^$got  " "$tmp/checksums.txt" || {
	echo "Checksum mismatch: $got is not in this release's checksums.txt." >&2
	echo "Nothing was installed. Report this - it should not happen." >&2
	exit 1
}
printf '  checksum ok\n'

tar xzf "$tmp/$asset" -C "$tmp" asgard-cli
# curl does not mark a download the way a browser does, but a file that reached
# this machine some other way might be, and the mark is what gets it killed.
xattr -d com.apple.quarantine "$tmp/asgard-cli" 2>/dev/null || true

dir=/usr/local/bin
if [ -w "$dir" ]; then
	install -m 0755 "$tmp/asgard-cli" "$dir/asgard-cli"
elif sudo -n true 2>/dev/null || [ -t 1 ]; then
	printf 'Installing to %s (needs sudo)\n' "$dir"
	sudo install -d -m 0755 "$dir"
	sudo install -m 0755 "$tmp/asgard-cli" "$dir/asgard-cli"
else
	# **No terminal to ask on**, which is what happens inside some CI and some
	# terminals when this is piped. Installing somewhere the user owns beats
	# failing, as long as it says so.
	dir="$HOME/.local/bin"
	install -d -m 0755 "$dir"
	install -m 0755 "$tmp/asgard-cli" "$dir/asgard-cli"
	printf '\nInstalled to %s, which may not be on your PATH. Add it:\n' "$dir"
	printf '  echo '\''export PATH="$HOME/.local/bin:$PATH"'\'' >> ~/.zshrc\n'
fi

printf 'Checking it runs (macOS scans a new binary once, which can take a minute)...\n'
"$dir/asgard-cli" version

printf '\n'
"$dir/asgard-cli" doctor || true
