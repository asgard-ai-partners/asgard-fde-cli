// Package work reads and writes the records a customer repository keeps of its
// own work: requests, task specs, and open questions.
//
// The state belongs to the repository, never to this CLI. A request is a file
// under requirements/requests/, a task is a file under requirements/tasks/, a
// question is a row in docs/open-questions.md, and an agent opening the repo
// reads exactly those. Nothing here is remembered anywhere else.
//
// What the CLI adds is that it stamps the date and moves a status in the spec
// and in the index together. Both were previously a person's job, done in three
// places by hand, which is why a repo could say `draft` in one file and
// `in-progress` in another and neither was wrong on its face.
package work

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

// Status is a point in the SDD status flow. These four symbols are the whole
// vocabulary, and docs/spec-driven-development.md in the generated repo says so
// too: in_progress with an underscore is not one of them.
type Status string

const (
	Draft      Status = "draft"
	Ready      Status = "ready"
	InProgress Status = "in-progress"
	Done       Status = "done"
)

// Statuses lists the flow in order.
var Statuses = []Status{Draft, Ready, InProgress, Done}

// Valid reports whether s is one of the four.
func (s Status) Valid() bool {
	return slices.Contains(Statuses, s)
}

// Active reports whether work at this status is still in flight. `done` is the
// only status that is not: a `draft` nobody is working on is still an obligation
// the repo has recorded.
func (s Status) Active() bool { return s.Valid() && s != Done }

// StatusList names the four for an error message.
func StatusList() string {
	names := make([]string, len(Statuses))
	for i, s := range Statuses {
		names[i] = string(s)
	}
	return strings.Join(names, ", ")
}

// Paths of the files this package owns, relative to the repository root.
var (
	RequestDir   = filepath.Join("requirements", "requests")
	TaskDir      = filepath.Join("requirements", "tasks")
	RequestIndex = filepath.Join(RequestDir, "_index.md")
	TaskIndex    = filepath.Join(TaskDir, "_index.md")
	QuestionFile = filepath.Join("docs", "open-questions.md")
	DecisionDir  = filepath.Join("docs", "decisions")
	DecisionTmpl = filepath.Join(DecisionDir, "_decision-template.md")
	ReferenceDir = "references"
)

// FiledReferences counts the customer material sitting in references/, ignoring
// the README the scaffold puts there. It exists to catch the one state nothing
// else in the repository can see: material has been filed and read, an
// interview has effectively happened, and no request records any of it.
//
// That state is not a missing file - it is a conversation that only exists in
// somebody's terminal. `check` and the interview prompt both report it, because
// the next run of `next` will otherwise say "nothing in flight" and be right.
func FiledReferences(root string) (int, error) {
	dir := filepath.Join(root, ReferenceDir)
	var n int
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == dir {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch name := d.Name(); {
		case name == "README.md", strings.HasPrefix(name, "_"), strings.HasPrefix(name, "."):
			return nil
		}
		n++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", ReferenceDir, err)
	}
	return n, nil
}

// TraceabilityAnchor marks the table in the living spec's README that lists
// every decision record. It is an HTML comment rather than the heading above the
// table because the generated repo writes its own documents in the customer's
// language, and an anchor that has to be matched literally should not depend on
// which language that is.
const TraceabilityAnchor = "<!-- traceability -->"

// Request is one thing the customer asked for. It is the unit of work an
// engagement actually receives: a stage describes how far a request has got,
// not how far the repository has got.
type Request struct {
	ID       string // REQ-001
	Title    string
	Slug     string // short name in the file name
	Status   Status
	Priority string
	Audience string // internal authenticated, or public anonymous
	Project  string // target project slug; empty until the split is decided
	Raised   string // YYYY-MM-DD
	SpecSlug string // living spec directory, for the links in the generated file
}

// File is the request's path relative to the repository root.
func (r Request) File() string {
	return filepath.Join(RequestDir, r.ID+"-"+r.Slug+".md")
}

// Task is one executable task spec.
type Task struct {
	ID         string // TASK-001
	Title      string
	Slug       string
	Status     Status
	Owner      string
	Complexity string // S, M or L
	Project    string
	Request    string // the REQ this implements, if any
	Created    string // YYYY-MM-DD
	SpecSlug   string
}

// File is the task's path relative to the repository root.
func (t Task) File() string {
	return filepath.Join(TaskDir, t.ID+"-"+t.Slug+".md")
}

// Question is one thing nobody has answered yet.
//
// An unanswered question has nowhere else to live: a decision record is for
// something settled, a task spec's open questions vanish when the task reaches
// done, and the living spec describes what is, not what nobody knows.
type Question struct {
	Number string
	Text   string
	Blocks string
	Owner  string
	Raised string // YYYY-MM-DD
}

// Active reports whether the request is still in flight.
func (r Request) Active() bool { return r.Status.Active() }

// Active reports whether the task is still in flight.
func (t Task) Active() bool { return t.Status.Active() }

// requestRow matches a row of the request registry. The ID may be a plain
// string or a markdown link, and the status is written in backticks.
var requestRow = regexp.MustCompile(`(?m)^\|\s*\[?(REQ-\d+)\]?[^|]*\|\s*(.*?)\s*\|\s*([^|]*?)\s*\|\s*` + "`" + `([a-z-]+)` + "`" + `[^|]*\|\s*([^|]*?)\s*\|`)

// taskRow matches a row of the task queue. The status may carry a trailing
// note, as in `superseded`(see TASK-010), which is why the status group stops at
// the closing backtick.
var taskRow = regexp.MustCompile(`(?m)^\|\s*\[?(TASK-\d+)\]?[^|]*\|\s*(.*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*` + "`" + `([a-z-]+)` + "`")

// questionRow matches a five-column row. The Answered table has six and the
// platform-unknowns table has three, so neither matches.
var questionRow = regexp.MustCompile(`(?m)^\|\s*(\d+)\s*\|\s*([^|]+?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*$`)

// ReadRequests parses the request registry. A repo without the file, or with an
// empty registry, has no requests - the normal state early on, not an error.
func ReadRequests(root string) ([]Request, error) {
	text, err := readOptional(filepath.Join(root, RequestIndex))
	if err != nil {
		return nil, err
	}

	var out []Request
	for _, m := range requestRow.FindAllStringSubmatch(text, -1) {
		out = append(out, Request{
			ID:       m[1],
			Title:    m[2],
			Priority: m[3],
			Status:   Status(m[4]),
			Project:  projectFromSpec(m[5]),
		})
	}
	return out, nil
}

// projectFromSpec reads the target project out of the registry's Spec column,
// which the CLI writes as "projects/<slug>" or as a task link. Anything else is
// left alone: a hand-written cell is not required to be machine-readable, and
// guessing would be worse than reporting nothing.
func projectFromSpec(cell string) string {
	cell = strings.TrimSpace(strings.Trim(cell, "`"))
	const prefix = "projects/"
	if !strings.HasPrefix(cell, prefix) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(cell, prefix), "/")
}

// ReadTasks parses the task queue.
func ReadTasks(root string) ([]Task, error) {
	text, err := readOptional(filepath.Join(root, TaskIndex))
	if err != nil {
		return nil, err
	}

	var out []Task
	for _, m := range taskRow.FindAllStringSubmatch(text, -1) {
		out = append(out, Task{
			ID:         m[1],
			Title:      m[2],
			Owner:      m[3],
			Complexity: m[4],
			Status:     Status(m[5]),
		})
	}
	return out, nil
}

// ReadQuestions parses the open half of docs/open-questions.md.
func ReadQuestions(root string) ([]Question, error) {
	text, err := readOptional(filepath.Join(root, QuestionFile))
	if err != nil {
		return nil, err
	}
	// Everything from the Answered heading on is history.
	if idx := strings.Index(text, "## Answered"); idx >= 0 {
		text = text[:idx]
	}

	var out []Question
	for _, m := range questionRow.FindAllStringSubmatch(text, -1) {
		out = append(out, Question{
			Number: m[1],
			Text:   m[2],
			Blocks: m[3],
			Owner:  m[4],
			Raised: m[5],
		})
	}
	return out, nil
}

// ActiveRequests returns the requests still in flight, in registry order.
func ActiveRequests(requests []Request) []Request {
	var out []Request
	for _, r := range requests {
		if r.Active() {
			out = append(out, r)
		}
	}
	return out
}

// ActiveTasks returns the tasks still in flight, in queue order.
func ActiveTasks(tasks []Task) []Task {
	var out []Task
	for _, t := range tasks {
		if t.Active() {
			out = append(out, t)
		}
	}
	return out
}

// AddRequest writes the request spec and registers it. The ID is assigned from
// what is already on disk when r.ID is empty, and the request returned carries
// the ID and file it ended up with.
func AddRequest(root string, r Request) (Request, error) {
	if r.Slug == "" {
		return r, fmt.Errorf("request needs a slug for its file name")
	}
	if r.ID == "" {
		id, err := nextID(filepath.Join(root, RequestDir), "REQ")
		if err != nil {
			return r, err
		}
		r.ID = id
	}
	if r.Status == "" {
		r.Status = Draft
	}
	if !r.Status.Valid() {
		return r, fmt.Errorf("unknown status %q; the flow is %s", r.Status, StatusList())
	}

	if err := writeNew(filepath.Join(root, r.File()), requestTemplate, r); err != nil {
		return r, err
	}

	spec := "TODO"
	if r.Project != "" {
		spec = "projects/" + r.Project
	}
	row := fmt.Sprintf("| [%s](%s) | %s | %s | `%s` | %s |",
		r.ID, r.ID+"-"+r.Slug+".md", r.Title, orDash(r.Priority), r.Status, spec)
	if err := appendRow(filepath.Join(root, RequestIndex), "", row); err != nil {
		return r, err
	}
	return r, nil
}

// AddTask writes the task spec, registers it, and points the index's Next Task
// section at it. The task returned carries the ID and file it ended up with.
func AddTask(root string, t Task) (Task, error) {
	if t.Slug == "" {
		return t, fmt.Errorf("task needs a slug for its file name")
	}
	if t.ID == "" {
		// Task IDs are global across projects: two branches numbering from
		// their own project is how a collision happens.
		id, err := nextID(filepath.Join(root, TaskDir), "TASK")
		if err != nil {
			return t, err
		}
		t.ID = id
	}
	if t.Status == "" {
		t.Status = Draft
	}
	if !t.Status.Valid() {
		return t, fmt.Errorf("unknown status %q; the flow is %s", t.Status, StatusList())
	}

	if err := writeNew(filepath.Join(root, t.File()), taskTemplate, t); err != nil {
		return t, err
	}

	row := fmt.Sprintf("| [%s](%s) | %s | %s | %s | `%s` |",
		t.ID, t.ID+"-"+t.Slug+".md", t.Title, orDash(t.Owner), orDash(t.Complexity), t.Status)
	if err := appendRow(filepath.Join(root, TaskIndex), "## Task Queue", row); err != nil {
		return t, err
	}
	if err := refreshNextTask(root); err != nil {
		return t, err
	}
	return t, nil
}

// SetRequestStatus moves a request's status in the registry and in its own spec
// file, and logs the transition with the date. Doing both in one call is the
// point: the two going out of step is the failure this replaces.
func SetRequestStatus(root, id string, to Status, date string) error {
	if !to.Valid() {
		return fmt.Errorf("unknown status %q; the flow is %s", to, StatusList())
	}

	requests, err := ReadRequests(root)
	if err != nil {
		return err
	}
	var from Status
	found := false
	for _, r := range requests {
		if r.ID == id {
			from, found = r.Status, true
			break
		}
	}
	if !found {
		return fmt.Errorf("no %s in %s", id, RequestIndex)
	}
	if from == to {
		return fmt.Errorf("%s is already `%s`", id, to)
	}

	if err := setIndexStatus(filepath.Join(root, RequestIndex), requestRow, id, to); err != nil {
		return err
	}
	return setSpecStatus(root, RequestDir, id, from, to, date, requestLogHeading)
}

// SetTaskStatus moves a task's status in the queue and in its own spec file,
// logs the transition, and refreshes the index's Next Task section.
func SetTaskStatus(root, id string, to Status, date string) error {
	if !to.Valid() {
		return fmt.Errorf("unknown status %q; the flow is %s", to, StatusList())
	}

	tasks, err := ReadTasks(root)
	if err != nil {
		return err
	}
	var from Status
	found := false
	for _, t := range tasks {
		if t.ID == id {
			from, found = t.Status, true
			break
		}
	}
	if !found {
		return fmt.Errorf("no %s in %s", id, TaskIndex)
	}
	if from == to {
		return fmt.Errorf("%s is already `%s`", id, to)
	}

	if err := setIndexStatus(filepath.Join(root, TaskIndex), taskRow, id, to); err != nil {
		return err
	}
	if err := setSpecStatus(root, TaskDir, id, from, to, date, taskLogHeading); err != nil {
		return err
	}
	return refreshNextTask(root)
}

// SetRequestProject records the target project on a request: the Spec column of
// the registry, the Target project line of its Meta, and a dated line in its
// log. It is what moves the interview stage on, so it is one command rather
// than two edits somebody has to remember to keep together.
func SetRequestProject(root, id, project, date string) error {
	requests, err := ReadRequests(root)
	if err != nil {
		return err
	}
	found := false
	for _, r := range requests {
		if r.ID == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no %s in %s", id, RequestIndex)
	}

	// Column 4 of "| Request ID | Title | Priority | Status | Spec |".
	if err := setCell(filepath.Join(root, RequestIndex), requestRow, id, 4, "projects/"+project); err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(root, RequestDir, id+"-*.md"))
	if err != nil {
		return fmt.Errorf("look for %s under %s: %w", id, RequestDir, err)
	}
	target := regexp.MustCompile(`(?m)^- Target project:.*$`)
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		text := target.ReplaceAllString(string(data), "- Target project: "+project)
		text, err = appendUnderHeading(text, requestLogHeading, fmt.Sprintf("- %s target project set to `%s`", orDash(date), project))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// AddQuestion appends a row to the Open table, numbered after the highest
// number already there.
func AddQuestion(root string, q Question) (Question, error) {
	if strings.TrimSpace(q.Text) == "" {
		return q, fmt.Errorf("a question needs text")
	}

	existing, err := ReadQuestions(root)
	if err != nil {
		return q, err
	}
	if q.Number == "" {
		highest := 0
		for _, e := range existing {
			if n, err := strconv.Atoi(e.Number); err == nil && n > highest {
				highest = n
			}
		}
		q.Number = strconv.Itoa(highest + 1)
	}

	row := fmt.Sprintf("| %s | %s | %s | %s | %s |",
		q.Number, q.Text, orDash(q.Blocks), orDash(q.Owner), orDash(q.Raised))
	if err := appendRow(filepath.Join(root, QuestionFile), "## Open", row); err != nil {
		return q, err
	}
	return q, nil
}

// AnswerQuestion moves a row out of Open and into Answered, keeping the row
// rather than deleting it: that a question was once open is what explains the
// shape of the design that answered it.
func AnswerQuestion(root, number, answer, decision, date string) error {
	path := filepath.Join(root, QuestionFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	text := string(data)

	answered := strings.Index(text, "## Answered")
	if answered < 0 {
		return fmt.Errorf("%s has no Answered section", QuestionFile)
	}

	var row Question
	found := false
	lines := strings.Split(text[:answered], "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		m := questionRow.FindStringSubmatch(line)
		if m != nil && m[1] == number {
			// Only the number and the text carry over: the Answered table
			// records what the answer was, not who was going to give it.
			row = Question{Number: m[1], Text: m[2]}
			found = true
			continue
		}
		kept = append(kept, line)
	}
	if !found {
		return fmt.Errorf("no open question numbered %s in %s", number, QuestionFile)
	}

	text = strings.Join(kept, "\n") + text[answered:]
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	moved := fmt.Sprintf("| %s | %s | %s | %s | %s |",
		row.Number, row.Text, answer, orDash(decision), orDash(date))
	return appendRow(path, "## Answered", moved)
}

// AddDecision writes a dated decision record from the repository's own
// template, and links it from the living spec's traceability table.
//
// The template is read from the repo rather than embedded here because the
// engagement is allowed to change it, and a copy in the CLI would quietly
// override that. It returns the path written, relative to root.
func AddDecision(root, topic, slug, specSlug, module, date string) (string, error) {
	if slug == "" {
		return "", fmt.Errorf("decision needs a slug for its file name")
	}

	data, err := os.ReadFile(filepath.Join(root, DecisionTmpl))
	if os.IsNotExist(err) {
		return "", fmt.Errorf("%s is missing; run `asgard-cli scaffold` first", DecisionTmpl)
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", DecisionTmpl, err)
	}

	text := string(data)
	// The first heading is the topic placeholder, and the first date placeholder
	// is the one on the "decided on" line. The later ones belong to the source
	// links, which stay placeholders because nobody knows them yet.
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			lines[i] = "# " + topic
			break
		}
	}
	text = strings.Replace(strings.Join(lines, "\n"), "YYYY-MM-DD", date, 1)

	path := filepath.Join(DecisionDir, date+"-"+slug+".md")
	full := filepath.Join(root, path)
	if _, err := os.Stat(full); err == nil {
		return "", fmt.Errorf("%s already exists", path)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}

	row := fmt.Sprintf("| [%s](../../decisions/%s) | %s | %s |",
		topic, date+"-"+slug+".md", date, orDash(module))
	index := filepath.Join(root, "docs", "spec", specSlug, "README.md")
	if err := appendRow(index, TraceabilityAnchor, row); err != nil {
		// The record is written and is the thing that matters; a spec README
		// without the anchor is an older scaffold, and saying so is better than
		// failing after the file is already on disk.
		return path, fmt.Errorf("wrote the record, but could not link it from %s: %w",
			filepath.Join("docs", "spec", specSlug, "README.md"), err)
	}
	return path, nil
}

// nextID returns the first unused ID with this prefix, read from the file names
// in dir. File names are the record rather than the index table, because a spec
// written without its row still owns its number.
func nextID(dir, prefix string) (string, error) {
	pattern := regexp.MustCompile(`^` + prefix + `-(\d+)`)

	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read %s: %w", dir, err)
	}

	highest := 0
	for _, e := range entries {
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, err := strconv.Atoi(m[1]); err == nil && n > highest {
			highest = n
		}
	}
	return fmt.Sprintf("%s-%03d", prefix, highest+1), nil
}

// writeNew renders tmpl into path, refusing to overwrite. A spec file is
// somebody's writing; regenerating over it would lose the part that mattered.
func writeNew(path, tmpl string, data any) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	t, err := template.New(filepath.Base(path)).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse template for %s: %w", path, err)
	}
	var out strings.Builder
	if err := t.Execute(&out, data); err != nil {
		return fmt.Errorf("render %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// appendRow inserts row at the end of the first markdown table in the named
// section. An empty section means the first table in the file.
func appendRow(path, section, row string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")

	start := 0
	if section != "" {
		start = -1
		for i, line := range lines {
			if strings.TrimSpace(line) == section {
				start = i + 1
				break
			}
		}
		if start < 0 {
			return fmt.Errorf("%s has no %q section", path, section)
		}
	}

	// The separator row is what identifies a table: a header line alone could
	// be any line starting with a pipe.
	sep := -1
	for i := start; i < len(lines); i++ {
		if section != "" && strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			break
		}
		if isTableSeparator(lines[i]) {
			sep = i
			break
		}
	}
	if sep < 0 {
		return fmt.Errorf("%s has no table in %s", path, orDash(section))
	}

	end := sep + 1
	for end < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[end]), "|") {
		end++
	}

	updated := make([]string, 0, len(lines)+1)
	updated = append(updated, lines[:end]...)
	updated = append(updated, row)
	updated = append(updated, lines[end:]...)

	if err := os.WriteFile(path, []byte(strings.Join(updated, "\n")), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// isTableSeparator reports whether the line is a markdown table's |---|---| row.
func isTableSeparator(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") || !strings.Contains(line, "-") {
		return false
	}
	return strings.Trim(line, "|-: \t") == ""
}

// setIndexStatus rewrites the status cell of one row, leaving every other cell
// as it is: a row's other columns may have been edited by hand.
func setIndexStatus(path string, row *regexp.Regexp, id string, to Status) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")

	replaced := false
	statusCell := regexp.MustCompile("`[a-z-]+`")
	for i, line := range lines {
		m := row.FindStringSubmatch(line)
		if m == nil || m[1] != id {
			continue
		}
		lines[i] = statusCell.ReplaceAllString(line, "`"+string(to)+"`")
		replaced = true
		break
	}
	if !replaced {
		return fmt.Errorf("no row for %s in %s", id, path)
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// setCell rewrites one cell of the row whose first column is id, counting data
// cells from zero and leaving every other cell alone: the other columns may have
// been edited by hand, and this is not the place to reformat them.
func setCell(path string, row *regexp.Regexp, id string, cell int, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		m := row.FindStringSubmatch(line)
		if m == nil || m[1] != id {
			continue
		}
		// A table row is "| a | b |", so splitting on the pipe leaves an empty
		// field at each end and cell n at index n+1.
		parts := strings.Split(line, "|")
		if len(parts) < cell+2 {
			return fmt.Errorf("%s: row for %s has %d cells, wanted at least %d", path, id, len(parts)-2, cell+1)
		}
		parts[cell+1] = " " + value + " "
		lines[i] = strings.Join(parts, "|")

		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		return nil
	}
	return fmt.Errorf("no row for %s in %s", id, path)
}

// specStatus matches the Status line of a spec file's Meta section.
var specStatus = regexp.MustCompile("(?m)^(- Status:\\s*)`[a-z-]+`")

// setSpecStatus rewrites the spec file's own Status line and appends a dated
// line to its log. The spec is found by ID prefix, so renaming its short name
// does not break the link.
//
// A missing spec file is not an error: a row may have been added by hand before
// anyone wrote the spec, and refusing to move the status would leave the index
// stuck instead.
func setSpecStatus(root, dir, id string, from, to Status, date, logSection string) error {
	matches, err := filepath.Glob(filepath.Join(root, dir, id+"-*.md"))
	if err != nil {
		return fmt.Errorf("look for %s under %s: %w", id, dir, err)
	}
	if len(matches) == 0 {
		return nil
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		text := specStatus.ReplaceAllString(string(data), "${1}`"+string(to)+"`")

		entry := fmt.Sprintf("- %s status `%s` -> `%s`", orDash(date), from, to)
		text, err = appendUnderHeading(text, logSection, entry)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// appendUnderHeading adds a line at the end of a heading's section.
func appendUnderHeading(text, heading, line string) (string, error) {
	lines := strings.Split(text, "\n")

	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == heading {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return text, fmt.Errorf("no %q section", heading)
	}

	end := start
	for end < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[end]), "## ") {
		end++
	}
	// Step back over trailing blank lines so the entry joins the list rather
	// than sitting below a gap.
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	updated := make([]string, 0, len(lines)+1)
	updated = append(updated, lines[:end]...)
	updated = append(updated, line)
	updated = append(updated, lines[end:]...)
	return strings.Join(updated, "\n"), nil
}

// nextTaskSection is the heading the task index keeps a pointer under. It is
// there so that an agent opening the repo does not have to read the whole queue
// to find out what is being worked on.
const nextTaskSection = "## Next Task"

// refreshNextTask rewrites the index's Next Task section from the queue: the
// task in progress if there is one, otherwise the first one ready, otherwise
// the first draft.
func refreshNextTask(root string) error {
	tasks, err := ReadTasks(root)
	if err != nil {
		return err
	}

	body := "Nothing in flight. The next task usually opens once a request has a target project."
	for _, want := range []Status{InProgress, Ready, Draft} {
		for _, t := range tasks {
			if t.Status == want {
				body = fmt.Sprintf("%s %s - `%s`", t.ID, t.Title, t.Status)
				break
			}
		}
		if !strings.HasPrefix(body, "Nothing") {
			break
		}
	}

	path := filepath.Join(root, TaskIndex)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")

	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == nextTaskSection {
			start = i + 1
			break
		}
	}
	// An index without the section is left alone rather than reshaped: it may
	// be a repo scaffolded by an older version, and the queue is still right.
	if start < 0 {
		return nil
	}

	end := start
	for end < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[end]), "## ") {
		end++
	}

	updated := make([]string, 0, len(lines))
	updated = append(updated, lines[:start]...)
	updated = append(updated, "", body, "")
	updated = append(updated, lines[end:]...)

	if err := os.WriteFile(path, []byte(strings.Join(updated, "\n")), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// readOptional returns the file's contents, or empty when it does not exist. A
// repo early in an onboarding has none of these files yet, and that is a state
// rather than a failure.
func readOptional(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

// slugPattern collapses everything that is not a lowercase ASCII word
// character into a single hyphen.
var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify derives a file-name fragment from a title. It returns empty when the
// title carries no usable ASCII, which happens whenever the customer's words
// are the title - the caller asks for an explicit slug then rather than writing
// a file nobody can name.
func Slugify(title string) string {
	s := slugPattern.ReplaceAllString(strings.ToLower(title), "-")
	s = strings.Trim(s, "-")

	// A file name reads better short, and a title is a sentence. Cut at the
	// last word boundary inside the limit rather than mid-word, which is how
	// REQ-001-warehouse-staff-want-to-ask-about-stoc happens.
	const limit = 32
	if len(s) > limit {
		s = s[:limit]
		if i := strings.LastIndex(s, "-"); i > 0 {
			s = s[:i]
		}
		s = strings.Trim(s, "-")
	}
	return s
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
