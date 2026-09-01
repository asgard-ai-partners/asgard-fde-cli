package usecase

import (
	"strings"
	"testing"
)

// TestEveryExtractIsActionable is the standard these were rewritten to meet: an
// agent should be able to author a CR from one, not merely understand it.
func TestEveryExtractIsActionable(t *testing.T) {
	list, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) < 9 {
		t.Fatalf("found %d extracts, want the full set", len(list))
	}

	for _, e := range list {
		t.Run(e.Name, func(t *testing.T) {
			content, err := Read(e.Name)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}

			// conventions is not a shape: it is what the shapes assume, so it
			// carries no decision and no verification of its own.
			if e.Name == "conventions" {
				for _, want := range []string{"templates/", "asgard-ai.com/", "```yaml"} {
					if !strings.Contains(content, want) {
						t.Errorf("missing %q", want)
					}
				}
				return
			}

			for _, want := range []struct{ section, why string }{
				{"## When this shape, and when not", "the decision comes before the YAML"},
				{"## Generate it", "hand-copying a skeleton is where the silent-failure fields get lost"},
				{"## The skeleton", "without one, a reader knows the rules and still cannot start"},
				{"## Verify", "a shape that cannot be checked is not finished"},
			} {
				if !strings.Contains(content, want.section) {
					t.Errorf("missing %q - %s", want.section, want.why)
				}
			}

			// Structure and traps are not enough: the generator writes those.
			// What it cannot write is the judgement, and that is what a reader
			// comes here for.
			if !strings.Contains(content, "Designing") && !strings.Contains(content, "Writing the prompt") {
				t.Error("no design guidance - the generated file marks those parts TODO, " +
					"and this is where a reader finds out how to fill them")
			}

			if !strings.Contains(content, "```yaml") {
				t.Error("no YAML at all; the skeleton has to be copyable")
			}
			if !strings.Contains(content, "templates/") {
				t.Error("does not say where the file goes")
			}
			if e.Title == "" || e.Summary == "" {
				t.Errorf("title=%q summary=%q; both are what the listing shows", e.Title, e.Summary)
			}
		})
	}
}

// TestExtractsNameNoCustomer guards what ships. These are embedded in a binary
// that runs inside every customer's repo.
func TestExtractsNameNoCustomer(t *testing.T) {
	leaked := []string{
		"unitech", "freyr", "xxentria", "finance-ai", "buy123",
		"auto-post", "xxtechec", "shopline", "netsuite",
	}

	list, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range list {
		content, err := Read(e.Name)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		lower := strings.ToLower(content)
		for _, name := range leaked {
			if strings.Contains(lower, name) {
				t.Errorf("%s mentions %q, which belongs to the deployment it was taken from", e.Name, name)
			}
		}
	}
}

func TestSearchNarrowsWithMoreTerms(t *testing.T) {
	broad, err := Search("agent")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	narrow, err := Search("agent anonymous visitor")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Every term must appear, so adding one can only narrow. The opposite would
	// make a long, specific query useless.
	if len(narrow) >= len(broad) {
		t.Errorf("broad matched %d, narrow matched %d; more terms should narrow", len(broad), len(narrow))
	}
}

// TestSearchFindsTheShapeFromARequirement checks the path that matters: someone
// has the customer's words and needs the shape's name.
func TestSearchFindsTheShapeFromARequirement(t *testing.T) {
	tests := []struct {
		requirement string
		query       string
		want        string
	}{
		{"staff ask across internal systems", "internal authenticated", "agent-hub"},
		{"a daily report with nobody watching", "schedule", "trigger"},
		{"anonymous visitors on a public page", "anonymous", "flow-agent-single"},
		{"product documents and FAQs", "documents", "knowledge-drive"},
		{"open-ended questions over a database", "open-ended", "semantic-layer"},
	}

	for _, tt := range tests {
		t.Run(tt.requirement, func(t *testing.T) {
			matches, err := Search(tt.query)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			for _, m := range matches {
				if m.Name == tt.want {
					return
				}
			}
			var got []string
			for _, m := range matches {
				got = append(got, m.Name)
			}
			t.Errorf("searching %q returned %v, want %s among them", tt.query, got, tt.want)
		})
	}
}

func TestReadUnknownExtract(t *testing.T) {
	_, err := Read("no-such-shape")
	if err == nil {
		t.Fatal("want an error for an unknown name")
	}
	if !strings.Contains(err.Error(), "asgard-cli usecase") {
		t.Errorf("error %q should say how to list them", err)
	}
}
