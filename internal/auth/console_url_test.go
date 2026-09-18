package auth

import "testing"

// **A guessed Console URL is worse than none**, because it renders correctly,
// gets copied onto a slide, and dies in front of a room. So the field is
// recorded rather than derived, and `ConsoleURL` reports honestly when nothing
// has recorded one.
func TestConsoleURLIsWhatWasRecordedAndNothingElse(t *testing.T) {
	if got, ok := (Profile{Console: "https://console.acme.internal"}).ConsoleURL(); !ok || got != "https://console.acme.internal" {
		t.Errorf("a recorded Console: got %q ok=%v", got, ok)
	}
	// An installation that names its own API and no Console has none. A rule
	// that stripped `-api` from the API would answer here and be wrong.
	for _, api := range []string{
		"https://platform-api.example.com",
		"https://asgard-api.internal.corp",
		"https://platform-api.asgard-ai.com.evil.test",
		"http://localhost:8080",
		"",
	} {
		p := Profile{PlatformAPI: api}
		if got, ok := p.ConsoleURL(); ok || got != "" {
			t.Errorf("PlatformAPI %q with no Console: got %q ok=%v, want no answer", api, got, ok)
		}
	}
}

// **The Console is the one field that does not inherit the hosted value on its
// own.** Every other field does, correctly: a profile naming only its own API
// still authenticates against the same Casdoor. Inheriting the Console would
// hand a self-hosted installation a link into ANOTHER organisation's console -
// one that resolves, renders, and is wrong.
func TestTheHostedConsoleIsInheritedOnlyWithTheHostedAPI(t *testing.T) {
	t.Setenv(EnvProfile, DefaultProfileName)

	r, err := ResolveWithOrigin("")
	if err != nil {
		t.Fatal(err)
	}
	if r.Console != hostedConsole || r.ConsoleFrom != FromHosted {
		t.Errorf("hosted default: Console=%q from=%v, want %q hosted", r.Console, r.ConsoleFrom, hostedConsole)
	}

	// The same resolution with somebody else's API and nothing else said.
	t.Setenv(EnvPlatformAPI, "https://asgard.acme.internal")
	r, err = ResolveWithOrigin("")
	if err != nil {
		t.Fatal(err)
	}
	if r.Console != "" {
		t.Errorf("a self-hosted API inherited a Console: %q - that is another organisation's", r.Console)
	}

	// And recording one is what makes it available.
	t.Setenv(EnvConsole, "https://console.acme.internal/")
	r, err = ResolveWithOrigin("")
	if err != nil {
		t.Fatal(err)
	}
	if r.Console != "https://console.acme.internal" {
		t.Errorf("recorded Console not used or not trimmed: %q", r.Console)
	}
}
