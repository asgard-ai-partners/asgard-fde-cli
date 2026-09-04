package skills

import (
	"crypto/sha256"
	"encoding/hex"
)

// Digest is the platform's digest, computed the same way, so a file's recorded
// digest and the one the platform sends for the same bytes are comparable.
//
// Twelve hex characters: long enough that a collision is not a thing that
// happens, short enough to read in a terminal beside a filename.
func Digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}
