package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// scope is what the CLI asks for. openid gets the subject, profile and email
// get a name to print in `whoami`. Nothing here asks for more than it shows.
const scope = "openid profile email"

// httpTimeout bounds every call to Casdoor. The browser half of the flow has no
// timeout of its own - a person may take a minute to find their password - but
// a single HTTP request that has not answered in half a minute is not going to.
const httpTimeout = 30 * time.Second

// pkce is one sign-in's proof that the program that started the flow is the
// program finishing it.
//
// The verifier never leaves the process until the code has already been
// received, so an attacker who intercepts the redirect - the reason a loopback
// flow needs this at all, since any program on the machine can race for the
// port - has a code they cannot exchange.
type pkce struct {
	verifier  string
	challenge string
}

func newPKCE() (pkce, error) {
	raw := make([]byte, 48)
	if _, err := rand.Read(raw); err != nil {
		return pkce{}, fmt.Errorf("generate a PKCE verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	return pkce{
		verifier:  verifier,
		challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
	}, nil
}

// randomState is the CSRF value echoed back on the redirect.
func randomState() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate a state value: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// authorizeURL is the page the browser is sent to.
//
// No client_secret appears here or at the token endpoint: this is a public
// client, and the code_challenge is what stands in for one.
func (p Profile) authorizeURL(redirectURI, state string, pk pkce) string {
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", scope)
	q.Set("state", state)
	q.Set("code_challenge", pk.challenge)
	q.Set("code_challenge_method", "S256")
	return p.AuthorizeURL() + "?" + q.Encode()
}

// tokenResponse is Casdoor's answer at the token endpoint. It answers OAuth
// errors in the same body as a success, so both shapes are read at once.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// toCredential turns a token response into a stored credential. expires_in is
// converted to an absolute time here, so that a credential read back tomorrow
// is not interpreted against tomorrow's clock as if it had just been issued.
func (t tokenResponse) toCredential() Credential {
	return Credential{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(t.ExpiresIn) * time.Second),
	}
}

// postToken calls the token endpoint and reads whichever of the two shapes came
// back.
func postToken(ctx context.Context, client *http.Client, p Profile, form url.Values) (tokenResponse, error) {
	form.Set("client_id", p.ClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, fmt.Errorf("build the token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("call %s: %w", p.TokenURL(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return tokenResponse{}, fmt.Errorf("read the token response: %w", err)
	}

	var out tokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return tokenResponse{}, fmt.Errorf("%s answered %s with something that is not JSON: %s",
			p.TokenURL(), resp.Status, truncate(string(body), 200))
	}
	if out.Error != "" {
		if out.ErrorDescription != "" {
			return tokenResponse{}, fmt.Errorf("%s: %s", out.Error, out.ErrorDescription)
		}
		return tokenResponse{}, fmt.Errorf("%s", out.Error)
	}
	if out.AccessToken == "" {
		return tokenResponse{}, fmt.Errorf("%s answered %s with no access token: %s",
			p.TokenURL(), resp.Status, truncate(string(body), 200))
	}
	return out, nil
}

// exchangeCode trades an authorization code for a token.
func exchangeCode(ctx context.Context, client *http.Client, p Profile, code, verifier, redirectURI string) (Credential, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	// Casdoor does not check redirect_uri at the token endpoint, but sending it
	// is what the spec asks for and costs nothing if that changes.
	form.Set("redirect_uri", redirectURI)

	out, err := postToken(ctx, client, p, form)
	if err != nil {
		return Credential{}, err
	}
	return out.toCredential(), nil
}

// Refresh exchanges a refresh token for a new access token.
//
// Casdoor accepts an empty client_secret on this endpoint too, so a public
// client can renew its own session; the refresh token it returns replaces the
// one used, and the caller stores whichever came back.
func Refresh(ctx context.Context, p Profile, refreshToken string) (Credential, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("scope", scope)

	out, err := postToken(ctx, newHTTPClient(), p, form)
	if err != nil {
		return Credential{}, err
	}
	cred := out.toCredential()
	if cred.RefreshToken == "" {
		// Keep the one that still works rather than dropping to a session that
		// cannot be renewed again.
		cred.RefreshToken = refreshToken
	}
	return cred, nil
}

// Userinfo is the subset of the OIDC userinfo response this CLI shows.
//
// This is the same endpoint the platform's IAM calls to verify a bearer token,
// which is what makes it the right check: a token it accepts is a token the
// platform accepts, so `whoami` reports the session the API would see rather
// than what the CLI believes about its own file.
type Userinfo struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	DisplayName       string `json:"displayName"`
	// Status and Msg appear only when Casdoor rejected the call: it answers
	// errors 200 with a status field rather than with an HTTP code.
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

// Who returns the display name to show for a session, preferring the one a
// person would recognise.
func (u Userinfo) Who() string {
	for _, s := range []string{u.DisplayName, u.Name, u.PreferredUsername, u.Email, u.Sub} {
		if s != "" {
			return s
		}
	}
	return "(unknown)"
}

// FetchUserinfo asks Casdoor who a token belongs to.
func FetchUserinfo(ctx context.Context, p Profile, accessToken string) (Userinfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.UserinfoURL(), nil)
	if err != nil {
		return Userinfo{}, fmt.Errorf("build the userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := newHTTPClient().Do(req)
	if err != nil {
		return Userinfo{}, fmt.Errorf("call %s: %w", p.UserinfoURL(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Userinfo{}, fmt.Errorf("read the userinfo response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Userinfo{}, fmt.Errorf("%s answered %s: %s", p.UserinfoURL(), resp.Status, truncate(string(body), 200))
	}

	var out Userinfo
	if err := json.Unmarshal(body, &out); err != nil {
		return Userinfo{}, fmt.Errorf("%s answered with something that is not JSON: %s", p.UserinfoURL(), truncate(string(body), 200))
	}
	if out.Status == "error" {
		return Userinfo{}, fmt.Errorf("%s rejected the token: %s", p.UserinfoURL(), out.Msg)
	}
	if out.Sub == "" {
		return Userinfo{}, fmt.Errorf("%s answered with no subject: %s", p.UserinfoURL(), truncate(string(body), 200))
	}
	return out, nil
}

func newHTTPClient() *http.Client { return &http.Client{Timeout: httpTimeout} }

// truncate keeps an unexpected response readable in an error message. A server
// that answers a login with an HTML error page would otherwise put the whole
// page on the terminal.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
