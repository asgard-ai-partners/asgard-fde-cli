package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// The asset a platform takes has to be the name the release actually carries,
// because getting it wrong is a 404 at the moment somebody is upgrading.
//
// **One asset serves both Macs.** The release carries a universal binary, so
// there is nothing for the person updating to know about Intel against Apple
// silicon - and the cost of choosing wrong is an exec format error rather than
// a message.
func TestTheAssetNameIsWhatTheReleaseCarries(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "asgard-cli_darwin_all.tar.gz"},
		{"darwin", "amd64", "asgard-cli_darwin_all.tar.gz"},
		{"linux", "amd64", "asgard-cli_linux_amd64.tar.gz"},
		{"linux", "arm64", "asgard-cli_linux_arm64.tar.gz"},
		{"windows", "amd64", "asgard-cli_windows_amd64.zip"},
	}
	for _, tc := range cases {
		if got := AssetName(tc.goos, tc.goarch); got != tc.want {
			t.Errorf("AssetName(%s, %s) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
}

// **Matched by hash rather than by name.** The release carries the same bytes
// twice - once with a version in the filename and once without, so the download
// URL needs no version - and only the versioned name is in checksums.txt. A
// check that looked the name up would find nothing and would have to either
// fail every update or skip verification.
func TestTheDownloadIsMatchedByHashAndNotByName(t *testing.T) {
	dir := t.TempDir()
	sums := filepath.Join(dir, "checksums.txt")
	body := "" +
		"9f2c" + "  asgard-fde-cli_0.1.2_linux_amd64.tar.gz\n" +
		"deadbeef  asgard-fde-cli_0.1.2_darwin_all.tar.gz\n"
	if err := os.WriteFile(sums, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	// The name in the file is the versioned one and the asset downloaded is
	// the version-less copy, so only the digest can match them.
	if err := verify(sums, "deadbeef"); err != nil {
		t.Errorf("a digest that is in checksums.txt was rejected: %v", err)
	}
	if err := verify(sums, "not-in-there"); err == nil {
		t.Error("a digest that is not in checksums.txt was accepted, so nothing is verified")
	}
}

// Both archive shapes the release publishes give up the binary, and an archive
// that does not carry one is an error rather than an empty file written over
// somebody's working install.
func TestTheBinaryComesOutOfEitherArchiveShape(t *testing.T) {
	dir := t.TempDir()
	const body = "#!/bin/sh\necho hello\n"

	t.Run("tar.gz", func(t *testing.T) {
		archive := filepath.Join(dir, "a.tar.gz")
		writeTarGz(t, archive, map[string]string{"README.md": "x", "asgard-cli": body})

		dest := filepath.Join(dir, "out-tar")
		if err := extract(archive, dest); err != nil {
			t.Fatalf("extract: %v", err)
		}
		if got, _ := os.ReadFile(dest); string(got) != body {
			t.Errorf("extracted %q, want %q", got, body)
		}
	})

	t.Run("zip", func(t *testing.T) {
		archive := filepath.Join(dir, "a.zip")
		writeZip(t, archive, map[string]string{"README.md": "x", "asgard-cli": body})

		dest := filepath.Join(dir, "out-zip")
		if err := extract(archive, dest); err != nil {
			t.Fatalf("extract: %v", err)
		}
		if got, _ := os.ReadFile(dest); string(got) != body {
			t.Errorf("extracted %q, want %q", got, body)
		}
	})

	t.Run("an archive with no binary in it", func(t *testing.T) {
		archive := filepath.Join(dir, "empty.tar.gz")
		writeTarGz(t, archive, map[string]string{"README.md": "x"})

		dest := filepath.Join(dir, "out-empty")
		if err := extract(archive, dest); err == nil {
			t.Fatal("an archive carrying no asgard-cli was accepted")
		}
		if _, err := os.Stat(dest); err == nil {
			t.Error("it wrote a file anyway, which is what would go over a working install")
		}
	})
}

// **Replacing the binary is a rename, and the mode comes from what was there.**
// A binary installed 0755 stays 0755, and one somebody made group-writable on
// purpose keeps that: the mode on the staged file is a floor rather than a
// decision about somebody else's install.
func TestReplaceSwapsTheContentAndKeepsTheModeThatWasThere(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "asgard-cli")
	if err := os.WriteFile(target, []byte("old"), 0o775); err != nil {
		t.Fatal(err)
	}
	// Explicitly, because the umask takes the group bit off a create and the
	// mode this is about is the one on the file rather than the one asked for.
	if err := os.Chmod(target, 0o775); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(dir, ".staged")
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replace(staged, target); err != nil {
		t.Fatalf("replace: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("target holds %q, want %q", got, "new")
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o775 {
		t.Errorf("mode is %v, want %v - the install's own mode was not carried across", info.Mode().Perm(), os.FileMode(0o775))
	}
}

func writeTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

// A digest is what the comparison is made of, so the helper that produces one
// has to agree with the format checksums.txt is written in.
func TestTheDigestIsHexSha256(t *testing.T) {
	sum := sha256.Sum256([]byte("x"))
	want := hex.EncodeToString(sum[:])

	dir := t.TempDir()
	sums := filepath.Join(dir, "checksums.txt")
	if err := os.WriteFile(sums, []byte(want+"  asgard-fde-cli_0.1.2_darwin_all.tar.gz\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verify(sums, want); err != nil {
		t.Errorf("a real sha256 in goreleaser's own format was rejected: %v", err)
	}
}
