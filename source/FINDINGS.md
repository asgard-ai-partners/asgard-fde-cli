# FINDINGS.md

Defects this tool has produced, and **why nothing caught them**. The second half
is the reusable part: each row names a class of failure and the gap it slipped
through, so the next one of its kind is recognisable before it ships.

> **Internal only**, like `SOURCES.md`. It sits outside `internal/` so `go:embed`
> cannot reach it. Carried forward across rewrites of `TASK.md`, because each
> line cost somebody a real investigation.

Each of these was invisible to every check that existed, and **why it was
invisible is the reusable part**.

| defect | why nothing caught it |
|---|---|
| **The generator wrote skeletons that could not pass the gate.** `add agent` hardcoded `skillSetNames: [sk-base]`, which nothing creates and which `add skillset` cannot create without a git URL the agent does not have, and it ignored the `SemanticLayer` in the same chart. The flowagent blueprint had the same hardcoded name. | The generator never read the chart it was writing into. Fixed by `generate.Resolve`: one layer is mounted (the choice is forced by "one system, one layer, one Agent"), several are refused with their names, a SkillSet is referenced only if one exists. |
| **Two CRs with one name rendered clean.** Several fixed query tools belong to one Toolset - the normal shape - and each `add querytool` emitted its own copy of that Toolset. Kubernetes keeps whichever applies last, so a tool silently vanished. | helm renders both, apply reports success twice. Now `verify` rejects a duplicate `kind/name`, and `Resolve` stops the second tool re-emitting the set. |
| **A project with no Syncer passed the gate and would fail CD.** | CD's requirements are not the apiserver's. Now `gate.Deployability` warns - a warning, not a failure, because zero Syncers is correct until skills are added and failing would leave the gate red through the middle of every engagement. |
| **The gate said what was wrong and not what to do.** `no SkillSet/sk-base` has an obvious cheap resolution - delete the reference - and it is the wrong one. | Written for a reader who already knew the design. Every dangling reference now names the `asgard-cli add` command, with the real project in it so it runs as printed. |
| **The generated `common/README.md` taught a file deleted two rounds earlier**, and recommended two other customers' repository layouts. | It was the one file excluded from that rename and never caught up, and prose compiles under no test. `TestGeneratedDocsOnlyNameFilesThatExist` was written to fail on any `common/` or `scripts/` path the scaffold does not write; it has since been removed. |
| **`gate.Read` failed on a non-mapping YAML document** where the Python original skipped it. | A captured stderr line would have failed the gate for the wrong reason - the exact case the Python tolerance existed for. |
| **`next --stage` was silently ignored with no config.** | A flag the code does not read, which `AGENTS.md` forbids. |
| **A live workspace id shipped in a stage prompt as an example.** | It is just a number, and prompts are prose. |
| **`add <kind> <name>` doubled the kind prefix**, writing `metadata.name: dc-dc-erp`. | helm lint and a server dry-run both accept it; only the xref gate catches it, and only once something references the name that was meant. |
| **`check_agent_split.py` crashed with a raw `ModuleNotFoundError`** when PyYAML was missing, while its siblings printed a friendly line. | Since ported to Go, so the dependency is gone entirely. |
| **Nothing supported a chat platform.** Every extract and the BotProvider template hardcoded `botProviderClass: generic`, so a customer wanting LINE - the most likely channel for a Taiwanese customer - had nothing. | Three reference deployments were listed as "not yet cited" and had not been read. The class vocabulary (`generic \| line \| telegram \| discord \| slack`) was only in the platform's CRD contract, which no extract drew from. Fixed by `usecase chat-channel` and `add flowagent --bot-class`. |
| **`expression` is documented as CEL and is JavaScript.** The platform's CRD documentation says CEL; every chart uses arrow functions, `const`, `String()`, `encodeURIComponent` and `??`, none of which CEL has. | Nobody had read the docs and the charts against each other. An agent handed "it is CEL" writes something that cannot work and has no way to find out why. Recorded in `source/SOURCES.md` as a docs-versus-reality conflict, and corrected in `workflow-chain`. |
| **Three mechanisms every shape is built on had no extract**: the two-call token chain, the `relationships` graph, and what crosses between processors (`prevPayload` being replaced by an `http-request`, `httpResponse` meaning "most recent"). | The extracts were organised by deployment shape, so a mechanism used by all of them belonged to none. |
| **Five extracts were reachable from nothing** - `browser-operation`, `write-path` and the three new ones - so the only way to find one was to browse a list of sixteen for a mechanism you did not know you needed. | Writing material and pointing at it are separate acts, and only the first had a test. `TestEveryExtractIsReachable` was written to fail the build when a generator kind or a stage prompt does not send the reader to an extract; it has since been removed. |
| **Two generators emitted CRs the apiserver rejects.** `semanticlayer` omitted `spec.joins` and `cubes[].measures`; `knowledgedrive`'s database Syncer omitted `database.query` and `database.batchSize`. All four are **required**, so every chart built from them would have failed CD after the FDE had filled in the TODOs - on fields the template never mentioned. | Nothing had ever compared a rendered CR against the platform's schema. `helm lint` validates YAML and templating, `check` validates repository structure, `verify` validates cross-references and CD's requirements; a required field missing from a CRD is outside all three. Found by rendering every kind and validating against `asgard-kube`'s `crd/*.yaml`. |
| **`knowledgedrive` put `isMaxValueColumn` on the `database` block instead of on a column**, and `knowledge-drive.md` taught the same shape. The apiserver drops an unknown field without a word, so the incremental cursor the template's own comment describes did not exist - the Syncer re-reads the whole table every run, for as long as nobody notices. | An unknown field is the one error with no signal anywhere: it renders, lints, applies and reports success. Only a schema comparison finds it. The correct shape was already in production in `unitech-e`, which nothing had been checked against. |
| **Every generated Trigger had write access to its semantic layers.** The `semanticLayers` config carried `allowQuery` and no `allowWrite`, and `stream-llm-completion-message` resolves a **missing** `allowWrite` to `true`. `trigger.md` stated the opposite as a reassurance: "allowQuery without allowWrite: the source systems stay read-only." | A default that inverts the safe reading, documented only in the platform's own types. `agent.yaml.tmpl` got it right because `Agent.managed.semanticLayers[]` is where the rule is written down; the processor config form is the same field in a different place, and the rule did not travel. The exposure was largest on the one shape with nobody watching it. |
| **`querytool` told the reader to put `name` and `description` on a Toolset `tools[]` entry.** Neither exists there - both are dropped, and the tool reaches the model unnamed. The text the model actually reads is the Workflow entry's `tooling` block, twenty lines further up the same file. | It was in a comment, and prose compiles under no test. The two fields exist on a neighbouring type, so the advice reads as plausible to anyone who has not checked. |
| **A stage prompt still taught a field the platform retired.** `06-knowledge.md` documented `destinationMemberKey` / `stateMemberKey` and "must name declared members" four days after the member registry was removed. `skill-set.md` had been updated; the prompt had not. | Platform vocabulary appears in three places - templates, extracts and stage prompts - and only the first two are near the code that changed. Nothing links a retired field name to every file that mentions it. |
| **A second CR sharing a values block silently got no entry.** `botProviders:` exists after the first flow agent, so the second one's whole snippet was skipped as "already there" and its template read `.Values.botProviders.<name>.disabled` off a map with no such entry. | `appendValues` compared only the top-level key. **Bare `helm lint` catches it** - which is exactly what the bare pass is for - but only after the file is on disk, and an env file overlaid masks it entirely. |

Two lessons that are mine rather than the code's:

- **Say a mechanism cannot deliver the guarantee before building it.** Two rounds
  went into packaging metadata that was inert on every install path this project
  actually has.
- **"Not deployable" and "not functional" are different claims.** The generated
  chart applies cleanly; it is empty. Using the wrong word hid the real
  deployability gap (the Syncer) for a round.

Verified against `unitech-e-asgard-kube`: its `internal` project declares the same
CR kinds the CLI generates and carries **zero TODOs**. Same skeleton, content
filled in during the engagement. So the structure is right, and what remains is
content - which is what the staged prompts are for.

## The three reversals the whole design rests on

Of the target repo's 13 task specs, **three were superseded and one reverted** -
and they are the three decisions that matter most:

| decision | first answer | final answer | why it flipped |
|---|---|---|---|
| how `website` reads catalog data | a SemanticLayer (TASK-003) | five zero-parameter query tools (TASK-010) | a layer without `allowedCubes` is arbitrary SQL over every cube, and the exposed surface grows by itself each time a table is added |
| how `website` is reached | the platform's agent hub (TASK-007) | Flow Agent: own BotProvider -> Workflow -> SandboxBlueprint (TASK-011) | an anonymous visitor cannot authenticate to the agent hub, and `BotProvider.entrypoint` takes a Workflow, never an Agent |
| how `website` stores knowledge | `KnowledgeBase` + `Loader` + Toolset (TASK-012) | `SourceSet` Drive + contextIndex + read-only mount (TASK-013) | the KnowledgeBase mechanism is deprecated platform-side |

**So a linear wizard is the wrong shape.** Each was answered once, implemented,
and reversed after contact with reality. What the next customer needs is not a
script that walks them into the same first answer - it is the question with the
trap already visible.

That is the CLI's highest-value output: **stages 4, 5 and 6 print the wrong
answer next to the right one**, because the wrong one is what looks obvious.

Stage prompts 4, 5 and 6 carry these; `asgard-cli wiki agents` and
`usecase fixed-query-tools` / `knowledge-drive` cite them.
