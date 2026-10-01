# references

`references/` stores background material for humans and agents. It is not the implementation source of truth.

Use this directory for:

- Design specs, decision records, and architecture notes.
- The customer's own system documentation - schema dumps, ERDs, field dictionaries, status-code tables.
- External API documentation and third-party integration notes.
- Kubernetes / Helm / platform behavior notes and migration references.
- Research, meeting notes, and non-final decisions.

Rules:

- Convert reference material into `requirements/` before implementation.
- If `references/` conflicts with `requirements/`, trust `requirements/` and record the conflict as an open question or decision.
- Prefer clear source names, such as `pm-spec/` or `api-reference.md`.
- Record where each document came from and when, at the top of the file: who
  supplied it, its date, and whether anything in it was verified against the real
  system. Material a customer wrote for their own staff describes the system they
  believe they have, and nothing on a stale page shows that it is stale.

## Filing something large

One customer manual or API reference is a file. A whole system's worth is a
directory, and it needs an index or nobody reads past the first page:

```
<system>/README.md              what this covers, where it came from, when
<system>/conventions.md         the cross-cutting rules to read first
<system>/api/<domain>.md        one file per domain, index table at the top
```

Two habits keep a reference set usable:

- Every index row carries its status: verified against the real system, or
  taken from the document and not yet checked. Mark unverified rows as
  unverified, so the reader knows which rows to trust.
- Say which values were measured and which were inferred. A value read off the
  running system is evidence; a value derived from another value is not. When a
  later reader hits a mismatch, this tells them whether they found drift or a
  guess.

Do not paste large material into a request or a task spec. Put it here and cite
it from there. A spec that inlines a whole API's documentation is hard to read as a spec.

> Related but not here: domain knowledge the *agent* needs at runtime (status-code semantics,
> cross-system entity mapping, aggregation conventions) belongs in a skill under
> `assets/skills/<skill>/SKILL.md`, because it has to be synced into the platform to be usable.
> `references/` is for humans and spec-writing agents; `assets/skills/` is for the running agent.
