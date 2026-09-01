package work

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scaffolded builds the three files a scaffolded repo has, in the shape the
// templates write them. Every writer here has to work against that shape, not
// against a shape invented by the test.
func scaffolded(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		RequestIndex: "# Requests Index\n\n" +
			"High-level requirements and governance documents live here.\n\n" +
			"| Request ID | Title | Priority | Status | Spec |\n|---|---|---|---|---|\n\n" +
			"## Rules\n\n- Create request specs as `REQ-xxx-short-name.md`.\n",
		TaskIndex: "# Task Index\n\n## Config\n- Mode: spec\n\n" +
			"## Next Task\n\nNothing in flight. The next task usually opens once a request has a target project.\n\n" +
			"## Task Queue\n\n| Task ID | Title | Owner | Complexity | Status |\n" +
			"|---------|-------|-------|-----------|--------|\n\n" +
			"## Conventions\n\n- Complexity: `S` / `M` / `L`.\n",
		QuestionFile: "# Open questions\n\n## How to use it\n\n- Add a row the moment it blocks.\n\n" +
			"## Open\n\n| # | Question | What it blocks | Who can answer | Raised |\n|---|---|---|---|---|\n\n" +
			"## Known platform unknowns\n\n| # | Question | When it matters |\n|---|---|---|\n" +
			"| P1 | a platform unknown | any customer with two teams |\n\n" +
			"## Answered\n\n| # | Question | Answer | Decision record | Answered |\n|---|---|---|---|---|\n",
		DecisionTmpl: "# <topic>\n\n> decided on: YYYY-MM-DD\n> source: a meeting note dated YYYY-MM-DD\n\n## Summary\n",
		filepath.Join("docs", "spec", "acme-asgard", "README.md"): "# Living Spec\n\n## Traceability\n\n" +
			TraceabilityAnchor + "\n\n| decision | date | module |\n|---|---|---|\n",
	}
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	return dir
}

func read(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

// TestRequestRoundTrip is the test that matters most: what AddRequest writes has
// to be what ReadRequests can read back. They are a writer and a parser of the
// same markdown table, and nothing else in the repo would catch them drifting.
func TestRequestRoundTrip(t *testing.T) {
	dir := scaffolded(t)

	first, err := AddRequest(dir, Request{
		Title: "warehouse staff need stock in chat", Slug: "stock-in-chat", Raised: "2026-09-01",
	})
	if err != nil {
		t.Fatalf("AddRequest: %v", err)
	}
	if first.ID != "REQ-001" {
		t.Errorf("first ID = %q, want REQ-001", first.ID)
	}

	second, err := AddRequest(dir, Request{
		Title: "visitors ask about products", Slug: "public-catalog",
		Project: "site", Raised: "2026-09-02",
	})
	if err != nil {
		t.Fatalf("AddRequest: %v", err)
	}
	if second.ID != "REQ-002" {
		t.Errorf("second ID = %q, want REQ-002", second.ID)
	}

	requests, err := ReadRequests(dir)
	if err != nil {
		t.Fatalf("ReadRequests: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("read %d requests, want 2: %+v", len(requests), requests)
	}
	if requests[0].Title != "warehouse staff need stock in chat" {
		t.Errorf("title = %q", requests[0].Title)
	}
	if requests[0].Status != Draft {
		t.Errorf("status = %q, want draft", requests[0].Status)
	}
	if requests[0].Project != "" {
		t.Errorf("project = %q, want empty until the audience is decided", requests[0].Project)
	}
	if requests[1].Project != "site" {
		t.Errorf("project = %q, want site", requests[1].Project)
	}

	// The row goes inside the table, not after the Rules heading below it.
	index := read(t, dir, RequestIndex)
	if strings.Index(index, "REQ-002") > strings.Index(index, "## Rules") {
		t.Errorf("rows were appended past the end of the table:\n%s", index)
	}

	spec := read(t, dir, second.File())
	for _, want := range []string{"# REQ-002 - visitors ask about products", "- Status: `draft`", "- Raised: 2026-09-02", "- Target project: site"} {
		if !strings.Contains(spec, want) {
			t.Errorf("spec is missing %q:\n%s", want, spec)
		}
	}
}

func TestAddRequestRefusesToOverwrite(t *testing.T) {
	dir := scaffolded(t)

	r := Request{ID: "REQ-001", Title: "a thing", Slug: "a-thing", Raised: "2026-09-01"}
	if _, err := AddRequest(dir, r); err != nil {
		t.Fatalf("AddRequest: %v", err)
	}
	// An explicit ID that already has a file is somebody's writing; the whole
	// point of the record is that it is not regenerated over.
	if _, err := AddRequest(dir, r); err == nil {
		t.Fatal("AddRequest overwrote an existing spec")
	}
}

// TestSetTaskStatusMovesEveryPlaceAtOnce is the defect this package exists to
// remove: the index, the spec's Meta and the spec's log were three separate
// edits, and a repo where two of them disagreed gave no way to tell which was
// current.
func TestSetTaskStatusMovesEveryPlaceAtOnce(t *testing.T) {
	dir := scaffolded(t)

	task, err := AddTask(dir, Task{
		Title: "expose stock levels", Slug: "stock-levels",
		Project: "erp", Request: "REQ-001", Complexity: "M", Created: "2026-09-01",
	})
	if err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if task.ID != "TASK-001" {
		t.Errorf("ID = %q, want TASK-001", task.ID)
	}

	if err := SetTaskStatus(dir, "TASK-001", Ready, "2026-09-03"); err != nil {
		t.Fatalf("SetTaskStatus: %v", err)
	}

	index := read(t, dir, TaskIndex)
	if !strings.Contains(index, "`ready`") {
		t.Errorf("index still does not say ready:\n%s", index)
	}
	if !strings.Contains(index, "TASK-001 expose stock levels - `ready`") {
		t.Errorf("Next Task section was not refreshed:\n%s", index)
	}

	spec := read(t, dir, task.File())
	if !strings.Contains(spec, "- Status: `ready`") {
		t.Errorf("spec Meta still does not say ready:\n%s", spec)
	}
	if !strings.Contains(spec, "- 2026-09-03 status `draft` -> `ready`") {
		t.Errorf("the transition was not logged with its date:\n%s", spec)
	}

	tasks, err := ReadTasks(dir)
	if err != nil {
		t.Fatalf("ReadTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Status != Ready {
		t.Errorf("tasks = %+v, want one at ready", tasks)
	}
}

func TestSetTaskStatusRejectsWhatItCannotDo(t *testing.T) {
	dir := scaffolded(t)
	if _, err := AddTask(dir, Task{Title: "a thing", Slug: "a-thing", Created: "2026-09-01"}); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if err := SetTaskStatus(dir, "TASK-001", "in_progress", "2026-09-03"); err == nil {
		t.Error("the underscore spelling should be rejected; the flow has four symbols")
	}
	if err := SetTaskStatus(dir, "TASK-009", Ready, "2026-09-03"); err == nil {
		t.Error("an unknown ID should be reported, not silently ignored")
	}
	if err := SetTaskStatus(dir, "TASK-001", Draft, "2026-09-03"); err == nil {
		t.Error("a transition to the status it already has should be reported")
	}
}

// TestReadTasks uses the shape both reference repos write, including a status
// with a trailing note, which is how they record a superseded task.
func TestReadTasks(t *testing.T) {
	dir := t.TempDir()
	index := filepath.Join(dir, TaskIndex)
	if err := os.MkdirAll(filepath.Dir(index), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "# Task Index\n\n## Task Queue\n\n" +
		"| Task ID | Title | Owner | Complexity | Status |\n" +
		"|---------|-------|-------|-----------|--------|\n" +
		"| [TASK-001](TASK-001-split.md) | 依系統拆分 | — | L | `done` |\n" +
		"| [TASK-002](TASK-002-notify.md) | 到貨通知 | — | L | `draft` |\n" +
		"| [TASK-003](TASK-003-layer.md) | 官網語意層 | — | M | `superseded`(見 `TASK-010`) |\n" +
		"| [TASK-004](TASK-004-drive.md) | 知識 Drive | — | L | `in-progress` |\n"
	if err := os.WriteFile(index, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tasks, err := ReadTasks(dir)
	if err != nil {
		t.Fatalf("ReadTasks: %v", err)
	}
	if len(tasks) != 4 {
		t.Fatalf("parsed %d tasks, want 4: %+v", len(tasks), tasks)
	}
	if tasks[2].Status != "superseded" {
		t.Errorf("status = %q, want the bare word without its trailing note", tasks[2].Status)
	}
	if tasks[0].Title != "依系統拆分" {
		t.Errorf("title = %q", tasks[0].Title)
	}

	// superseded is not one of the four symbols, so it is not in flight either.
	active := ActiveTasks(tasks)
	if len(active) != 2 {
		t.Fatalf("active = %+v, want the draft and the in-progress one", active)
	}
	for _, want := range []string{"TASK-002", "TASK-004"} {
		found := false
		for _, a := range active {
			if a.ID == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s should be active", want)
		}
	}
}

func TestReadingAnUnscaffoldedRepoIsNotAnError(t *testing.T) {
	// Early in an onboarding none of these files exist, which is a state.
	dir := t.TempDir()

	requests, err := ReadRequests(dir)
	if err != nil || len(requests) != 0 {
		t.Errorf("ReadRequests = %+v, %v; want none and no error", requests, err)
	}
	tasks, err := ReadTasks(dir)
	if err != nil || len(tasks) != 0 {
		t.Errorf("ReadTasks = %+v, %v; want none and no error", tasks, err)
	}
	questions, err := ReadQuestions(dir)
	if err != nil || len(questions) != 0 {
		t.Errorf("ReadQuestions = %+v, %v; want none and no error", questions, err)
	}
}

func TestQuestionsAreNumberedAndAnsweredInPlace(t *testing.T) {
	dir := scaffolded(t)

	for i, text := range []string{"which stock figure is authoritative", "who owns the API credentials"} {
		q, err := AddQuestion(dir, Question{Text: text, Blocks: "REQ-001", Owner: "the warehouse lead", Raised: "2026-09-01"})
		if err != nil {
			t.Fatalf("AddQuestion: %v", err)
		}
		if want := []string{"1", "2"}[i]; q.Number != want {
			t.Errorf("number = %q, want %q", q.Number, want)
		}
	}

	questions, err := ReadQuestions(dir)
	if err != nil {
		t.Fatalf("ReadQuestions: %v", err)
	}
	// The platform-unknowns table has three columns and the Answered table has
	// six, so neither should be picked up as an open question.
	if len(questions) != 2 {
		t.Fatalf("read %d questions, want 2: %+v", len(questions), questions)
	}
	if questions[0].Raised != "2026-09-01" || questions[0].Owner != "the warehouse lead" {
		t.Errorf("question = %+v, want the date and the owner kept", questions[0])
	}

	if err := AnswerQuestion(dir, "1", "location 608 only", "2026-09-04-safety-stock.md", "2026-09-04"); err != nil {
		t.Fatalf("AnswerQuestion: %v", err)
	}

	open, err := ReadQuestions(dir)
	if err != nil {
		t.Fatalf("ReadQuestions: %v", err)
	}
	if len(open) != 1 || open[0].Number != "2" {
		t.Fatalf("open = %+v, want only question 2", open)
	}

	body := read(t, dir, QuestionFile)
	answered := body[strings.Index(body, "## Answered"):]
	// The row moves rather than being deleted: that it was once open is what
	// explains the shape of the design that answered it.
	if !strings.Contains(answered, "which stock figure is authoritative") {
		t.Errorf("the row did not reach the Answered table:\n%s", answered)
	}
	if !strings.Contains(answered, "location 608 only") {
		t.Errorf("the answer was not recorded:\n%s", answered)
	}

	if err := AnswerQuestion(dir, "9", "no", "", "2026-09-04"); err == nil {
		t.Error("answering a question that is not open should be reported")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"warehouse staff need stock in chat", "warehouse-staff-need-stock-in"},
		{"Let visitors ask about products!", "let-visitors-ask-about-products"},
		{"  spaces  and---dashes  ", "spaces-and-dashes"},
		// A title in the customer's own words is the normal case, and it has no
		// ASCII to derive a file name from. The caller asks for --slug rather
		// than writing a file nobody can name.
		{"客戶要能在官網問庫存", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Slugify(tt.in); got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}

	long := Slugify(strings.Repeat("a-very-long-title ", 10))
	if len(long) > 32 {
		t.Errorf("Slugify produced a %d character slug: %q", len(long), long)
	}
	if strings.HasSuffix(long, "-") {
		t.Errorf("a truncated slug should not end in a hyphen: %q", long)
	}
}

func TestStatusFlow(t *testing.T) {
	for _, s := range Statuses {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}
	for _, s := range []Status{"in_progress", "superseded", "", "DONE"} {
		if Status(s).Valid() {
			t.Errorf("%q should not be valid", s)
		}
		if Status(s).Active() {
			t.Errorf("%q should not count as in flight", s)
		}
	}
	if Done.Active() {
		t.Error("done is the one status that is not in flight")
	}
	for _, s := range []Status{Draft, Ready, InProgress} {
		if !s.Active() {
			t.Errorf("%q is work the repo has recorded and not finished", s)
		}
	}
}

// TestSetRequestProjectIsWhatMovesTheInterviewOn pins the two places that used
// to be a hand edit each: the registry cell the stage machinery reads, and the
// Meta line a person reads.
func TestSetRequestProjectIsWhatMovesTheInterviewOn(t *testing.T) {
	dir := scaffolded(t)

	r, err := AddRequest(dir, Request{Title: "stock in chat", Slug: "stock", Raised: "2026-09-01"})
	if err != nil {
		t.Fatalf("AddRequest: %v", err)
	}

	if err := SetRequestProject(dir, r.ID, "erp", "2026-09-03"); err != nil {
		t.Fatalf("SetRequestProject: %v", err)
	}

	requests, err := ReadRequests(dir)
	if err != nil {
		t.Fatalf("ReadRequests: %v", err)
	}
	if len(requests) != 1 || requests[0].Project != "erp" {
		t.Fatalf("requests = %+v, want one targeting erp", requests)
	}
	// The other cells survive: they may have been edited by hand.
	if requests[0].Title != "stock in chat" || requests[0].Status != Draft {
		t.Errorf("the rest of the row changed: %+v", requests[0])
	}

	spec := read(t, dir, r.File())
	if !strings.Contains(spec, "- Target project: erp") {
		t.Errorf("spec Meta was not updated:\n%s", spec)
	}
	if !strings.Contains(spec, "- 2026-09-03 target project set to `erp`") {
		t.Errorf("the change was not logged with its date:\n%s", spec)
	}

	if err := SetRequestProject(dir, "REQ-009", "erp", "2026-09-03"); err == nil {
		t.Error("an unknown request should be reported")
	}
}

func TestSetRequestStatusMovesTheRegistryAndTheSpec(t *testing.T) {
	dir := scaffolded(t)

	r, err := AddRequest(dir, Request{Title: "stock in chat", Slug: "stock", Raised: "2026-09-01"})
	if err != nil {
		t.Fatalf("AddRequest: %v", err)
	}

	if err := SetRequestStatus(dir, r.ID, Ready, "2026-09-03"); err != nil {
		t.Fatalf("SetRequestStatus: %v", err)
	}

	requests, err := ReadRequests(dir)
	if err != nil {
		t.Fatalf("ReadRequests: %v", err)
	}
	if len(requests) != 1 || requests[0].Status != Ready {
		t.Fatalf("requests = %+v, want one at ready", requests)
	}
	if len(ActiveRequests(requests)) != 1 {
		t.Error("a ready request is still in flight")
	}

	spec := read(t, dir, r.File())
	if !strings.Contains(spec, "- Status: `ready`") {
		t.Errorf("spec Meta was not updated:\n%s", spec)
	}
	if !strings.Contains(spec, "- 2026-09-03 status `draft` -> `ready`") {
		t.Errorf("the transition was not logged:\n%s", spec)
	}

	if err := SetRequestStatus(dir, r.ID, Done, "2026-09-04"); err != nil {
		t.Fatalf("SetRequestStatus: %v", err)
	}
	requests, err = ReadRequests(dir)
	if err != nil {
		t.Fatalf("ReadRequests: %v", err)
	}
	if len(ActiveRequests(requests)) != 0 {
		t.Error("a done request is not in flight")
	}

	if err := SetRequestStatus(dir, "REQ-009", Ready, "2026-09-04"); err == nil {
		t.Error("an unknown request should be reported")
	}
	if err := SetRequestStatus(dir, r.ID, "shipped", "2026-09-04"); err == nil {
		t.Error("a status outside the flow should be rejected")
	}
}

func TestAddDecisionUsesTheRepositorysOwnTemplate(t *testing.T) {
	dir := scaffolded(t)

	path, err := AddDecision(dir, "the work splits by audience", "project-split",
		"acme-asgard", "architecture.md", "2026-09-03")
	if err != nil {
		t.Fatalf("AddDecision: %v", err)
	}
	if path != filepath.Join(DecisionDir, "2026-09-03-project-split.md") {
		t.Errorf("path = %q, want the date in the file name", path)
	}

	record := read(t, dir, path)
	if !strings.Contains(record, "# the work splits by audience") {
		t.Errorf("the topic did not reach the heading:\n%s", record)
	}
	if !strings.Contains(record, "> decided on: 2026-09-03") {
		t.Errorf("the decided-on date was not stamped:\n%s", record)
	}
	// Only the first placeholder is the decision's own date; the source link's
	// date is still unknown and stays a placeholder.
	if !strings.Contains(record, "source: a meeting note dated YYYY-MM-DD") {
		t.Errorf("the source placeholder should be left alone:\n%s", record)
	}

	index := read(t, dir, filepath.Join("docs", "spec", "acme-asgard", "README.md"))
	if !strings.Contains(index, "2026-09-03-project-split.md") {
		t.Errorf("the traceability table was not updated:\n%s", index)
	}

	if _, err := AddDecision(dir, "again", "project-split", "acme-asgard", "", "2026-09-03"); err == nil {
		t.Error("a decision record is immutable; writing over one should be refused")
	}

	// A repo scaffolded before the anchor existed still gets its record, and is
	// told what could not be linked - failing after the file is on disk would
	// leave the caller guessing.
	if err := os.WriteFile(filepath.Join(dir, "docs", "spec", "acme-asgard", "README.md"),
		[]byte("# Living Spec\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	path, err = AddDecision(dir, "a second one", "second", "acme-asgard", "", "2026-09-04")
	if path == "" {
		t.Errorf("the record should still be written: %v", err)
	}
	if err == nil {
		t.Error("the missing anchor should be reported")
	}
}

func TestAddDecisionNeedsTheTemplate(t *testing.T) {
	// An unscaffolded repo has no template, and inventing one here would put a
	// second copy of the format in the CLI.
	if _, err := AddDecision(t.TempDir(), "a topic", "a-topic", "acme-asgard", "", "2026-09-03"); err == nil {
		t.Fatal("AddDecision should report the missing template")
	}
}
