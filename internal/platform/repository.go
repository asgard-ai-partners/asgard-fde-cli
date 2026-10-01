package platform

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// A GitHub credential for one repository, and a new repository, both through
// the workspace's GitHub Connection - the GitHub App installation, never the
// person's own account. They exist for the Workbench assistant's sandbox,
// where nobody has a GitHub login: `pipeline git-credential` hands the token
// to git, and `pipeline repo create` makes the repository a new pipeline will
// bind.

// IacAppPermissionNotGranted is the error code (403) of a token or a create
// the GitHub App installation was not granted the permission for. A 403 from
// the platform's own permissions carries no error code.
const IacAppPermissionNotGranted = "iac_app_permission_not_granted"

// RepositoryToken is a short-lived installation token for exactly one
// repository. git uses it as the password with the username
// "x-access-token".
//
// Token is a secret. String leaves it out, so a RepositoryToken that lands in
// an error or a log prints no credential; read the field explicitly, and only
// to hand it to git.
type RepositoryToken struct {
	Token        string    `json:"token"`
	Repository   string    `json:"repository"`
	Permission   string    `json:"permission"`
	ExpiresAt    time.Time `json:"expires_at"`
	ConnectionID string    `json:"connection_id"`
}

func (t RepositoryToken) String() string {
	return fmt.Sprintf("RepositoryToken{%s %s until %s}", t.Repository, t.Permission, t.ExpiresAt.UTC().Format(time.RFC3339))
}

// GoString is %#v.
func (t RepositoryToken) GoString() string { return t.String() }

// MintRepositoryToken asks for a token reaching repository ("owner/name").
// write asks for contents:write, which needs workspace administration; read
// is open to members.
func (c *Client) MintRepositoryToken(ctx context.Context, repository string, write bool) (*RepositoryToken, error) {
	level := "read"
	if write {
		level = "write"
	}
	var out RepositoryToken
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/repository-tokens/" + level,
		body:   map[string]string{"repository": repository},
		out:    &out,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatedRepository is a repository CreateRepository made: empty, with no
// initial commit.
type CreatedRepository struct {
	Repository Repository `json:"repository"`
	// CloneURL is the HTTPS clone URL, with no credentials in it.
	CloneURL string `json:"clone_url"`
}

// CreateRepository creates an empty repository under the connection's
// organization, private unless public. It needs workspace administration,
// works only on an organization's connection, and is never retried: a create
// whose answer was lost has still happened.
func (c *Client) CreateRepository(ctx context.Context, connectionID, name, description string, public bool) (*CreatedRepository, error) {
	var out CreatedRepository
	err := c.do(ctx, request{
		method:     http.MethodPost,
		sideEffect: true,
		path:       "/v1/iac/connections/" + url.PathEscape(connectionID) + "/repositories",
		body: map[string]any{
			"name":        name,
			"description": description,
			"public":      public,
		},
		out: &out,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
