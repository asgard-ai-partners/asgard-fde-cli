// Package platform calls the Asgard platform API.
//
// It is a wrapper and nothing more. The pipeline the CLI drives - lint, render,
// dry run against the cluster's CRDs, plan, apply - runs on the platform, and
// none of it is reimplemented here: a second copy of a rule is a copy that
// disagrees with the server the first time either changes, and the checks worth
// the most (the apiserver's own CEL, pattern and required validation) need a
// cluster the CLI is deliberately never given credentials for.
//
// So what an agent does with this package is push its work and read back what
// the platform made of it. The local half of the loop is the native tools -
// `helm lint`, `helm template` - plus `asgard-cli verify`, which checks the
// things that pass a dry run and still fail at runtime.
package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
)

// timeout bounds one API call.
//
// It is longer than a page load because some of these calls wait on the
// platform doing real work, and shorter than a run because nothing here polls
// inside a single request - `runs watch` polls by making more calls, so a
// stalled one fails and is retried rather than holding the terminal.
const timeout = 60 * time.Second

// WorkspaceHeader is the header every workspace-scoped route requires.
const WorkspaceHeader = "x-asgard-workspace"

// Client is an authenticated caller of one platform.
type Client struct {
	profile   auth.Profile
	token     string
	workspace string
	http      *http.Client
}

// New builds a client from a resolved session. The workspace may be empty for
// the routes that do not take one - listing workspaces is the one that matters,
// because it is how a workspace is chosen in the first place.
func New(session *auth.Session, workspace string) *Client {
	return &Client{
		profile:   session.Profile,
		token:     session.Token,
		workspace: workspace,
		http:      &http.Client{Timeout: timeout},
	}
}

// Profile reports which platform this client talks to, for messages that have
// to say which one answered.
func (c *Client) Profile() auth.Profile { return c.profile }

// Workspace reports the workspace id every scoped call is made against.
func (c *Client) Workspace() string { return c.workspace }

// APIError is a non-2xx answer from the platform, decoded.
type APIError struct {
	Status     int
	Method     string
	Path       string
	Message    string
	ReasonCode int32
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	switch e.Status {
	case http.StatusUnauthorized:
		return fmt.Sprintf("the platform rejected the session (%d %s); run `asgard-cli login`", e.Status, msg)
	case http.StatusForbidden:
		// Members may read a pipeline and edit variables; approving, running and
		// deleting need workspace administration. Saying so here saves reading
		// the permission matrix to find out which half a command needed.
		return fmt.Sprintf("not allowed (%d %s); viewing a pipeline and editing variables are open to workspace members, and running, approving and deleting need workspace administration", e.Status, msg)
	case http.StatusNotFound:
		return fmt.Sprintf("not found (%d %s): %s %s", e.Status, msg, e.Method, e.Path)
	}
	return fmt.Sprintf("the platform answered %d: %s", e.Status, msg)
}

// NotFound reports whether err is a 404, which several callers treat as an
// answer rather than a failure - a release that has never deployed has no live
// manifest, and that is information.
func NotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

// Unauthorized reports whether err is the platform refusing the session.
func Unauthorized(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized
}

// envelope is the platform's response wrapper. Every route answers in it, and
// an error answers in a shape that overlaps it, so one struct reads both.
type envelope struct {
	Data       json.RawMessage `json:"data"`
	Paging     *Paging         `json:"paging"`
	Message    string          `json:"message"`
	Success    bool            `json:"success"`
	ReasonCode int32           `json:"reason_code"`
}

// Paging is the platform's page descriptor.
type Paging struct {
	Total int64 `json:"total"`
	Index int64 `json:"index"`
	Size  int64 `json:"size"`
}

// ProjectHeader is the header the project-scoped routes require, alongside the
// workspace one.
const ProjectHeader = "x-asgard-project"

// request is one call: method, path below the API root, optional query and
// body, and where to put the decoded `data`.
type request struct {
	method string
	// path is appended to the API base, starting with a slash and including the
	// version segment - "/v1/iac/pipelines".
	path  string
	query url.Values
	body  any
	// out receives the `data` field. Nil discards it.
	out any
	// paging receives the `paging` field when the caller wants it.
	paging *Paging
	// noWorkspace skips the workspace header, for the routes that take none.
	noWorkspace bool
	// project sets the project header, for the routes scoped to one.
	project string
}

// do makes one call and unwraps the envelope.
func (c *Client) do(ctx context.Context, req request) error {
	endpoint := c.profile.API + req.path
	if len(req.query) > 0 {
		endpoint += "?" + req.query.Encode()
	}

	var bodyReader io.Reader
	if req.body != nil {
		encoded, err := json.Marshal(req.body)
		if err != nil {
			return fmt.Errorf("encode the request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("build the request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Accept", "application/json")
	if req.body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if !req.noWorkspace {
		if c.workspace == "" {
			return errors.New("no workspace selected; pass --workspace, or record one with `asgard-cli init --workspace-id`")
		}
		httpReq.Header.Set(WorkspaceHeader, c.workspace)
	}
	if req.project != "" {
		httpReq.Header.Set(ProjectHeader, req.project)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call %s %s: %w", req.method, endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("read the response to %s %s: %w", req.method, endpoint, err)
	}

	var env envelope
	decodeErr := json.Unmarshal(raw, &env)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		apiErr := &APIError{Status: resp.StatusCode, Method: req.method, Path: req.path}
		if decodeErr == nil {
			apiErr.Message = env.Message
			apiErr.ReasonCode = env.ReasonCode
		} else {
			// A body that is not the envelope is a gateway or proxy answering,
			// not the API. Keeping a slice of it is what tells those apart.
			apiErr.Message = truncate(string(raw), 200)
		}
		return apiErr
	}
	if decodeErr != nil {
		return fmt.Errorf("%s %s answered %s with something that is not the platform's response envelope: %s",
			req.method, endpoint, resp.Status, truncate(string(raw), 200))
	}

	if req.paging != nil && env.Paging != nil {
		*req.paging = *env.Paging
	}
	if req.out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, req.out); err != nil {
		return fmt.Errorf("%s %s answered with a body this build does not understand: %w", req.method, endpoint, err)
	}
	return nil
}

// Workspace is one workspace the signed-in user can reach.
type Workspace struct {
	ID   string `json:"workspace_id"`
	Name string `json:"display_name"`
}

// ListWorkspaces returns every workspace the session can see.
//
// It takes no workspace header, which makes it the one call that works before a
// workspace has been chosen - and the check that a session is good for the API
// and not only for Casdoor.
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var out []Workspace
	err := c.do(ctx, request{
		method:      http.MethodGet,
		path:        "/v1/workspace",
		out:         &out,
		noWorkspace: true,
	})
	return out, err
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
