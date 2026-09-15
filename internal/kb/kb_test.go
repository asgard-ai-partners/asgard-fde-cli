package kb

import "testing"

// The documents below are the shapes this package actually receives: a page
// with no frontmatter, a page with one, and the three ways a `---` appears in a
// document that has none.

const plainDoc = `# Management Console - permissions and Workspace

The Console does no business work. It is where a workspace is
administered.

**Checked:** against asgard-docs 1a2b3c4, every permission layer named.
Not the pages below Settings.

**Unchecked:** which of the two pages wins when they disagree.

## Permission layers

Body text.
`

const frontDoc = `---
group: Products and scope
description: "the permission layers: what each is for"
---
# Management Console - permissions and Workspace

The Console does no business work. It is where a workspace is
administered.

**Checked:** against asgard-docs 1a2b3c4, every permission layer named.

**Unchecked:** which of the two pages wins when they disagree.

A line that opens description: the way a key does, below the block.
`

// unknownKeyDoc is a skill's frontmatter shape: the Agent Skills contract puts
// fields there that this parser has no field for.
const unknownKeyDoc = `---
name: db-query
group: Building a chart
version: 1.0.0
alwaysApply: false
description: what to ask before a query goes into a Toolset
---
# Querying a customer's source system

Modelling anything against a customer's data starts here.
`

// commentDoc carries a YAML comment in its frontmatter, which is a `# ` line
// above the title.
const commentDoc = `---
# a group is the question the section heading asks
group: Products and scope
description: the permission layers, and which pages disagree
---
# Management Console - permissions and Workspace

The Console does no business work.
`

// ruleDoc is a horizontal rule in the middle of a page, which is not
// frontmatter and must not be read as one.
const ruleDoc = `# How a workflow is assembled

A workflow is a graph, and the edges are what carry the payload.

---

group: not frontmatter
description: also not frontmatter

**Checked:** against asgard-kube cbd8d70.
`

// unterminatedDoc is the block somebody forgot to close.
const unterminatedDoc = `---
group: Products and scope
description: the permission layers

# Management Console - permissions and Workspace

The Console does no business work.

description: and this line is prose, not a field
`

func TestParseWithoutFrontmatter(t *testing.T) {
	d := Parse("console", []byte(plainDoc))
	want := Doc{
		Name:      "console",
		Title:     "Management Console - permissions and Workspace",
		Summary:   "The Console does no business work. It is where a workspace is administered.",
		Checked:   "against asgard-docs 1a2b3c4, every permission layer named. Not the pages below Settings.",
		Unchecked: "which of the two pages wins when they disagree.",
	}
	checkDoc(t, d, want)
	if !d.Verified() {
		t.Error("Verified() = false on a document carrying both markers")
	}
}

func TestParseEmpty(t *testing.T) {
	checkDoc(t, Parse("nothing", nil), Doc{Name: "nothing"})
}

func TestParseFrontmatter(t *testing.T) {
	d := Parse("console", []byte(frontDoc))
	checkDoc(t, d, Doc{
		Name:        "console",
		Group:       "Products and scope",
		Description: "the permission layers: what each is for",
		Title:       "Management Console - permissions and Workspace",
		Summary:     "The Console does no business work. It is where a workspace is administered.",
		Checked:     "against asgard-docs 1a2b3c4, every permission layer named.",
		Unchecked:   "which of the two pages wins when they disagree.",
	})
}

func TestParseUnknownFrontmatterKey(t *testing.T) {
	d := Parse("db-query", []byte(unknownKeyDoc))
	checkDoc(t, d, Doc{
		Name:        "db-query",
		Group:       "Building a chart",
		Description: "what to ask before a query goes into a Toolset",
		Title:       "Querying a customer's source system",
		Summary:     "Modelling anything against a customer's data starts here.",
	})
}

func TestParseHorizontalRuleIsNotFrontmatter(t *testing.T) {
	d := Parse("workflow", []byte(ruleDoc))
	checkDoc(t, d, Doc{
		Name:    "workflow",
		Title:   "How a workflow is assembled",
		Summary: "A workflow is a graph, and the edges are what carry the payload.",
		Checked: "against asgard-kube cbd8d70.",
	})
	if got := Body(ruleDoc); got != ruleDoc {
		t.Errorf("Body() changed a document whose `---` is a horizontal rule:\n%q", got)
	}
}

func TestParseUnterminatedFrontmatter(t *testing.T) {
	d := Parse("console", []byte(unterminatedDoc))
	checkDoc(t, d, Doc{
		Name:    "console",
		Title:   "Management Console - permissions and Workspace",
		Summary: "The Console does no business work.",
	})
	if got := Body(unterminatedDoc); got != unterminatedDoc {
		t.Errorf("Body() stripped an unterminated block:\n%q", got)
	}
}

func TestScalar(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{"double quoted", `  "the permission layers: what each is for"  `, "the permission layers: what each is for"},
		{"escaped inner quote", `"the page that says \"no\" twice"`, `the page that says "no" twice`},
		{"single quoted", `  'a value YAML needs quoting'  `, "a value YAML needs quoting"},
		{"apostrophe inside double quotes", `"what the customer's own agent reads"`, "what the customer's own agent reads"},
		{"unquoted colon and hash", `the split follows the audience: ask against watch, #2 either way`, "the split follows the audience: ask against watch, #2 either way"},
		{"one double quote", `"`, `"`},
		{"one single quote", `'`, `'`},
		{"opening quote only", `"the layers`, `"the layers`},
		{"closing quote only", `the layers"`, `the layers"`},
		{"quotes that do not match", `'the layers"`, `'the layers"`},
		{"empty", ``, ``},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := scalar(c.in); got != c.want {
				t.Errorf("scalar(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestBody(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{
			"strips the block and nothing else",
			"---\ngroup: Products and scope\ndescription: the layers\n---\n# Console\n\nIt does no business work.\n",
			"# Console\n\nIt does no business work.\n",
		},
		{
			"a document with no frontmatter is unchanged",
			plainDoc,
			plainDoc,
		},
		{
			"a later `---` survives",
			"---\ngroup: Products and scope\n---\n# Console\n\nOne.\n\n---\n\nTwo.\n",
			"# Console\n\nOne.\n\n---\n\nTwo.\n",
		},
		{
			"a `---` that is not at the top is not a block",
			ruleDoc,
			ruleDoc,
		},
		{
			"an unterminated block is not a block",
			unterminatedDoc,
			unterminatedDoc,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := Body(c.in); got != c.want {
				t.Errorf("Body() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestParseOfBodyAgrees is the property that ties the three together. What
// lands keeps the frontmatter and what prints strips it, so the two paths must
// not disagree about what the document says.
func TestParseOfBodyAgrees(t *testing.T) {
	for _, c := range []struct{ name, doc string }{
		{"frontmatter", frontDoc},
		{"unknown keys", unknownKeyDoc},
		{"a YAML comment in the block", commentDoc},
		{"no frontmatter", plainDoc},
		{"a horizontal rule", ruleDoc},
		{"an unterminated block", unterminatedDoc},
	} {
		t.Run(c.name, func(t *testing.T) {
			whole := Parse("x", []byte(c.doc))
			printed := Parse("x", []byte(Body(c.doc)))
			if whole.Title != printed.Title {
				t.Errorf("Title: whole %q, printed %q", whole.Title, printed.Title)
			}
			if whole.Summary != printed.Summary {
				t.Errorf("Summary: whole %q, printed %q", whole.Summary, printed.Summary)
			}
			if whole.Checked != printed.Checked || whole.Unchecked != printed.Unchecked {
				t.Errorf("markers: whole %q/%q, printed %q/%q",
					whole.Checked, whole.Unchecked, printed.Checked, printed.Unchecked)
			}
		})
	}
}

// checkDoc compares the fields Parse fills from the document's own text. Links
// and Sources are the other patterns' business and have their own coverage.
func checkDoc(t *testing.T, got, want Doc) {
	t.Helper()
	for _, f := range []struct{ name, got, want string }{
		{"Name", got.Name, want.Name},
		{"Title", got.Title, want.Title},
		{"Summary", got.Summary, want.Summary},
		{"Group", got.Group, want.Group},
		{"Description", got.Description, want.Description},
		{"Checked", got.Checked, want.Checked},
		{"Unchecked", got.Unchecked, want.Unchecked},
	} {
		if f.got != f.want {
			t.Errorf("%s = %q, want %q", f.name, f.got, f.want)
		}
	}
}
