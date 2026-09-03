// Package brief answers a different question from `next`.
//
// `next` answers "where am I", derived from the repository, and it does that
// well. The question it cannot answer is "the thing I am about to do - where
// will I get it wrong", and that is the question with the expensive answers,
// for three reasons.
//
// The riskiest activity leaves no trace in a repository. Talking to a customer
// changes no file, so nothing state-derived can prepare anybody for it. The
// walk is arranged by position, and meetings happen at every position - an
// engagement at the entry-point stage still has a meeting tomorrow. And the
// material is arranged for reading at the point of use, which is right, except
// where what you need to know precedes where it is filed.
//
// The knowledge was already here and already marked: every stage names the
// decision that gets answered wrong, every extract carries an Unchecked block,
// and platform-unknowns is a page of nothing else. What was missing was a way
// in that is addressed by intent rather than by position.
package brief

import (
	"fmt"
	"sort"
	"strings"
)

// Item is one thing that gets got wrong, and where the right version lives.
type Item struct {
	Subject string
	Wrong   string
	Right   string
	Where   string
}

// Activity is something an FDE is about to do.
type Activity struct {
	Name  string
	When  string
	Lead  string
	Items []Item
	Close string
}

// Activities are the ones with a known way to go wrong. The test for adding one
// is whether somebody has actually got it wrong, not whether it is important.
var Activities = []Activity{
	{
		Name: "customer-meeting",
		When: "before any conversation with the customer, at any stage",
		Lead: `**Four of these five fail in the same direction: the intuitive answer
undersells the platform or overstates a limit.** So the error is not neutral -
it gives away scope, and from the customer's side being careful and being wrong
look identical.

This stage's output is speech. Every other stage produces a file, which is
edited and re-rendered; a wrong sentence is in their notes.`,
		Items: []Item{
			{
				"how we reach a system inside their network",
				`"a VPN, an allowlist or a jump host - whichever suits you"`,
				"One shape only: they add our outbound addresses to their allowlist. Asgard is hosted and the agent runs in a sandbox in that cloud, so there is nothing of ours to place on their network. Offering options gets one chosen that does not apply. **Ask who can change the firewall - do not hand out the addresses**",
				"asgard-cli wiki operations",
			},
			{
				"whether an anonymous channel knows who is asking",
				`"it cannot tell who they are, so identity has to wait"`,
				"An anonymous channel answers \"where is MY order\" perfectly well. The caller supplies identity server-side every turn; only the **model** supplies nothing. **LINE's webhook carries a userId.** Deferring this gives away something you already had",
				"asgard-cli next --stage read-path",
			},
			{
				"how many agents",
				`"one agent, then"`,
				"A flow-agent shape renders **zero `Agent` CRs** - the prompt is on a Workflow processor and the capabilities on the blueprint. A quote built on the instinct is priced for work that does not exist",
				"asgard-cli size flow-agent-single",
			},
			{
				"how an anonymous caller is identified",
				`"they can give us their ticket number and we look it up"`,
				"**A ticket number is not authentication.** They are usually sequential, so a lookup keyed on one alone lets anybody enumerate other people's cases. Any self-service query on a public channel has to say what identifies the person - and if the answer is a number they type, there is no answer yet",
				"asgard-cli next --stage read-path",
			},
			{
				"how a write is tested",
				`"during the test, shall we really create the ticket or just draft it?"`,
				"**Ask whether they have a test environment first.** That question assumes they only have production, and it gives away the strongest version of the first delivery: writing into a test system proves the fields, the validation rules and the status codes, and a mock proves none of them. A mock is the fallback when there is no test environment - or when they will not let us write to production, which is their call",
				"asgard-cli usecase write-path",
			},
			{
				"whether it can do things, not only answer",
				`"we would have to build an approval step for that"`,
				"The approval gate is **what the platform is built around**. A write stops, names the tool, and nothing runs until a person allows it. It is not custom work",
				"asgard-cli usecase write-path",
			},
			{
				"whether it can email or message people",
				`"sure, we can have it send you a summary"`,
				"**The platform sends nothing.** No SMTP, no mail toolset, nothing in the core. It can call an endpoint of theirs that sends mail - if they have one. Promising a notification without asking that first is promising something with no way to build it",
				"asgard-cli wiki integration",
			},
			{
				"what the agent can reach",
				`"only the systems you connect to it"`,
				"**Not true.** Every sandbox carries the coding CLI's own tools - web search and fetch among them - and no Toolset or blueprint setting removes them. The only control is an instruction in the prompt",
				"asgard-cli wiki tools",
			},
			{
				"whether they can use their own model or their own key",
				`"yes, we can point it at your account"`,
				"**Only on Odin.** Sindri and Mimir use the platform's designated models and the LLM cannot be swapped there. So the answer depends on which product the capability lands in - which is question 2b, decided before anyone thinks about billing. A customer with a model contract or a rule about where inference happens needs this while the shape is still open",
				"asgard-cli wiki fehu",
			},
			{
				"whether they can have a specific model",
				`"of course, we will configure that one"`,
				"You can, and it costs something worth saying: a builtin tier is a **logical model backed by several providers with automatic failover**, and a custom `CompletionModel` is one provider, one key, one point of failure. A good reason - compliance, an existing contract, a model they tested against - is fine. A preference usually is not",
				"asgard-cli wiki settings",
			},
			{
				"what the limits are",
				`"30 steps and 3 minutes per request is the limit"`,
				"Those are **defaults**, raised by contacting sales or service@asgard-ai.com. And **do not quote the step count at all**: nothing defines what a step is, so the next question has no answer",
				"asgard-cli wiki integration",
			},
		},
		Close: `**A slide is not an agenda and not a tracking list.** "Who is
responsible for this?" is right to ask in the room and wrong to print - asking it
says you intend to follow through, printing it reads as managing their
organisation. The carrier is the test, not the content: is this answer somebody
we chase after the meeting, or something we build? Six of one deck's twenty
corrections were this, more than any other kind, and three of those were copied
out of this tool's own material.

**Ask what they have already read**, before describing anything. The
product site tells them Odin is for non-technical staff, gives them a vocabulary
- Basic Function, Template - that maps to nothing here, and says Mimir simulates
the future. ` + "`asgard-cli wiki what-they-read`" + ` has the three, and it costs
one sentence to find out which of them you are correcting.

**Before the meeting, not during it:** anything on
` + "`asgard-cli wiki platform-unknowns`" + ` that this engagement touches is ours to
chase, not theirs to hear about. Standing in front of a customer saying we do
not know what our own product does is not honesty.

**Never on their screen:** implementation nouns, another customer, coordinates -
theirs *or* our own outbound addresses - and this repository's own bookkeeping,
question numbers included. The ` + "`proposal-deck`" + ` skill has the rest.`,
	},
	{
		Name: "write-chart",
		When: "before authoring or editing CRs",
		Lead: `Three decisions in this repository's history were made the obvious way,
built, and reversed. They are obvious in the same way again each time.`,
		Items: []Item{
			{
				"the read surface for a public audience",
				"a SemanticLayer, like everything else",
				"Five zero-parameter query tools. A layer without `allowedCubes` is arbitrary SQL over every cube, and the surface grows by itself each time one is added",
				"asgard-cli usecase fixed-query-tools",
			},
			{
				"the entry point for an anonymous visitor",
				"the platform's agent hub, like everything else",
				"Its own BotProvider -> Workflow -> SandboxBlueprint. An anonymous visitor cannot authenticate to the hub, and `BotProvider.entrypoint` takes a Workflow, never an Agent",
				"asgard-cli usecase flow-agent-single",
			},
			{
				"where unstructured knowledge goes",
				"a KnowledgeBase with Loaders and a retrieval workflow",
				"A SourceSet Drive with `contextIndex`. Both are live; the Drive is the recommendation for new work",
				"asgard-cli usecase knowledge-drive",
			},
			{
				"a write inside a scheduled run",
				"omitting `allowWrite` because it is read-only anyway",
				"**The definitions give `allowWrite` a default of true and `allowQuery` a default of false** - the safe field is off by default and the dangerous one is on. So configuring a processor by adding only what you want produces a write path, and the rendered chart does not show it. Write `allowWrite: false` out on every entry that has a layer, and note that `query-database` carries the same pair",
				"asgard-cli wiki processors",
			},
			{
				"what an Expression may use",
				"modern JavaScript",
				"**ECMA5 only** - no `let`, no arrow functions, no optional chaining - in every field of every processor. Which is why every documented example is defensively written",
				"asgard-cli wiki processors",
			},
			{
				"a SandboxBlueprint's subagents",
				"it deployed, so the blueprint is right",
				"**Exactly one of `baseAgentName` or `aliasName`, and no CRD checks it** - the shape rides inside a JSON string where CEL cannot see it, so the controller enforces it at evaluation time. Both set, or neither, deploys green and fails the first time somebody talks to the agent. The only rule worth reading a blueprint by hand for",
				"asgard-cli wiki crd-rules",
			},
			{
				"whether the chart is enough",
				"a green render means it is done",
				"A Workflow needs a `ConfigMap` of node positions or its editor opens as a pile, and `project-environment-id` or the editor opens blank. Neither is an Asgard CR, so nothing in the gate mentions them",
				"asgard-cli wiki workflow",
			},
		},
		Close: `` + "`asgard-cli check`" + ` and ` + "`asgard-cli verify`" + ` catch the shape. None of the
above is a shape problem, which is why they are here.`,
	},
	{
		Name: "handover",
		When: "before telling anyone it is live, and before a training session",
		Lead: `A green deploy is not a working deployment, and the step between them is
not in this repository.`,
		Items: []Item{
			{
				"why the customer says there is nothing there",
				"a chart problem, so look at the chart",
				"Building a resource does not make it visible. Somebody opens the **Management Console**, finds the product's page, uses *Manage Accounts in* and adds the people - **per resource, because permissions do not inherit**. Nothing in a chart, in `check`, in `verify` or in CD can see that it was skipped",
				"asgard-cli wiki console",
			},
			{
				"what to walk them through",
				"the chart, because that is what we built",
				"The order from a credential to an agent somebody can talk to - and there is no Sindri step in it, because every Managed Agent is published there already",
				"asgard-cli wiki setup-path",
			},
			{
				"which screens to show",
				"whatever is in the documentation",
				"Two of the recommended images carry a `ts-` prefix and the build console's own navigation. **Crop first**, and open every one - nothing records when any was captured",
				"asgard-cli wiki screenshots",
			},
		},
		Close: `A handover deck may show the console; a proposal may not. They are
different documents - ` + "`proposal-deck`" + ` in ` + "`.agents/skills/`" + ` has the table.`,
	},
}

// Find returns an activity by name.
func Find(name string) (Activity, bool) {
	for _, a := range Activities {
		if a.Name == name {
			return a, true
		}
	}
	return Activity{}, false
}

// Names lists the activities for an error message.
func Names() []string {
	out := make([]string, 0, len(Activities))
	for _, a := range Activities {
		out = append(out, a.Name)
	}
	sort.Strings(out)
	return out
}

// Render writes one activity's briefing.
func (a Activity) Render(w interface{ Write([]byte) (int, error) }) {
	fmt.Fprintf(w, "%s\n%s\n\n%s\n\n", a.Name, a.When, a.Lead)
	for _, it := range a.Items {
		fmt.Fprintf(w, "  %s\n", it.Subject)
		fmt.Fprintf(w, "    said:  %s\n", it.Wrong)
		fmt.Fprintf(w, "    true:  %s\n", wrap(it.Right, 68, "           "))
		fmt.Fprintf(w, "    read:  %s\n\n", it.Where)
	}
	fmt.Fprintf(w, "%s\n", a.Close)
}

// wrap breaks a line at word boundaries, indenting continuations.
func wrap(s string, width int, indent string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	line := 0
	for i, word := range words {
		if line > 0 && line+1+len(word) > width {
			b.WriteString("\n" + indent)
			line = 0
		} else if i > 0 {
			b.WriteString(" ")
			line++
		}
		b.WriteString(word)
		line += len(word)
	}
	return b.String()
}
