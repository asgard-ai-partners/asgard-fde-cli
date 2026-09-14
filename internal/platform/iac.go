package platform

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// The `/v1/iac` surface. Field names mirror the platform's REST models exactly,
// because a rename here is a bug that only shows up as an empty column.
//
// Behaviour is specified in asgard-odin-pm `docs/spec/studio/pipeline.md`, and
// this file deliberately holds no rules of its own: which lint rule fails a
// plan, when a run is superseded, what a variable's type coercion does are all
// the platform's, and the CLI reports what it is told.

// DeclaredTrigger is a release's `on:` block.
type DeclaredTrigger struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

// DeclaredKey is one entry of chartValues / appSecret / appConfigMap.
type DeclaredKey struct {
	Kind      string `json:"kind"`
	Key       string `json:"key"`
	Required  bool   `json:"required"`
	ValueType string `json:"value_type,omitempty"`
}

// DeclaredRelease is one `releases:` entry as the platform read it.
type DeclaredRelease struct {
	Name  string           `json:"name"`
	On    *DeclaredTrigger `json:"on,omitempty"`
	Chart string           `json:"chart"`
	Keys  []*DeclaredKey   `json:"keys,omitempty"`
}

// ConfigSync is the outcome of the last read of `.asgard-pipeline.yaml`.
type ConfigSync struct {
	CommitSha string     `json:"commit_sha"`
	Ok        bool       `json:"ok"`
	Error     string     `json:"error,omitempty"`
	SyncedAt  *time.Time `json:"synced_at"`
	// Ref is the ref the sync read: the pipeline's config_ref, or its default
	// branch while that is empty. Empty against a gateway older than
	// asgard-platform-api#482, which dropped it.
	Ref string `json:"ref,omitempty"`
}

// DeclarationSource says which `.asgard-pipeline.yaml` a set of declarations was
// read from.
//
// It is the only explanation for why a key is marked Orphan: a release that has
// run reads its own most recent run's config, and one that has not falls back
// to the pipeline's snapshot. Neither decides a deployment - a run always reads
// the yaml at its own commit.
type DeclarationSource struct {
	Ref       string `json:"ref"`
	CommitSha string `json:"commit_sha,omitempty"`
	// RunNumber is the run it came from, as "#N"; 0 when it is the fallback.
	RunNumber            int64 `json:"run_number"`
	FromPipelineSnapshot bool  `json:"from_pipeline_snapshot"`
}

// Describe renders a declaration source the way the spec says to state it.
func (d *DeclarationSource) Describe() string {
	if d == nil || d.Ref == "" {
		return "unknown"
	}
	short := d.CommitSha
	if len(short) > 7 {
		short = short[:7]
	}
	if d.FromPipelineSnapshot {
		return "not run yet; showing " + d.Ref + "'s declarations for now"
	}
	if short == "" {
		return "read from " + d.Ref
	}
	return "read from " + d.Ref + " (" + short + ")"
}

// RunSummary is a run as a list row carries it: no steps, no report.
type RunSummary struct {
	RunId        string     `json:"run_id"`
	Number       int64      `json:"number"`
	ReleaseId    string     `json:"release_id"`
	ReleaseName  string     `json:"release_name"`
	Trigger      string     `json:"trigger"`
	Ref          string     `json:"ref"`
	State        string     `json:"state"`
	Actor        string     `json:"actor"`
	QueuedAt     *time.Time `json:"queued_at"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// Pipeline binds one repository to a set of releases.
type Pipeline struct {
	PipelineId       string `json:"pipeline_id"`
	WorkspaceId      string `json:"workspace_id"`
	Name             string `json:"name"`
	ConnectionId     string `json:"connection_id"`
	Provider         string `json:"provider"`
	ConnectionStatus string `json:"connection_status"`
	RepositoryId     string `json:"repository_id"`
	RepoFullName     string `json:"repo_full_name"`
	DefaultBranch    string `json:"default_branch"`
	// ConfigRef pins the declaration source; empty follows DefaultBranch.
	ConfigRef        string             `json:"config_ref,omitempty"`
	ConfigPath       string             `json:"config_path"`
	LastConfigSync   *ConfigSync        `json:"last_config_sync,omitempty"`
	DeclaredReleases []*DeclaredRelease `json:"declared_releases,omitempty"`
	ReleaseCount     int32              `json:"release_count"`
	LastRun          *RunSummary        `json:"last_run,omitempty"`
	CreatedBy        string             `json:"created_by"`
	CreatedAt        *time.Time         `json:"created_at"`
	UpdatedAt        *time.Time         `json:"updated_at"`
}

// Release is one deployable unit: one chart into one platform Project's
// namespace - the platform's own object, not a project in a customer's repo.
type Release struct {
	ReleaseId            string             `json:"release_id"`
	PipelineId           string             `json:"pipeline_id"`
	WorkspaceId          string             `json:"workspace_id"`
	Name                 string             `json:"name"`
	ProjectId            string             `json:"project_id"`
	Namespace            string             `json:"namespace"`
	ProjectEnvironmentId string             `json:"project_environment_id"`
	HelmReleaseName      string             `json:"helm_release_name"`
	AppSecretName        string             `json:"app_secret_name"`
	AppConfigMapName     string             `json:"app_config_map_name"`
	AutoApply            bool               `json:"auto_apply"`
	State                string             `json:"state"`
	DeleteStep           string             `json:"delete_step,omitempty"`
	DeleteError          string             `json:"delete_error,omitempty"`
	DetachError          string             `json:"detach_error,omitempty"`
	Declaration          *DeclaredRelease   `json:"declaration,omitempty"`
	DeclarationSource    *DeclarationSource `json:"declaration_source,omitempty"`
	PendingDeploy        bool               `json:"pending_deploy"`
	LastRun              *RunSummary        `json:"last_run,omitempty"`
	CreatedBy            string             `json:"created_by"`
	CreatedAt            *time.Time         `json:"created_at"`
	UpdatedAt            *time.Time         `json:"updated_at"`
}

// Release states, as the platform names them. A release is Active until
// somebody starts a teardown; Deleting is the runner working through it, and
// DeleteFailed is where one that failed a step is parked - with DeleteStep
// naming the step, which is what makes a retry meaningful rather than a
// second attempt at the whole thing.
const (
	ReleaseActive       = "active"
	ReleaseDeleting     = "deleting"
	ReleaseDeleteFailed = "delete_failed"
)

// Variable is one stored value, addressed by (kind, key).
//
// Value is always empty for a secret: the platform does not read one back once
// it is stored, and IsSet plus LineCount are all it reveals. LineCount is what
// lets somebody confirm a pasted PEM arrived whole without seeing it.
type Variable struct {
	ReleaseId   string     `json:"release_id"`
	Kind        string     `json:"kind"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	IsSet       bool       `json:"is_set"`
	LineCount   int32      `json:"line_count"`
	Description string     `json:"description,omitempty"`
	Declared    bool       `json:"declared"`
	Required    bool       `json:"required"`
	ValueType   string     `json:"value_type,omitempty"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// VariableList is the stored rows plus the declared keys that have none.
type VariableList struct {
	Variables           []*Variable    `json:"variables"`
	MissingDeclaredKeys []*DeclaredKey `json:"missing_declared_keys"`
	// DeclarationSource says which config the declared / required / missing
	// answers came from. State it above the table.
	DeclarationSource *DeclarationSource `json:"declaration_source,omitempty"`
}

// LintFinding is one rule hit in a plan report.
type LintFinding struct {
	Level    string `json:"level"`
	Rule     string `json:"rule"`
	Location string `json:"location,omitempty"`
	Message  string `json:"message"`
}

// VariableDiff is the plan report's row for one variable. It carries no value,
// for any kind.
type VariableDiff struct {
	Kind   string `json:"kind"`
	Key    string `json:"key"`
	Change string `json:"change"`
}

// ResourceDiff is the plan report's row for one rendered CR.
type ResourceDiff struct {
	ApiVersion  string `json:"api_version"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Change      string `json:"change"`
	UnifiedDiff string `json:"unified_diff,omitempty"`
}

// DryRunResult is the server-side dry run's outcome.
type DryRunResult struct {
	Ok      bool     `json:"ok"`
	Summary string   `json:"summary,omitempty"`
	Output  []string `json:"output,omitempty"`
}

// PlanReport is what a reviewer reads before approving.
type PlanReport struct {
	Lint      []*LintFinding  `json:"lint,omitempty"`
	Variables []*VariableDiff `json:"variables,omitempty"`
	Resources []*ResourceDiff `json:"resources,omitempty"`
	DryRun    *DryRunResult   `json:"dry_run,omitempty"`
	NoChanges bool            `json:"no_changes"`
}

// RunStep is one of the six fixed steps.
type RunStep struct {
	Name         string     `json:"name"`
	State        string     `json:"state"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	DurationMs   int64      `json:"duration_ms"`
	LogLineCount int32      `json:"log_line_count"`
	Error        string     `json:"error,omitempty"`
}

// Run is one Plan, Review and Apply of one release.
type Run struct {
	RunId         string      `json:"run_id"`
	PipelineId    string      `json:"pipeline_id"`
	ReleaseId     string      `json:"release_id"`
	ReleaseName   string      `json:"release_name"`
	WorkspaceId   string      `json:"workspace_id"`
	Number        int64       `json:"number"`
	Trigger       string      `json:"trigger"`
	RepoFullName  string      `json:"repo_full_name"`
	Ref           string      `json:"ref"`
	CommitSha     string      `json:"commit_sha"`
	CommitMessage string      `json:"commit_message,omitempty"`
	Actor         string      `json:"actor"`
	State         string      `json:"state"`
	Steps         []*RunStep  `json:"steps,omitempty"`
	Report        *PlanReport `json:"report,omitempty"`
	Reviewer      string      `json:"reviewer,omitempty"`
	ReviewedAt    *time.Time  `json:"reviewed_at"`
	ReviewComment string      `json:"review_comment,omitempty"`
	SupersededBy  string      `json:"superseded_by,omitempty"`
	QueuedAt      *time.Time  `json:"queued_at"`
	StartedAt     *time.Time  `json:"started_at"`
	FinishedAt    *time.Time  `json:"finished_at"`
	ErrorMessage  string      `json:"error_message,omitempty"`
	FailedStep    string      `json:"failed_step,omitempty"`
}

// StepLog is one step's captured output.
type StepLog struct {
	Lines     []string `json:"lines"`
	Truncated bool     `json:"truncated"`
}

// LiveObject is one object of a release's deployed manifest, read back from the
// cluster verbatim.
//
// Found is false both when the object was deleted outside the pipeline and when
// it could not be read; Error tells the two apart, and only one of them is a
// problem with the caller's access rather than with the cluster's contents.
type LiveObject struct {
	ApiVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Found      bool   `json:"found"`
	Yaml       string `json:"yaml,omitempty"`
	Error      string `json:"error,omitempty"`
}

// LiveManifest is what a release has on the cluster right now.
//
// Nothing in it is compared with anything: the platform has no diff endpoint on
// purpose, because the IaC source is here and not there. Comparing is this
// CLI's job.
type LiveManifest struct {
	HelmRevision int32         `json:"helm_revision"`
	DeployedAt   *time.Time    `json:"deployed_at"`
	HelmStatus   string        `json:"helm_status"`
	Objects      []*LiveObject `json:"objects"`
}

// Delivery is one entry of a pipeline's webhook history, including the reasons
// an event created no run.
type Delivery struct {
	DeliveryId         string     `json:"delivery_id"`
	PipelineId         string     `json:"pipeline_id"`
	ProviderDeliveryId string     `json:"provider_delivery_id,omitempty"`
	Event              string     `json:"event"`
	RepoFullName       string     `json:"repo_full_name"`
	Ref                string     `json:"ref,omitempty"`
	CommitSha          string     `json:"commit_sha,omitempty"`
	Outcome            string     `json:"outcome"`
	RunIds             []string   `json:"run_ids,omitempty"`
	ReceivedAt         *time.Time `json:"received_at"`
}

// ── pipelines ────────────────────────────────────────────────────────────

// ListPipelines returns the workspace's pipelines.
func (c *Client) ListPipelines(ctx context.Context) ([]*Pipeline, error) {
	return listAll[Pipeline](ctx, c, "/v1/iac/pipelines", nil)
}

// GetPipeline returns one pipeline, with its declared releases.
func (c *Client) GetPipeline(ctx context.Context, pipelineID string) (*Pipeline, error) {
	var out Pipeline
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/pipelines/" + url.PathEscape(pipelineID),
		out:    &out,
	})
	return &out, err
}

// RefreshConfig re-reads the declaration and updates the pipeline's snapshot.
//
// It changes no run and triggers no deployment: every run reads the config on
// its own commit, so this only moves what the platform displays.
func (c *Client) RefreshConfig(ctx context.Context, pipelineID string) (*Pipeline, error) {
	var out Pipeline
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/pipelines/" + url.PathEscape(pipelineID) + "/refresh-config",
		out:    &out,
	})
	return &out, err
}

// ListDeliveries returns a pipeline's recent webhook events and what came of
// each, which is where an event that created no run says why.
func (c *Client) ListDeliveries(ctx context.Context, pipelineID string, limit int) ([]*Delivery, error) {
	var out []*Delivery
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/pipelines/" + url.PathEscape(pipelineID) + "/deliveries",
		query:  url.Values{"size": {strconv.Itoa(limit)}},
		out:    &out,
	})
	return out, err
}

// ── releases ─────────────────────────────────────────────────────────────

// ListReleases returns a pipeline's releases.
func (c *Client) ListReleases(ctx context.Context, pipelineID string) ([]*Release, error) {
	return listAll[Release](ctx, c, "/v1/iac/pipelines/"+url.PathEscape(pipelineID)+"/releases", nil)
}

// GetRelease returns one release.
func (c *Client) GetRelease(ctx context.Context, releaseID string) (*Release, error) {
	var out Release
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID),
		out:    &out,
	})
	return &out, err
}

// UpdateReleaseInput is what PATCH accepts. The field is a pointer so that
// "leave it alone" and "set it to false" are different requests - and today
// auto_apply is the whole of it: the project and the name are fixed at create
// time and the platform moves neither.
type UpdateReleaseInput struct {
	AutoApply *bool `json:"auto_apply,omitempty"`
}

// UpdateRelease changes a release's mutable settings.
func (c *Client) UpdateRelease(ctx context.Context, releaseID string, in UpdateReleaseInput) (*Release, error) {
	var out Release
	err := c.do(ctx, request{
		method: http.MethodPatch,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID),
		body:   in,
		out:    &out,
	})
	return &out, err
}

// DestroyRelease starts the teardown that reaches the cluster: helm uninstall,
// then the release's Secret and ConfigMap, its RBAC, then the record.
//
// It returns as soon as the platform has accepted the work, with the release in
// `deleting`; the runner walks the steps afterwards and a failure parks it in
// `delete_failed` with the step that failed in DeleteStep. So what comes back
// is "started", never "gone", and RetryDestroyRelease is what resumes one.
func (c *Client) DestroyRelease(ctx context.Context, releaseID string) (*Release, error) {
	var out Release
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/destroy",
		out:    &out,
	})
	return &out, err
}

// RetryDestroyRelease resumes a `delete_failed` teardown from the step that
// failed.
func (c *Client) RetryDestroyRelease(ctx context.Context, releaseID string) (*Release, error) {
	var out Release
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/retry-destroy",
		out:    &out,
	})
	return &out, err
}

// DetachRelease removes the platform side and leaves the cluster alone: the
// deploy identity's RBAC is revoked and helm's release history is deleted
// WITHOUT uninstalling, then the record goes with its variables and its runs.
//
// Every CR, Secret and ConfigMap the release applied keeps running, for the
// project UI to own from then on. It is synchronous, and it answers with no
// release because there is none left to answer with.
func (c *Client) DetachRelease(ctx context.Context, releaseID string) error {
	return c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/detach",
	})
}

// GetLiveManifest reads back the objects the release's last helm revision
// deployed, as the cluster holds them now.
//
// managedFields is left out unless asked for: it is large and mostly noise, and
// it is also the only record of which field manager owns which field, which is
// what tells "the pipeline set this" from "somebody changed it in the UI".
func (c *Client) GetLiveManifest(ctx context.Context, releaseID string, includeManagedFields bool) (*LiveManifest, error) {
	q := url.Values{}
	if includeManagedFields {
		q.Set("include_managed_fields", "true")
	}
	var out LiveManifest
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/manifest",
		query:  q,
		out:    &out,
	})
	return &out, err
}

// ── variables ────────────────────────────────────────────────────────────

// GetVariables returns a release's stored values and the declared keys with no
// value yet.
func (c *Client) GetVariables(ctx context.Context, releaseID string) (*VariableList, error) {
	var out VariableList
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/variables",
		out:    &out,
	})
	return &out, err
}

// VariableWrite is one upsert in a PutVariables call.
//
// Description is a pointer so that "leave it alone" and "clear it" are
// different requests; Value is not, because a variable always has one.
type VariableWrite struct {
	Kind        string  `json:"kind"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
}

// PutVariables upserts values.
//
// It writes to the platform only. Nothing reaches the cluster until a run
// applies it, which is why the release then reports pending_deploy.
// It answers with the release's rows after the write, not with a VariableList:
// the declared-key half is a read concern and is not recomputed here.
func (c *Client) PutVariables(ctx context.Context, releaseID string, writes []VariableWrite) ([]*Variable, error) {
	var out []*Variable
	err := c.do(ctx, request{
		method: http.MethodPut,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/variables",
		body:   map[string]any{"variables": writes},
		out:    &out,
	})
	return out, err
}

// DeleteVariable removes one stored value.
func (c *Client) DeleteVariable(ctx context.Context, releaseID, kind, key string) error {
	return c.do(ctx, request{
		method: http.MethodDelete,
		path: "/v1/iac/releases/" + url.PathEscape(releaseID) +
			"/variables/" + url.PathEscape(kind) + "/" + url.PathEscape(key),
	})
}

// AddDeclaredKeys creates an empty row for every declared key that has none,
// which is the UI's "Add declared keys".
func (c *Client) AddDeclaredKeys(ctx context.Context, releaseID string) ([]*Variable, error) {
	var out []*Variable
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/variables/add-declared",
		out:    &out,
	})
	return out, err
}

// ── runs ─────────────────────────────────────────────────────────────────

// RunFilter narrows ListRuns.
type RunFilter struct {
	PipelineID string
	ReleaseID  string
	States     []string
	Size       int
}

// ListRuns returns runs newest first. Rows carry no steps and no report.
func (c *Client) ListRuns(ctx context.Context, f RunFilter) ([]*RunSummary, error) {
	q := url.Values{}
	if f.PipelineID != "" {
		q.Set("pipeline_id", f.PipelineID)
	}
	if f.ReleaseID != "" {
		q.Set("release_id", f.ReleaseID)
	}
	for _, s := range f.States {
		q.Add("state", s)
	}
	size := f.Size
	if size <= 0 {
		size = 20
	}
	q.Set("size", strconv.Itoa(size))

	var out []*RunSummary
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/runs",
		query:  q,
		out:    &out,
	})
	return out, err
}

// GetRun returns one run in full: its six steps and its plan report.
func (c *Client) GetRun(ctx context.Context, runID string) (*Run, error) {
	var out Run
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/runs/" + url.PathEscape(runID),
		out:    &out,
	})
	return &out, err
}

// GetStepLog returns one step's captured log.
func (c *Client) GetStepLog(ctx context.Context, runID, step string) (*StepLog, error) {
	var out StepLog
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/runs/" + url.PathEscape(runID) + "/steps/" + url.PathEscape(step) + "/log",
		out:    &out,
	})
	return &out, err
}

// CreateRun starts a manual run of a release at a ref.
//
// The release's trigger pattern is not checked; a mismatch is a plan warning
// rather than a refusal. A ref that does not resolve still leaves a run, parked
// at Checkout with the reason on it, so an attempt is always visible.
func (c *Client) CreateRun(ctx context.Context, releaseID, ref string) (*Run, error) {
	var out Run
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/releases/" + url.PathEscape(releaseID) + "/runs",
		body:   map[string]any{"ref": ref},
		out:    &out,
	})
	return &out, err
}

// ApproveRun approves a run waiting for review and starts its apply.
func (c *Client) ApproveRun(ctx context.Context, runID, comment string) (*Run, error) {
	return c.reviewRun(ctx, runID, "approve", comment)
}

// RejectRun rejects a run waiting for review.
func (c *Client) RejectRun(ctx context.Context, runID, comment string) (*Run, error) {
	return c.reviewRun(ctx, runID, "reject", comment)
}

func (c *Client) reviewRun(ctx context.Context, runID, action, comment string) (*Run, error) {
	body := map[string]any{}
	if comment != "" {
		body["comment"] = comment
	}
	var out Run
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/runs/" + url.PathEscape(runID) + "/" + action,
		body:   body,
		out:    &out,
	})
	return &out, err
}

// CancelRun stops a run.
//
// The outcome depends on where it was: cancelling before apply is `cancelled`,
// and cancelling during apply is `apply_failed`, because the cluster may
// already be partly updated and nothing is rolled back.
func (c *Client) CancelRun(ctx context.Context, runID string) (*Run, error) {
	var out Run
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/runs/" + url.PathEscape(runID) + "/cancel",
		out:    &out,
	})
	return &out, err
}

// ── run states ───────────────────────────────────────────────────────────

// Run states, as the platform names them.
const (
	RunQueued         = "queued"
	RunPlanning       = "planning"
	RunPlanFailed     = "plan_failed"
	RunAwaitingReview = "awaiting_review"
	RunApplying       = "applying"
	RunSucceeded      = "succeeded"
	RunApplyFailed    = "apply_failed"
	RunRejected       = "rejected"
	RunSuperseded     = "superseded"
	RunExpired        = "expired"
	RunCancelled      = "cancelled"
)

// Terminal reports whether a run state will not change again. It is the whole
// stopping condition for `runs watch`, which is why it lives with the states
// rather than in the command.
func Terminal(state string) bool {
	switch state {
	case RunPlanFailed, RunSucceeded, RunApplyFailed, RunRejected, RunSuperseded, RunExpired, RunCancelled:
		return true
	}
	return false
}

// Failed reports whether a terminal state is a bad one, which is what decides
// the exit code an agent branches on.
func Failed(state string) bool {
	switch state {
	case RunPlanFailed, RunApplyFailed, RunRejected, RunExpired, RunCancelled:
		return true
	}
	return false
}

// StepNames are the six steps, in order.
var StepNames = []string{"checkout", "lint", "variables", "render_dry_run", "review", "apply"}

// ValidStep reports whether name is one of the six.
func ValidStep(name string) bool {
	for _, s := range StepNames {
		if s == name {
			return true
		}
	}
	return false
}

// StepList renders the six names for an error message.
func StepList() string {
	out := ""
	for i, s := range StepNames {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

// ErrNoSuchRelease reports a release name that the pipeline does not have.
type ErrNoSuchRelease struct {
	Name     string
	Pipeline string
	Known    []string
}

func (e *ErrNoSuchRelease) Error() string {
	if len(e.Known) == 0 {
		return fmt.Sprintf("pipeline %s has no release %q, and no releases at all yet", e.Pipeline, e.Name)
	}
	return fmt.Sprintf("pipeline %s has no release %q; it has %v", e.Pipeline, e.Name, e.Known)
}

// ── connections and repositories ─────────────────────────────────────────

// VcsConnection is one provider installation a workspace can build pipelines
// on.
type VcsConnection struct {
	ConnectionId   string `json:"connection_id"`
	WorkspaceId    string `json:"workspace_id"`
	Provider       string `json:"provider"`
	InstallationId string `json:"installation_id"`
	AccountLogin   string `json:"account_login"`
	AccountType    string `json:"account_type"`
	Status         string `json:"status"`
	PipelineCount  int32  `json:"pipeline_count"`
	// RepositorySelection is what the installation grants on the provider:
	// "all" or "selected". Read only, and not ours to change — the control for
	// it is on the provider. It is the upper bound on what any pipeline through
	// this connection can reach, which is why it is worth a column.
	RepositorySelection string `json:"repository_selection"`
	// SharedWithWorkspaceCount is how many OTHER workspaces hold this same
	// installation. A count rather than names: "this account is already in use
	// elsewhere" is the part that changes what you do next.
	SharedWithWorkspaceCount int32      `json:"shared_with_workspace_count"`
	CreatedBy                string     `json:"created_by"`
	CreatedAt                *time.Time `json:"created_at"`
	UpdatedAt                *time.Time `json:"updated_at"`
}

// Repository is one repository visible through a connection.
//
// Pipelines bind by RepositoryId rather than by name, so a rename on the
// provider does not break one.
type Repository struct {
	RepositoryId  string `json:"repository_id"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	// Set when a pipeline in this workspace already binds this repository at
	// the same config path.
	AlreadyBoundByPipelineId string `json:"already_bound_by_pipeline_id,omitempty"`
}

// ListConnections returns the workspace's VCS connections.
// GetConnection returns one connection. Used when the platform reports a flow
// as connected but the workspace listing has not caught up: reporting a bare id
// would be worse than one more request.
func (c *Client) GetConnection(ctx context.Context, connectionID string) (*VcsConnection, error) {
	var out VcsConnection
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/connections/" + url.PathEscape(connectionID),
		out:    &out,
	})
	return &out, err
}

func (c *Client) ListConnections(ctx context.Context) ([]*VcsConnection, error) {
	return listAll[VcsConnection](ctx, c, "/v1/iac/connections", nil)
}

// listPageSize is what every paged listing asks for. It is the platform's
// maximum, so it is the fewest round trips, not a limit on the answer.
const listPageSize = 100

// maxListPages bounds the walk so a server that never stops advertising a
// total cannot spin here forever. It is not a product limit: at this page size
// it is a hundred thousand rows, which no engagement has.
const maxListPages = 100

// listAll fetches EVERY page of a paged collection.
//
// The platform answers a page at a time. A caller that takes the first page and
// stops does not get a short list — it gets a WRONG one, silently, and only for
// the accounts large enough to have a second page. That is how an organisation
// with 206 repositories was told its App could not reach a repository the App
// was in fact granted: the name sorted onto page three, and nothing in the
// tool, the message, or the help said a page was all it had.
func listAll[T any](ctx context.Context, c *Client, path string, query url.Values) ([]*T, error) {
	out := make([]*T, 0, listPageSize)
	for index := int64(0); index < maxListPages; index++ {
		q := url.Values{}
		for k, v := range query {
			q[k] = v
		}
		q.Set("size", strconv.FormatInt(listPageSize, 10))
		q.Set("page", strconv.FormatInt(index, 10))

		var batch []*T
		var paging Paging
		err := c.do(ctx, request{method: http.MethodGet, path: path, query: q, out: &batch, paging: &paging})
		if err != nil {
			return nil, err
		}
		out = append(out, batch...)
		// Stop on a short page as well as on the total: a server that reports
		// no total at all should end the walk, not restart it forever.
		if len(batch) < listPageSize || (paging.Total > 0 && int64(len(out)) >= paging.Total) {
			return out, nil
		}
	}
	return out, nil
}

// ListRepositories returns the repositories a connection can reach.
func (c *Client) ListRepositories(ctx context.Context, connectionID string) ([]*Repository, error) {
	return listAll[Repository](ctx, c, "/v1/iac/connections/"+url.PathEscape(connectionID)+"/repositories", nil)
}

// CreatePipelineInput is what binding a repository needs.
type CreatePipelineInput struct {
	Name         string `json:"name"`
	ConnectionId string `json:"connection_id"`
	RepositoryId string `json:"repository_id"`
	// ConfigPath is relative to the repository root; empty means the default.
	ConfigPath string `json:"config_path,omitempty"`
	// ConfigRef pins the declaration source; empty follows the repository's
	// default branch. It reaches the platform only once the gateway carries it
	// (asgard-platform-api#482); against an older one it is silently ignored,
	// which is why the pipeline's own config_ref is worth reading back.
	ConfigRef string `json:"config_ref,omitempty"`
}

// CreatePipeline binds a repository to a new pipeline.
func (c *Client) CreatePipeline(ctx context.Context, in CreatePipelineInput) (*Pipeline, error) {
	var out Pipeline
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/pipelines",
		body:   in,
		out:    &out,
	})
	return &out, err
}

// CreateReleaseInput is what creating a release needs. The platform resolves
// the project's namespace and main environment id itself.
type CreateReleaseInput struct {
	Name      string `json:"name"`
	ProjectId string `json:"project_id"`
	AutoApply bool   `json:"auto_apply"`
}

// CreateRelease creates a release under a pipeline, bound to one project.
func (c *Client) CreateRelease(ctx context.Context, pipelineID string, in CreateReleaseInput) (*Release, error) {
	var out Release
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/pipelines/" + url.PathEscape(pipelineID) + "/releases",
		body:   in,
		out:    &out,
	})
	return &out, err
}

// Providers the platform can connect. Phase 1 has one, and the list exists so
// that a CLI can name what is available rather than assume: the connection
// entity carries a provider field precisely because more are expected.
const ProviderGitHub = "github"

// Providers lists the providers a connection can be created for.
func Providers() []string { return []string{ProviderGitHub} }

// BeginInstall is what starting a provider installation returns: a URL to open
// and a single-use state bound to the workspace and the caller.
type BeginInstall struct {
	InstallUrl string     `json:"install_url"`
	State      string     `json:"state"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// BeginGitHubInstall mints the state and returns the GitHub App installation
// URL carrying it.
//
// The flow ends at the platform's own setup callback rather than back at the
// caller, so there is nothing here to wait on directly. What a caller does
// instead is watch the workspace's connections for a new one, which is the same
// thing to do whatever the provider turns out to be.
// InstallStatus is how one install flow ended, for whoever started it.
//
// It exists because the middle of that flow is a browser talking to the
// provider: this process never sees the callbacks, so without asking, every
// failure looks like the same non-event — no connection appeared.
type InstallStatus struct {
	// Outcome is one of the platform's outcome names. Compared as a string
	// rather than switched on an enum here: a value this binary has not heard
	// of should still print its Message, which is the part a person reads.
	Outcome string `json:"outcome"`
	// Message says what to do next. Present for every outcome.
	Message string `json:"message"`
	// ConnectionId is set only when the flow connected.
	ConnectionId string     `json:"connection_id"`
	ExpiresAt    *time.Time `json:"expires_at"`
	// Expired reports that the window closed with nothing decided.
	Expired bool `json:"expired"`
	// AttachState and Installations are set when the flow ended in a choice
	// rather than a connection.
	AttachState   string              `json:"attach_state"`
	Installations []*UserInstallation `json:"installations"`
}

// Settled reports whether this flow has stopped being worth waiting for.
func (s *InstallStatus) Settled() bool {
	return s != nil && (s.Expired || (s.Outcome != "" && s.Outcome != "pending"))
}

// GetInstallStatus asks how an install flow ended. The state is the handle
// BeginGitHubInstall returned.
func (c *Client) GetInstallStatus(ctx context.Context, state string) (*InstallStatus, error) {
	var out InstallStatus
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/iac/connections/install-status",
		query:  url.Values{"state": {state}},
		out:    &out,
	})
	return &out, err
}

// UserInstallation is one installation the authorizing person can reach.
type UserInstallation struct {
	InstallationId string `json:"installation_id"`
	AccountLogin   string `json:"account_login"`
	AccountType    string `json:"account_type"`
	// "all" or "selected" — what the installation grants on the provider.
	RepositorySelection string `json:"repository_selection"`
	// Set when this workspace already holds it.
	ConnectionId string `json:"connection_id"`
}

// AttachChoices is what the attach path returns instead of a connection.
type AttachChoices struct {
	Installations []*UserInstallation `json:"installations"`
	AttachState   string              `json:"attach_state"`
}

// BeginGitHubAttach starts the other way in: identify the person on the
// provider first, then pick from what it says they reach.
func (c *Client) BeginGitHubAttach(ctx context.Context) (*BeginInstall, error) {
	var out struct {
		AuthorizeUrl string     `json:"authorize_url"`
		State        string     `json:"state"`
		ExpiresAt    *time.Time `json:"expires_at"`
	}
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/connections/begin-github-attach",
		out:    &out,
	})
	if err != nil {
		return nil, err
	}
	return &BeginInstall{InstallUrl: out.AuthorizeUrl, State: out.State, ExpiresAt: out.ExpiresAt}, nil
}

// AttachInstallation connects one of the installations the authorizing person
// was shown to reach.
func (c *Client) AttachInstallation(ctx context.Context, attachState, installationID string) (*VcsConnection, error) {
	var out VcsConnection
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/connections/attach",
		body: map[string]string{
			"attach_state":    attachState,
			"installation_id": installationID,
		},
		out: &out,
	})
	return &out, err
}

func (c *Client) BeginGitHubInstall(ctx context.Context) (*BeginInstall, error) {
	var out BeginInstall
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/iac/connections/begin-github-install",
		out:    &out,
	})
	return &out, err
}

// ── projects ─────────────────────────────────────────────────────────────

// Project is one of the workspace's projects, which is what a release deploys
// into: the project decides the namespace, and the platform injects that plus
// the project's main environment id into every run.
type Project struct {
	ID          string `json:"project_id"`
	WorkspaceId string `json:"workspace_id"`
	Name        string `json:"project_name"`
	Namespace   string `json:"k8s_namespace_name"`
}

// CreateProject creates a project in the workspace, and with it the default
// environment that makes the project usable.
//
// **The environment is not a second call.** The platform's create does both -
// it writes the project, creates its default environment and stamps the main
// environment annotation - which matters because `release create` refuses a
// project that has none. A project made this way cannot be in that state; the
// ones that can are older than the requirement.
//
// It is not part of `/v1/iac` for the same reason ListProjects is not: a
// project is a platform concept the pipeline borrows. The workspace comes from
// the header every request here already carries.
//
// **It consumes quota.** The route is behind a subscription check, so a refusal
// here is an account limit rather than a bad argument, and the caller says so.
func (c *Client) CreateProject(ctx context.Context, name string, annotations map[string]any) (*Project, error) {
	body := map[string]any{"project_name": name}
	if len(annotations) > 0 {
		body["annotations"] = annotations
	}
	var out Project
	err := c.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/project",
		body:   body,
		out:    &out,
	})
	return &out, err
}

// ListProjects returns the workspace's projects.
//
// It is not part of `/v1/iac` - a project is a platform concept the pipeline
// borrows - but a release cannot be created without one, so the CLI needs it
// here rather than sending somebody to another tool for an id.
func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	var out []*Project
	err := c.do(ctx, request{
		method: http.MethodGet,
		path:   "/v1/project",
		query:  url.Values{"size": {"200"}},
		out:    &out,
	})
	return out, err
}

// ProjectEnvironment is one of a project's environments.
//
// A release needs the project's main one: the platform injects its id as
// `.Values.asgard.projectEnvironmentId`, and creating a release against a
// project that has none is refused. So this is what tells, before that refusal,
// which projects can host a release at all.
type ProjectEnvironment struct {
	// Key is the environment id.
	Key string `json:"key"`
	// Value is its display name.
	Value  string `json:"value"`
	IsMain bool   `json:"is_main"`
}

// ListProjectEnvironments returns a project's environments.
func (c *Client) ListProjectEnvironments(ctx context.Context, projectID string) ([]*ProjectEnvironment, error) {
	var out []*ProjectEnvironment
	err := c.do(ctx, request{
		method:  http.MethodGet,
		path:    "/v1/project/environments",
		project: projectID,
		out:     &out,
	})
	return out, err
}

// MainEnvironment returns a project's main environment, or nil when it has
// none.
func (c *Client) MainEnvironment(ctx context.Context, projectID string) (*ProjectEnvironment, error) {
	envs, err := c.ListProjectEnvironments(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, e := range envs {
		if e.IsMain {
			return e, nil
		}
	}
	return nil, nil
}
