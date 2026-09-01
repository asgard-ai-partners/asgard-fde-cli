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

> Related but **not** here: domain knowledge the *agent* needs at runtime (status-code semantics,
> cross-system entity mapping, aggregation conventions) belongs in a skill under
> `common/skills/<skill>/SKILL.md`, because it has to be synced into the platform to be usable.
> `references/` is for humans and spec-writing agents; `common/skills/` is for the running agent.
