# A credential the caller supplies, per turn

The agent calls a customer API **as the person talking to it**, using a
short-lived credential the caller puts in the request rather than one the chart
holds. The query scope is fixed by the credential, not by the prompt.

**Seen in:** a commerce back-office where the front end passes a user JWT and an
optional per-brand API token on every turn, and a public support widget where the
site forwards a scope-limited token so the agent can read that customer's own
orders and nothing else.

**Checked:** 2026-09-02: written directly from a supervisor's SandboxBlueprint hook, whose own comment records the incident that shaped it. Checked against the CRD's hook events.

**Unchecked:** the public-widget variant, which is described in a case study rather than read out of a chart.

**Read the platform side first:** `asgard-cli wiki api` -
the endpoint, the SSE event sequence, and the four integration patterns. This page assumes you have.

## When this shape, and when not

Use it when **which rows the agent may see depends on who is asking**. A customer
service agent reading "your orders", an internal tool acting with the operator's
own permissions, anything multi-tenant where the tenant is decided per request.

The alternative that looks simpler is a semantic layer or a query tool holding one
service credential, with the prompt told to filter by the caller's id. **Do not do
that for per-user data.** The model then chooses which id to query, and a model
that chooses can be argued into choosing differently. Here the scope is enforced
one layer down: the credential itself cannot see anything else.

Do **not** use it when:

- the credential belongs to the *service* rather than to a user - a static API key
  or an OAuth client-credentials token is `asgard-cli usecase api-oauth`
- nothing about the answer depends on who asked. A public catalogue does not need
  per-turn identity, and adding it means the front end now has to hold something

## The shape

    caller  ->  BotProvider payload   identity + token, every turn
      -> Workflow                     prevPayload.* carries them
        -> SandboxBlueprint
             hooks  user-prompt-submit    writes them into the sandbox filesystem
      -> a SkillSet that reads that file and calls the API

The credentials never enter a CR spec, never reach a Secret, and are not part of
what helm deploys. They exist for the length of one turn.

## Generate it

There is no generator kind for this - it is a hook expression on a blueprint the
`flowagent` or supervisor generator already wrote. Add it to an existing
`SandboxBlueprint`.

## The skeleton

```yaml
spec:
  hooks:
    expression: |-
      (() => {
        const cfg = {
          api_base_url: {{ .Values.<name>ApiBaseUrl | quote }},
          tenant: prevPayload.tenant || null,
          user: prevPayload.user
            ? {
                id: prevPayload.user.id,
                auth: prevPayload.user.auth?.access_token
                  ? { access_token: prevPayload.user.auth.access_token }
                  : null,
              }
            : null,
        };
        const json = JSON.stringify(cfg, null, 2);
        return [{
          event: "user-prompt-submit",
          handlerType: "command",
          commandHandler: {
            command: `umask 077; cat > /tmp/.cfg.json.tmp <<'EOF'\n${json}\nEOF\nmv /tmp/.cfg.json.tmp /tmp/cfg.json`
          }
        }];
      })()
```

## Fields that are not obvious

### It must be `user-prompt-submit`, not `session-start`

This is the whole reason the shape is written this way, and it cost a live
incident to find.

A `session-start` hook is **part of the Sandbox CR spec**. A JWT carries `jti` and
`iat`, so its string changes on every issue; a hook whose content changes bumps
the CR's generation, and a generation bump **recreates the pod - mid-conversation**.

`user-prompt-submit` is re-evaluated by the driver from that turn's payload and
delivered with the task. It never enters the spec, so nothing is recreated. It
also means the token is fresh every turn, which incidentally fixed a separate
problem: a token that had been captured once went stale after eight hours.

The two remaining hook events are not an option either: `pre-tool-call` and
`post-tool-call` were never implemented and declaring one is a silent no-op.

### Three details in the write command, each for a reason

```
umask 077; cat > /tmp/.cfg.json.tmp <<'EOF' ... EOF; mv ... /tmp/cfg.json
```

- **`umask 077`** so the file holding a token is mode 600.
- **A quoted heredoc plus `JSON.stringify`** so token contents cannot break out
  into the shell. A raw interpolation is an injection waiting for a token with a
  quote in it.
- **Write a temp file and `mv`** - the replacement has to be atomic. A user who
  sends a second message mid-run triggers a rewrite, and a tool already running
  would otherwise read half a file.

### Optional fields become `null`, and the reader must expect it

A payload omits what does not apply - no tenant, no connected brand, an
unauthenticated visitor. Write `null` rather than omitting the key, and say in
the consuming skill that `null` is a normal value. A skill that assumes the field
is present fails on the anonymous path only, which is the path nobody tests.

### The prompt still has a job, and it is not enforcement

Tell the agent what it may do when the credential is absent - and make that
branch explicit, because "no token" is a state, not an error:

    <<if logged in>> you may look up this customer's own orders through the
    API, using the connection details injected this turn. Never guess a URL or
    rewrite the host.
    <<else>>       you cannot look up any order. Say that signing in is
    required, and do not try another route.

Naming the injected values as the only ones it may use is worth writing, but it
is a guardrail on top of the real one. If the prompt were the only control, the
shape would not be worth the hooks.

## Verify

`asgard-cli check` and `verify` see none of this - it is a JavaScript expression
in a string field, and no schema validates its contents.

- Render the chart and read the hook back. An expression that throws is a runtime
  failure on the first turn, not a deploy failure.
- Send one turn with the credential and one without, and confirm the second is
  refused rather than answered from a stale file.
- Confirm the file is mode 600 in the sandbox, and that a second message
  mid-run does not corrupt it.

## Source

- Written from a commerce back-office supervisor's `SandboxBlueprint`, whose
  hook comment records the 2026-08-21 pod-recreation incident that moved it off
  `session-start`.
- The public-widget variant is the same mechanism reached from the other side: the
  site forwards a scope-limited token instead of a user JWT.
- `SandboxHookEvent` and the deprecation of `pre-tool-call` / `post-tool-call`
  are from the platform CRD.
