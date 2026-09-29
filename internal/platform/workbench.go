package platform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// The Workbench is the workspace's issue tracker: `ISS-N` issues the FDE and
// the customer work through, with a timeline, comments and attachments. The
// routes live under /v1/workbench and take the workspace header like every
// other scoped route.

// ViaAssistantHeader marks a write as the Workbench assistant's.
//
// **It grants nothing.** Authorization is still the signed-in member's; what
// it changes is that the timeline says "via Asgard AI", and that the platform
// refuses the handful of actions only a member may take - deleting, pinning,
// locking, managing labels, deleting an attachment, and every comment write.
// The CLI sends it on every Workbench write, because whoever is running this
// command is an agent acting for the member, and the timeline should say so.
const ViaAssistantHeader = "X-Asgard-Via-Assistant"

// The error codes the Workbench routes put in `details.error_code`.
const (
	// WorkbenchConflictRetry is a write that lost a database race twice in a
	// row (409). Nothing was written.
	WorkbenchConflictRetry = "workbench_conflict_retry"
	// WorkbenchAssistantForbidden is an action only the member may take,
	// refused because the request said it was the assistant's (403).
	WorkbenchAssistantForbidden = "workbench_assistant_forbidden"
	// WorkbenchIssueLocked is a comment on a locked issue by somebody who is
	// not a workspace admin (403).
	WorkbenchIssueLocked = "workbench_issue_locked"
	// WorkbenchParentAlreadySet is a sub-issue added to a parent while it
	// already has another one (400).
	WorkbenchParentAlreadySet = "workbench_parent_already_set"
)

// workbenchErrorText says what a Workbench refusal means, for the ones where
// the generic text would send an agent the wrong way. The generic 403 is about
// pipeline permissions, which is not what a Workbench 403 is about.
func workbenchErrorText(e *APIError, msg string) (string, bool) {
	if !strings.HasPrefix(e.Path, "/v1/workbench") {
		return "", false
	}
	switch {
	case e.ErrorCode == WorkbenchConflictRetry:
		return fmt.Sprintf("the platform answered %d (%s): **nothing was written** - two writes to the same issue raced "+
			"and this one lost twice. Running the same command again is safe", e.Status, WorkbenchConflictRetry), true
	case e.ErrorCode == WorkbenchAssistantForbidden:
		return fmt.Sprintf("the platform answered %d (%s): this CLI writes to the Workbench as the member's assistant, "+
			"and this action is one only the member may take - deleting, pinning or locking an issue, managing labels, "+
			"deleting an attachment, or writing a comment. Nothing was written. Ask the member to do it in the Workbench page",
			e.Status, WorkbenchAssistantForbidden), true
	case e.ErrorCode == WorkbenchIssueLocked:
		return fmt.Sprintf("the platform answered %d (%s): the issue is locked, and only workspace admins may comment on it. "+
			"Its fields can still be changed", e.Status, WorkbenchIssueLocked), true
	case e.ErrorCode == WorkbenchParentAlreadySet:
		return fmt.Sprintf("the platform answered %d (%s): the sub-issue already has a parent, and an issue has at most one. "+
			"Nothing was written. Clear it first with --parent 0 on the sub-issue, or leave it where it is", e.Status, WorkbenchParentAlreadySet), true
	case e.Status == http.StatusForbidden:
		return fmt.Sprintf("not allowed (%d %s); the Workbench is open to every member of the workspace, so this account "+
			"is probably not a member of it, or the action needs workspace administration", e.Status, msg), true
	}
	return "", false
}

// CursorPaging is the Workbench routes' page descriptor.
type CursorPaging struct {
	NextPageToken string `json:"next_page_token"`
	// TotalCount is set on the issue list only: issues matching the filter
	// across every page.
	TotalCount int64 `json:"total_count"`
}

// WorkbenchIssue is one issue.
type WorkbenchIssue struct {
	Number      int64  `json:"number"`
	WorkspaceID string `json:"workspace_id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	// PipelineID is the pipeline the issue is about; empty when none.
	PipelineID       string                     `json:"pipeline_id"`
	AssigneeIDs      []string                   `json:"assignee_ids"`
	LabelIDs         []string                   `json:"label_ids"`
	DueDate          string                     `json:"due_date"`
	Relations        []WorkbenchIssueRelation   `json:"relations"`
	SubIssueProgress *WorkbenchSubIssueProgress `json:"sub_issue_progress,omitempty"`
	Source           *WorkbenchIssueSource      `json:"source,omitempty"`
	Pinned           bool                       `json:"pinned"`
	Locked           bool                       `json:"locked"`
	CommentCount     int64                      `json:"comment_count"`
	CreatedBy        string                     `json:"created_by"`
	CreatedAt        string                     `json:"created_at"`
	UpdatedAt        string                     `json:"updated_at"`
	DoneAt           *string                    `json:"done_at"`
}

// WorkbenchIssueRelation is an edge to another issue of the same workspace.
type WorkbenchIssueRelation struct {
	// Type is parent, sub_issue, blocked_by, blocking or duplicate_of.
	Type    string `json:"type"`
	Number  int64  `json:"number"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Deleted bool   `json:"deleted"`
}

// WorkbenchSubIssueProgress counts an issue's sub-issues.
type WorkbenchSubIssueProgress struct {
	Done  int64 `json:"done"`
	Total int64 `json:"total"`
}

// WorkbenchIssueSource is where a system-opened feedback issue came from.
type WorkbenchIssueSource struct {
	Product    string `json:"product"`
	Project    string `json:"project"`
	ChannelID  string `json:"channel_id"`
	MessageID  string `json:"message_id"`
	ReporterID string `json:"reporter_id"`
}

// WorkbenchEvent is one system event on an issue's timeline.
type WorkbenchEvent struct {
	ID           string  `json:"id"`
	At           string  `json:"at"`
	ActorID      string  `json:"actor_id"`
	ViaAssistant bool    `json:"via_assistant"`
	BatchID      string  `json:"batch_id"`
	Kind         string  `json:"kind"`
	Field        string  `json:"field,omitempty"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	FromDisplay  *string `json:"from_display"`
	ToDisplay    *string `json:"to_display"`
}

// WorkbenchComment is one comment. Comments are not events; a timeline merges
// the two by time.
type WorkbenchComment struct {
	ID           string  `json:"id"`
	IssueNumber  int64   `json:"issue_number"`
	AuthorID     string  `json:"author_id"`
	Body         string  `json:"body"`
	BatchID      string  `json:"batch_id"`
	ViaAssistant bool    `json:"via_assistant"`
	CreatedAt    string  `json:"created_at"`
	EditedAt     *string `json:"edited_at"`
}

// WorkbenchLabel is one of the workspace's labels. There is one kind: every
// label is the workspace's own, `blocked` and `not planned` included.
type WorkbenchLabel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	IssueCount int64  `json:"issue_count"`
}

// WorkbenchMember is a workspace member, as the assignee picker sees one.
type WorkbenchMember struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	// IsMember is false for somebody who has left the workspace; the name is
	// still there, because the timeline still names them. Only set by
	// GetWorkbenchMembers.
	IsMember bool `json:"is_member"`
}

// WorkbenchAttachment is one file attached to an issue, with its provenance
// under the names `asgard-cli reference add` uses.
type WorkbenchAttachment struct {
	ID               string `json:"id"`
	IssueNumber      int64  `json:"issue_number"`
	OriginalFilename string `json:"original_filename"`
	ByteSize         int64  `json:"byte_size"`
	SHA256           string `json:"sha256"`
	What             string `json:"what"`
	From             string `json:"from"`
	Dated            string `json:"dated"`
	Supersedes       string `json:"supersedes_attachment_id"`
	UploaderID       string `json:"uploader_id"`
	UploadedAt       string `json:"uploaded_at"`
	ViaAssistant     bool   `json:"via_assistant"`
}

// WorkbenchIssueFilter is the list's query. Values within one field are ORed,
// fields are ANDed; a nil field does not filter.
type WorkbenchIssueFilter struct {
	Status   []string
	Type     []string
	Priority []string
	// Pipeline holds pipeline ids; "none" selects issues about none.
	Pipeline []string
	// Assignee holds user ids; "me" is the signed-in user.
	Assignee []string
	// Label holds label ids.
	Label []string
	Q     string
	Sort  string
}

func (f WorkbenchIssueFilter) query() url.Values {
	q := url.Values{}
	add := func(key string, values []string) {
		for _, v := range values {
			q.Add(key, v)
		}
	}
	add("status", f.Status)
	add("type", f.Type)
	add("priority", f.Priority)
	add("pipeline_id", f.Pipeline)
	add("assignee", f.Assignee)
	add("label", f.Label)
	if f.Q != "" {
		q.Set("q", f.Q)
	}
	if f.Sort != "" {
		q.Set("sort", f.Sort)
	}
	return q
}

// workbenchPageSize is the largest page the issue and comment lists serve.
const workbenchPageSize = 100

// ListWorkbenchIssues returns up to limit issues matching the filter, pinned
// first, and how many match in all. A limit of 0 reads every page.
func (c *Client) ListWorkbenchIssues(ctx context.Context, f WorkbenchIssueFilter, limit int) ([]*WorkbenchIssue, int64, error) {
	var out []*WorkbenchIssue
	var total int64
	token := ""
	for {
		q := f.query()
		size := workbenchPageSize
		if limit > 0 && limit-len(out) < size {
			size = limit - len(out)
		}
		q.Set("page_size", strconv.Itoa(size))
		if token != "" {
			q.Set("page_token", token)
		}
		var batch []*WorkbenchIssue
		var paging CursorPaging
		if err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/issues", query: q, out: &batch, cursor: &paging}); err != nil {
			return nil, 0, err
		}
		out = append(out, batch...)
		total = paging.TotalCount
		token = paging.NextPageToken
		if token == "" || len(batch) == 0 || (limit > 0 && len(out) >= limit) {
			return out, total, nil
		}
	}
}

// GetWorkbenchIssue returns one issue by its number, the N of ISS-N.
func (c *Client) GetWorkbenchIssue(ctx context.Context, number int64) (*WorkbenchIssue, error) {
	var out WorkbenchIssue
	err := c.do(ctx, request{method: http.MethodGet, path: issuePath(number), out: &out})
	return &out, err
}

// CreateWorkbenchIssueRequest is the body of a create.
type CreateWorkbenchIssueRequest struct {
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Body        string   `json:"body,omitempty"`
	Status      string   `json:"status,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	PipelineID  string   `json:"pipeline_id,omitempty"`
	AssigneeIDs []string `json:"assignee_ids,omitempty"`
	LabelIDs    []string `json:"label_ids,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
}

// CreateWorkbenchIssue opens an issue, as the member's assistant.
func (c *Client) CreateWorkbenchIssue(ctx context.Context, body CreateWorkbenchIssueRequest) (*WorkbenchIssue, error) {
	var out WorkbenchIssue
	err := c.do(ctx, request{method: http.MethodPost, path: "/v1/workbench/issues", body: body, out: &out, viaAssistant: true})
	return &out, err
}

// UpdateWorkbenchIssueRequest is the body of an update. **Every field is a
// pointer or a list because omitted means unchanged**, and the zero value of
// several is a meaningful change: an empty due date removes it, an empty
// pipeline detaches the issue, and parent 0 clears the parent.
type UpdateWorkbenchIssueRequest struct {
	Title      *string `json:"title,omitempty"`
	Body       *string `json:"body,omitempty"`
	Type       *string `json:"type,omitempty"`
	Status     *string `json:"status,omitempty"`
	Priority   *string `json:"priority,omitempty"`
	DueDate    *string `json:"due_date,omitempty"`
	PipelineID *string `json:"pipeline_id,omitempty"`

	AddAssigneeIDs    []string `json:"add_assignee_ids,omitempty"`
	RemoveAssigneeIDs []string `json:"remove_assignee_ids,omitempty"`
	AddLabelIDs       []string `json:"add_label_ids,omitempty"`
	RemoveLabelIDs    []string `json:"remove_label_ids,omitempty"`

	ParentNumber           *int64  `json:"parent_number,omitempty"`
	DuplicateOfNumber      *int64  `json:"duplicate_of_number,omitempty"`
	AddBlockedByNumbers    []int64 `json:"add_blocked_by_numbers,omitempty"`
	RemoveBlockedByNumbers []int64 `json:"remove_blocked_by_numbers,omitempty"`
	AddSubIssueNumbers     []int64 `json:"add_sub_issue_numbers,omitempty"`
	RemoveSubIssueNumbers  []int64 `json:"remove_sub_issue_numbers,omitempty"`
}

// UpdateWorkbenchIssue changes the fields set in body, in one batch, as the
// member's assistant.
func (c *Client) UpdateWorkbenchIssue(ctx context.Context, number int64, body UpdateWorkbenchIssueRequest) (*WorkbenchIssue, error) {
	var out WorkbenchIssue
	err := c.do(ctx, request{method: http.MethodPatch, path: issuePath(number), body: body, out: &out, viaAssistant: true})
	return &out, err
}

// ListWorkbenchEvents returns an issue's system events, oldest first.
func (c *Client) ListWorkbenchEvents(ctx context.Context, number int64) ([]*WorkbenchEvent, error) {
	var out []*WorkbenchEvent
	err := c.do(ctx, request{method: http.MethodGet, path: issuePath(number) + "/events", out: &out})
	return out, err
}

// ListWorkbenchComments returns every comment on an issue, oldest first.
func (c *Client) ListWorkbenchComments(ctx context.Context, number int64) ([]*WorkbenchComment, error) {
	var out []*WorkbenchComment
	token := ""
	for {
		q := url.Values{"page_size": {strconv.Itoa(workbenchPageSize)}}
		if token != "" {
			q.Set("page_token", token)
		}
		var batch []*WorkbenchComment
		var paging CursorPaging
		if err := c.do(ctx, request{method: http.MethodGet, path: issuePath(number) + "/comments", query: q, out: &batch, cursor: &paging}); err != nil {
			return nil, err
		}
		out = append(out, batch...)
		if paging.NextPageToken == "" || len(batch) == 0 {
			return out, nil
		}
		token = paging.NextPageToken
	}
}

// ListWorkbenchLabels returns every label of the workspace.
func (c *Client) ListWorkbenchLabels(ctx context.Context) ([]*WorkbenchLabel, error) {
	var out []*WorkbenchLabel
	err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/labels", out: &out})
	return out, err
}

// SearchWorkbenchMembers returns up to six members whose name or email
// contains q.
func (c *Client) SearchWorkbenchMembers(ctx context.Context, q string) ([]*WorkbenchMember, error) {
	var out []*WorkbenchMember
	err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/members", query: url.Values{"q": {q}}, out: &out})
	return out, err
}

// membersPerCall is the most ids the lookup takes at once.
const membersPerCall = 50

// GetWorkbenchMembers looks members up by user id, including those who have
// left. An id the platform does not recognise comes back with no name.
func (c *Client) GetWorkbenchMembers(ctx context.Context, ids []string) ([]*WorkbenchMember, error) {
	var out []*WorkbenchMember
	for start := 0; start < len(ids); start += membersPerCall {
		end := min(start+membersPerCall, len(ids))
		var batch []*WorkbenchMember
		q := url.Values{"ids": {strings.Join(ids[start:end], ",")}}
		if err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/members", query: q, out: &batch}); err != nil {
			return nil, err
		}
		out = append(out, batch...)
	}
	return out, nil
}

// ListWorkbenchAttachments returns the attachments an issue still has, oldest
// first. A deleted one is not listed.
func (c *Client) ListWorkbenchAttachments(ctx context.Context, number int64) ([]*WorkbenchAttachment, error) {
	var out []*WorkbenchAttachment
	err := c.do(ctx, request{method: http.MethodGet, path: issuePath(number) + "/attachments", out: &out})
	return out, err
}

// ListWorkspaceAttachments returns the attachments that still exist on the
// named issues or on the issues about the named pipelines (pipeline ids;
// "none" is issues about none), across every page. A nil filter on both
// returns every attachment in the workspace.
func (c *Client) ListWorkspaceAttachments(ctx context.Context, pipelines []string, issues []int64) ([]*WorkbenchAttachment, error) {
	var out []*WorkbenchAttachment
	token := ""
	for {
		q := url.Values{"page_size": {strconv.Itoa(workbenchPageSize)}}
		for _, p := range pipelines {
			q.Add("pipeline_id", p)
		}
		for _, n := range issues {
			q.Add("issue", strconv.FormatInt(n, 10))
		}
		if token != "" {
			q.Set("page_token", token)
		}
		var batch []*WorkbenchAttachment
		var paging CursorPaging
		if err := c.do(ctx, request{method: http.MethodGet, path: "/v1/workbench/attachments", query: q, out: &batch, cursor: &paging}); err != nil {
			return nil, err
		}
		out = append(out, batch...)
		if paging.NextPageToken == "" || len(batch) == 0 {
			return out, nil
		}
		token = paging.NextPageToken
	}
}

// ErrChecksumMismatch is a download whose bytes do not hash to what the
// platform recorded.
var ErrChecksumMismatch = errors.New("checksum mismatch")

// DownloadWorkbenchAttachment writes an attachment's original bytes to w and
// returns their SHA-256. It fails with ErrChecksumMismatch when the bytes do
// not hash to what the platform sent in X-Attachment-Sha256 - in which case
// w has already received them, and the caller is the one to discard them.
func (c *Client) DownloadWorkbenchAttachment(ctx context.Context, a *WorkbenchAttachment, w io.Writer) (string, error) {
	path := issuePath(a.IssueNumber) + "/attachments/" + url.PathEscape(a.ID) + "/download"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.profile.PlatformAPI+path, nil)
	if err != nil {
		return "", fmt.Errorf("build the request: %w", err)
	}
	if c.workspace == "" {
		return "", errors.New("no workspace selected; pass --workspace, or record one with `asgard-cli workspace use <id>`")
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set(ClientHeader, clientValue())
	httpReq.Header.Set(WorkspaceHeader, c.workspace)

	// A download can outlast the per-call timeout on a slow link, and the
	// bytes are bounded by the platform's own upload limit.
	httpClient := *c.http
	httpClient.Timeout = 0
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		apiErr := &APIError{Status: resp.StatusCode, Method: http.MethodGet, Path: path, Message: truncate(string(raw), 200)}
		var env envelope
		if json.Unmarshal(raw, &env) == nil {
			apiErr.Message, apiErr.ReasonCode, apiErr.ErrorCode = env.Message, env.ReasonCode, errorCode(env.Details)
		}
		return "", apiErr
	}

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(w, h), resp.Body); err != nil {
		return "", fmt.Errorf("read %s: %w", a.OriginalFilename, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	for _, want := range []string{resp.Header.Get("X-Attachment-Sha256"), a.SHA256} {
		if want != "" && !strings.EqualFold(want, got) {
			return got, fmt.Errorf("%w: %s hashes to %s and the platform recorded %s", ErrChecksumMismatch, a.OriginalFilename, got, want)
		}
	}
	return got, nil
}

func issuePath(number int64) string {
	return "/v1/workbench/issues/" + strconv.FormatInt(number, 10)
}
