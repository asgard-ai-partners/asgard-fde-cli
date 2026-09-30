package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// The workspace's audit log is Asgard Console's (Explore), read through the
// platform with the member's own token. Console decides who may read it - a
// workspace owner or a platform admin, IAM action audit-log/read - and what
// the rows look like; the platform relays both unchanged, a 403 included.

// AuditDimensions are the dimensions AuditOptions takes, in the order a
// summary reads them.
var AuditDimensions = []string{"event", "user_identity_hint", "namespace", "bot_provider", "agent", "completion_model", "toolset", "semantic_model"}

// AuditOptionValue is one value a dimension takes in a range, with how many
// events carry it.
type AuditOptionValue struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// AuditOptions is one dimension's values.
type AuditOptions struct {
	Dimension string             `json:"dimension"`
	Values    []AuditOptionValue `json:"values"`
}

// AuditAccount is an IAM identity as a row's user_identity_hint names it.
type AuditAccount struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// AuditProject is a project as a row's namespace names it.
type AuditProject struct {
	ID   string `json:"project_id"`
	Name string `json:"project_name"`
}

// AuditDictionary translates the raw keys on audit rows to display names. A
// key with no entry is shown as itself.
type AuditDictionary struct {
	Products map[string]string       `json:"products"`
	Events   map[string]string       `json:"events"`
	Projects map[string]AuditProject `json:"projects"`
	Accounts map[string]AuditAccount `json:"accounts"`
	// Per namespace: raw CR name -> display name.
	Agents           map[string]map[string]string `json:"agents"`
	BotProviders     map[string]map[string]string `json:"bot_providers"`
	CompletionModels map[string]map[string]string `json:"completion_models"`
	Toolsets         map[string]map[string]string `json:"toolsets"`
	SemanticModels   map[string]map[string]string `json:"semantic_models"`
}

// GetAuditDictionary reads the workspace's audit dictionary.
func (c *Client) GetAuditDictionary(ctx context.Context) (*AuditDictionary, error) {
	var out AuditDictionary
	err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/audit-log/dictionary", out: &out})
	return &out, err
}

// GetAuditOptions reads the values one dimension takes in [from, to).
func (c *Client) GetAuditOptions(ctx context.Context, dimension string, from, to time.Time) (*AuditOptions, error) {
	var out AuditOptions
	q := url.Values{"from": {from.UTC().Format(time.RFC3339)}, "to": {to.UTC().Format(time.RFC3339)}}
	err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/audit-log/options/" + url.PathEscape(dimension), query: q, out: &out})
	return &out, err
}

// QueryAuditLog runs a Console Explore query (body is Console's query shape)
// and returns its answer as it streams: JSON Lines, a `meta` line, `row`
// lines, then `end` - or an `error` line once the stream has started. The
// caller closes it.
func (c *Client) QueryAuditLog(ctx context.Context, body json.RawMessage) (io.ReadCloser, error) {
	resp, err := c.doRaw(ctx, rawRequest{
		method:      http.MethodPost,
		path:        "/v1/workbench/audit-log/query",
		body:        bytes.NewReader(body),
		contentType: "application/json",
		accept:      "application/x-ndjson",
		noTimeout:   true,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
