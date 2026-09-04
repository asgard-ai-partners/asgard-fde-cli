package platform

import (
	"context"
	"sync/atomic"
)

// `/v1/docs` is the reference material about the server itself: which
// processors exist, what each config key is called, what it defaults to.
//
// **It is served rather than shipped inside this binary, and the reason is
// on-prem.** A customer's Asgard server can be several versions behind this CLI
// or ahead of it, and a document saying what `llm-completion` takes is only
// true of one of them. Compiled in, it would be pinned to whichever release the
// customer installed - which has nothing to do with the server they deploy
// against - and unfixable without shipping them a new binary. Fetched, it
// describes the server their runs actually go to.
//
// It takes no workspace: it describes the software, not a tenant's data.

// DocsVersionHeader carries the platform's current version on every response.
//
// It is on responses to calls the CLI already makes, which is the whole point:
// an agent finds out its reference material is behind inside the loop it is
// already in - watching a run, listing variables - rather than by remembering
// to poll for it. Nothing has to be added to the agent's routine.
const DocsVersionHeader = "x-asgard-docs-version"

// lastDocsVersion is the most recent value any response carried.
//
// Package-level because the warning is about the process rather than about one
// client: commands build a client, make a call and return, and whoever prints
// the warning at the end has no handle on the client that saw the header.
var lastDocsVersion atomic.Value

// LastDocsVersion returns the version the platform last reported, or "" if no
// call in this process carried the header.
func LastDocsVersion() string {
	v, _ := lastDocsVersion.Load().(string)
	return v
}

// DocsBundle is the material and its identity.
type DocsBundle struct {
	Version     string       `json:"version"`
	GeneratedAt string       `json:"generated_at"`
	Sources     []DocsSource `json:"sources"`
	Files       []DocsFile   `json:"files"`
}

// DocsVersion is the identity alone, for asking whether what is on disk is
// still current without downloading what would replace it.
type DocsVersion struct {
	Version     string       `json:"version"`
	GeneratedAt string       `json:"generated_at"`
	Sources     []DocsSource `json:"sources"`
}

// DocsSource is one upstream the material was rendered from, so that "the
// processors moved" and "the CRDs moved" can be told apart.
type DocsSource struct {
	Name    string `json:"name"`
	Digest  string `json:"digest"`
	Records int    `json:"records"`
}

// DocsFile is one file to write. Path is relative to whatever directory this
// CLI keeps skills in: the content is the server's and the location is ours.
type DocsFile struct {
	Path    string `json:"path"`
	Digest  string `json:"digest"`
	Content string `json:"content"`
}

// DocsSkills fetches the material.
func (c *Client) DocsSkills(ctx context.Context) (*DocsBundle, error) {
	var out DocsBundle
	err := c.do(ctx, request{method: "GET", path: "/v1/docs/skills", out: &out, noWorkspace: true})
	return &out, err
}

// DocsVersionOnly asks only what the platform's version is.
func (c *Client) DocsVersionOnly(ctx context.Context) (*DocsVersion, error) {
	var out DocsVersion
	err := c.do(ctx, request{method: "GET", path: "/v1/docs/version", out: &out, noWorkspace: true})
	return &out, err
}
