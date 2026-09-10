// Package needs answers what a scenario has to be obtained from the customer
// before it can be built.
//
// **It is a view over material that already existed and could not be reached.**
// Every row below is written down somewhere - the interview's coordinates
// question, the per-channel credential table, the four-shape ladder for a system
// nobody here has integrated - and reaching it meant reading a 1200-line
// interview, a wiki page and three extracts and assembling the answer. The
// expensive failure it exists for is not choosing the wrong shape: it is week
// three, when the allowlist somebody never asked for turns out to need a ticket,
// an approval and a window.
//
// So no row states a fact of its own. Each carries `From`, the document that
// owns it, and `audit-material --links` resolves those the way it resolves every
// other pointer here - which is what stops this becoming a fifth place the same
// fact is written.
package needs

import (
	"fmt"
	"sort"
	"strings"

	"testing/fstest"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
)

// Item is one thing to obtain, and where the material says so.
type Item struct {
	// Ask is what to ask for, in the shape to ask for it. Specific on purpose:
	// offering options invites the other side to pick one that does not apply,
	// and the week it takes to find that out is the week being saved.
	Ask string `json:"ask"`

	// Why is what it changes, so somebody can drop it when it does not apply.
	Why string `json:"why"`

	// From is the document that owns this. A pointer, not a restatement.
	From string `json:"from"`
}

// Shape is a deployment shape and what it needs from the customer.
type Shape struct {
	Name  string `json:"shape"`
	What  string `json:"what"`
	Items []Item `json:"needs"`
}

var shapes = []Shape{{
	Name: "semantic-layer",
	What: "a database we can read, with an agent asking questions of it",
	Items: []Item{
		{Ask: "host, port, database or schema, and the account name", Why: "there is no connection without them", From: "asgard-cli guide requirements"},
		{Ask: "**is that account read-only?** Ask explicitly", Why: "the one offered first usually is not, and finding out later means going back for a second credential", From: "asgard-cli guide requirements"},
		{Ask: "**that Asgard's four outbound addresses go on their allowlist** - ask their network team for exactly that, not for \"a VPN, an allowlist or a jump host\"", Why: "Asgard is hosted and the agent runs in the platform's own cloud; there is nothing of ours to put on their network. **This is the single most expensive thing to discover in week three**, and it is a ticket, an approval and a window in most companies", From: "asgard-cli wiki operations"},
		{Ask: "whether the allowlist change can be done, and roughly when", Why: "a date changes the plan. Do not ask who approves it - a name changes nothing we build", From: "asgard-cli guide requirements"},
		{Ask: "a description of every cube, dimension and measure, in the customer's own words", Why: "the CRD requires a description on each, and it is what the model matches on - not the column name", From: "asgard-cli usecase semantic-layer"},
	},
}, {
	Name: "external-api",
	What: "a system with an HTTP API rather than a database",
	Items: []Item{
		{Ask: "the base URL, the auth scheme, and a credential for it", Why: "endpoints and non-secret settings become chart values; a token is a secret", From: "asgard-cli usecase external-api"},
		{Ask: "**whether there is a test environment**, before designing a mock", Why: "writing into a real test environment proves the fields, the validation rules and the status codes; a mock proves none of them", From: "asgard-cli usecase write-path"},
		{Ask: "if it is production-only, **whether they permit testing against it**", Why: "in the meeting, not assumed here - the answer decides whether the first delivery can be proved at all", From: "asgard-cli wiki taiwan-channels"},
		{Ask: "the rate limit", Why: "it decides whether a Syncer can keep up, and whether a tool can be called per turn", From: "asgard-cli guide requirements"},
	},
}, {
	Name: "chat-channel",
	What: "the agent reached from a chat platform the customer's users already use",
	Items: []Item{
		{Ask: "**which channel**, in the same breath as who is on the other end", Why: "`botProviderClass` is immutable once created, so changing it later is a new BotProvider rather than an edit", From: "asgard-cli guide requirements"},
		{Ask: "LINE: Channel Secret and Channel Access Token — and somebody who can paste a Webhook URL back into the LINE Developers Console and enable Use webhook", Why: "**LINE is the only two-way setup**: Asgard produces a URL that has to go back. The rest only take credentials inward", From: "asgard-cli wiki integration"},
		{Ask: "Slack: Client ID, Client Secret, Signing Secret and the permission scopes", Why: "a Slack app has to exist and subscribe to bot events", From: "asgard-cli wiki integration"},
		{Ask: "Discord or Telegram: the Bot Token", Why: "Discord also needs the bot invited to the server; Telegram's comes from BotFather", From: "asgard-cli wiki integration"},
		{Ask: "whether anything sits between the channel and us", Why: "an existing bot, a middleware, a support desk already on that channel - it changes the entry point", From: "asgard-cli guide requirements"},
	},
}, {
	Name: "knowledge-drive",
	What: "documents the agent reads - manuals, FAQs, pages",
	Items: []Item{
		{Ask: "the documents themselves, or the place they live and access to it", Why: "a Drive syncs from somewhere; without the source there is nothing to index", From: "asgard-cli usecase knowledge-drive"},
		{Ask: "who keeps them current, and how often they change", Why: "it decides the Syncer's schedule, and whether a stale answer is a real risk", From: "asgard-cli guide requirements"},
	},
}, {
	Name: "write-path",
	What: "the agent doing something rather than answering",
	Items: []Item{
		{Ask: "**whether there is a test environment**, first", Why: "reaching for a mock before asking loses the strongest version of the first delivery", From: "asgard-cli usecase write-path"},
		{Ask: "who is on the other end when the gate stops for approval", Why: "on a public channel the person approving is the visitor, not staff - and what that looks like is a platform unknown", From: "asgard-cli wiki platform-unknowns"},
	},
}, {
	Name: "browser-operation",
	What: "a system with no database we can read and no API",
	Items: []Item{
		{Ask: "a login to the back office, and whether a non-production one exists", Why: "the whole shape is driving their UI; there is nothing else to reach", From: "asgard-cli usecase browser-operation"},
		{Ask: "how many pages and operations actually matter", Why: "**SHOPLINE is what this costs: 88 page entry points mapped before the first useful call.** One back-office-only system among four sets the cost of the whole item", From: "asgard-cli wiki taiwan-channels"},
	},
}, {
	Name: "skill-set",
	What: "skills the deployed agent loads at runtime",
	Items: []Item{
		{Ask: "a git repository, and a token for it if it is private", Why: "a SkillSet syncs from a repo; a private one needs a PAT the platform can hold", From: "asgard-cli usecase skill-set"},
	},
}}

// Shapes returns every shape, sorted by name.
func Shapes() []Shape {
	out := append([]Shape(nil), shapes...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Names lists the shapes that have a dependency list.
func Names() []string {
	out := make([]string, 0, len(shapes))
	for _, s := range shapes {
		out = append(out, s.Name)
	}
	sort.Strings(out)
	return out
}

// ── Landing ───────────────────────────────────────────────────────────────
//
// `asgard-cli init` writes these shapes into a customer repository as files,
// one per shape, so that an FDE's agent reaches a row by grepping for the word
// the customer used - `allowlist`, `read-only`, `test environment` - rather
// than by knowing this command exists.
//
// **One file per shape rather than one file for all seven**, because a grep hit
// then carries which shape it belongs to. A row is only actionable with that:
// "ask whether there is a test environment" means a different conversation for
// a write path than for an external API.

// Document renders one shape as the markdown that lands at `needs/<name>.md`.
func (s Shape) Document() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: what to get from the customer\n\n", s.Name)
	fmt.Fprintf(&b, "**%s**\n\n", s.What)
	b.WriteString(intro)
	for _, i := range s.Items {
		fmt.Fprintf(&b, "\n## %s\n\n%s\n\nStated in %s.\n", i.Ask, i.Why, "`"+kb.Landed("../", i.From)+"`")
	}
	b.WriteString(provenance)
	return b.String()
}

// provenance is the same on every shape, because it is the same claim: the
// checking is per row rather than per document. Without it these read as
// material nobody has held against anything, which `audit-material
// --unverified` would say and would be the wrong shape of true.
const provenance = `
**Checked:** each row above names the document that owns its claim, and
` + "`asgard-cli audit-material --links`" + ` resolves those. That is the whole of the
checking: a row is as good as the document it cites.

**Unchecked:** the list itself. Nothing holds it against a finished engagement,
so a shape can be missing something every one of its rows is right about.
`

// intro is on every shape rather than in one file they all point at: a reader
// arrives here by grepping for a word in one row, and the rule that governs how
// to ask is worth more at that moment than a pointer to it.
const intro = `**This is theirs to provide, not ours to design.** Ask for exactly the thing
named - offering options invites the other side to pick one that does not
apply, and the week it takes to find that out is the week you were saving.
`

// Documents renders every shape, for the export and for the audit that resolves
// the pointers in them.
func Documents() []struct{ Name, Body string } {
	out := make([]struct{ Name, Body string }, 0, len(shapes))
	for _, s := range Shapes() {
		out = append(out, struct{ Name, Body string }{s.Name, s.Document()})
	}
	return out
}

// corpus is the rendered documents behind one `kb.Corpus`, so the audits and
// the landing read these the way they read every other body of material. They
// have no files - each is rendered from the shapes above - so the FS is built
// from them.
var corpus = func() kb.Corpus {
	files := fstest.MapFS{}
	for _, d := range Documents() {
		files["needs/"+d.Name+".md"] = &fstest.MapFile{Data: []byte(d.Body)}
	}
	return kb.Corpus{
		FS:      files,
		Dir:     "needs",
		Noun:    "shape",
		Command: "ls .agents/skills/asgard-platform/needs/",
	}
}()

// List returns every shape as a document.
func List() ([]kb.Doc, error) { return corpus.List() }
