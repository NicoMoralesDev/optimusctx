# Phase 24: Intent-led init conversation and scope targeting - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Rework the interactive `optimusctx init` conversation so supported-client onboarding starts from user intent and destination choice, not backend terms like preview and write.

</domain>

<decisions>
## Implementation Decisions

### Conversation Model
- Plain interactive `optimusctx init` keeps one short numbered conversation.
- The operator still chooses the supported client from the same flow instead of re-running with `--client`.
- The action prompt should ask whether to configure now or review the exact change first.

### Destination Selection
- Every supported client should ask where OptimusCtx should live before mutation happens.
- The choice must show the exact config path or native registration target inline before confirmation.
- Codex should support repo-local and shared config targets in the interactive flow.
- Claude CLI should keep its native scope model but present it in operator terms.

### Control Surface
- Explicit direct control with `optimusctx init --client <client> [--write]` must remain intact for scripts and operators who want deterministic flag-based behavior.
- Interactive choices should still converge on the existing `InstallRequest` backend instead of introducing a parallel install path.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/cli/init.go` already owns both repository bootstrap and explicit supported-client onboarding entrypoints.
- `internal/cli/init_prompt.go` already contains the interactive prompt flow introduced in `v1.3.2`.
- `internal/app/install.go` centralizes default path resolution and install execution for supported clients.

### Integration Points
- `internal/cli/init_prompt.go` needs the destination-first prompt language and exact-target display.
- `internal/cli/init.go` needs to pass repository-root context through the interactive path.
- `internal/app/install.go` needs helpers for default destination resolution so the prompt can display real paths before mutation.

</code_context>

<specifics>
## Specific Ideas

- Ask for destination before action so the operator sees where OptimusCtx will live before choosing configure-now or review-first.
- Default Codex to repo-local config because that matches the safer local-first operator expectation.
- Default Claude CLI to project scope for the same reason.
- Keep the prompt copy short enough that the interaction still feels like a CLI, not a wizard.

</specifics>

<deferred>
## Deferred Ideas

- Additional first-class hosts
- Capability preflight before write-backed registration
- Registration removal or lifecycle management
- Any full-screen wizard or TUI

</deferred>
