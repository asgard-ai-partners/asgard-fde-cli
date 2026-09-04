package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// Source says where a session's token came from.
type Source string

const (
	// SourceEnv is ASGARD_TOKEN. Nothing is read from or written to the
	// credential store, which is what makes it the right thing for CI and for
	// an agent sandbox: no state on the machine, and no browser.
	SourceEnv Source = "env"
	// SourceStore is a credential written by `login`.
	SourceStore Source = "store"
)

// Session is an authenticated profile: which platform, and a token that was
// valid when it was resolved.
type Session struct {
	Profile Profile
	Token   string
	Source  Source

	// Subject, Email and Name are what the store recorded at sign-in. They are
	// empty for an ASGARD_TOKEN session, which carries no such record - `whoami`
	// asks userinfo rather than guessing.
	Subject string
	Email   string
	Name    string
}

// NeedsLoginError reports a command that needs a session where there is none,
// and carries the one line that fixes it.
//
// It is a type rather than a sentinel because the instruction has to name the
// profile: an FDE with a dev session running a prod command gets "run `login
// --profile prod`", not a generic "not signed in" that they will satisfy by
// checking the session they already have.
type NeedsLoginError struct {
	Profile string
	// Reason is what was wrong: absent, expired past refreshing, or refused.
	Reason string
}

func (e *NeedsLoginError) Error() string {
	profile := ""
	if e.Profile != "" && e.Profile != DefaultProfileName {
		profile = " --profile " + e.Profile
	}
	reason := e.Reason
	if reason == "" {
		reason = "no session for profile " + e.Profile
	}
	return fmt.Sprintf("%s; run `asgard-cli login%s`", reason, profile)
}

func (e *NeedsLoginError) Is(target error) bool { return target == ErrNotSignedIn }

// Resolve produces a usable session for a profile, renewing the access token
// when it has expired.
//
// A renewed token is written back before it is used, so that a command which
// then fails for its own reasons has not thrown away the refresh it just did -
// the next command would otherwise refresh again, and Casdoor issues a new
// refresh token each time.
func Resolve(ctx context.Context, profileName string) (*Session, error) {
	p, err := ResolveProfile(profileName)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv(EnvToken); token != "" {
		return &Session{Profile: p, Token: token, Source: SourceEnv}, nil
	}

	cred, err := LoadCredential(p)
	if errors.Is(err, ErrNotSignedIn) {
		return nil, &NeedsLoginError{Profile: p.Name, Reason: "not signed in to the " + p.Name + " platform"}
	}
	if err != nil {
		return nil, err
	}

	if cred.Expired() {
		if !cred.CanRefresh() {
			return nil, &NeedsLoginError{Profile: p.Name, Reason: "the " + p.Name + " session has expired"}
		}
		refreshed, err := Refresh(ctx, p, cred.RefreshToken)
		if err != nil {
			// A refresh token is good for 30 days, so a refusal here is almost
			// always a session that was revoked or simply left too long. Either
			// way the remedy is the same one.
			return nil, &NeedsLoginError{
				Profile: p.Name,
				Reason:  fmt.Sprintf("the %s session could not be renewed (%v)", p.Name, err),
			}
		}
		refreshed.Subject, refreshed.Email, refreshed.Name = cred.Subject, cred.Email, cred.Name
		if err := SaveCredential(p, refreshed); err != nil {
			return nil, err
		}
		cred = refreshed
	}

	return &Session{
		Profile: p,
		Token:   cred.Token(),
		Source:  SourceStore,
		Subject: cred.Subject,
		Email:   cred.Email,
		Name:    cred.Name,
	}, nil
}

// Token returns the access token. It is a method so that a credential is never
// printed by accident through a struct format verb.
func (c Credential) Token() string { return c.AccessToken }
