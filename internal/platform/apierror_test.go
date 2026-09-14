package platform

import (
	"net/http"
	"strings"
	"testing"
)

// **A 5xx on a write does not mean the write did not happen**, and the message
// used to read as though it did.
//
// The platform answered 500 to two of four `pipeline project create` calls and
// had created both. `pipeline projects` listed all four, each once. The obvious
// response to a 500 is to retry, and a retry there makes a duplicate platform
// Project that only the Console can remove - and removing one removes whatever
// is deployed into it.
func TestAPIErrorSaysAWriteMayHaveLanded(t *testing.T) {
	write := (&APIError{Status: 500, Method: http.MethodPost, Path: "/projects",
		Message: "Internal server error"}).Error()
	if !strings.Contains(write, "may have been applied") {
		t.Errorf("a 500 on a POST reads as a failure:\n%s", write)
	}

	// A read is free to retry, and saying otherwise would train the reader to
	// ignore the sentence where it matters.
	read := (&APIError{Status: 500, Method: http.MethodGet, Path: "/projects",
		Message: "Internal server error"}).Error()
	if strings.Contains(read, "may have been applied") {
		t.Errorf("a 500 on a GET should not warn about a partial write:\n%s", read)
	}

	// The statuses that already say something specific keep saying it.
	for _, c := range []struct{ status, want int }{{401, 401}, {403, 403}, {404, 404}} {
		got := (&APIError{Status: c.status, Method: http.MethodPost}).Error()
		if strings.Contains(got, "may have been applied") {
			t.Errorf("%d is a definite answer and should not read as unknown:\n%s", c.status, got)
		}
	}
}
