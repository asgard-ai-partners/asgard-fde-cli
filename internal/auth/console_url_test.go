package auth

import "testing"

// **A guessed Console URL is worse than none**, because it renders correctly,
// gets copied onto a slide, and dies in front of a room. The hosted API and the
// hosted Console differ by more than a path, and that `platform-api.X` and
// `platform.X` look related is a coincidence of naming rather than a rule - so
// an installation that names its own API gets no answer at all.
func TestConsoleURLIsKnownOnlyForTheHostedInstallation(t *testing.T) {
	got, ok := Profile{PlatformAPI: hostedPlatformAPI}.ConsoleURL()
	if !ok || got != hostedConsole {
		t.Errorf("hosted profile: got %q ok=%v, want %q true", got, ok, hostedConsole)
	}
	if got == hostedPlatformAPI {
		t.Error("the Console is not the API host, and this returned the API's")
	}

	// Every other shape of profile, including ones that look like the hosted
	// one. A rule that stripped `-api` would pass the first of these and be
	// wrong about every installation that does not follow that naming.
	for _, api := range []string{
		"https://platform-api.example.com",
		"https://asgard-api.internal.corp",
		"https://platform-api.asgard-ai.com.evil.test",
		"http://localhost:8080",
		"",
	} {
		p := Profile{PlatformAPI: api}
		if got, ok := p.ConsoleURL(); ok || got != "" {
			t.Errorf("PlatformAPI %q: got %q ok=%v, want no answer", api, got, ok)
		}
	}
}
