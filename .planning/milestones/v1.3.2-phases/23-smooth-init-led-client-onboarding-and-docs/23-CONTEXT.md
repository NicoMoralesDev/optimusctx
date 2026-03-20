# Phase 23: Smooth init-led client onboarding and docs update - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Turn `optimusctx init` into one smoother operator entrypoint for both repository bootstrap and supported-client onboarding across the current Claude and Codex host set, while keeping explicit `--client ... [--write]` flows available for direct and non-interactive use.

</domain>

<decisions>
## Implementation Decisions

### Onboarding Flow
- Plain interactive `optimusctx init` should offer supported-client onboarding during the same invocation instead of always ending with a second-command handoff.
- The interactive path must stay optional: the operator can skip onboarding and keep the repository bootstrap result.
- Explicit direct flows (`optimusctx init --client <client> [--write]`) remain first-class for scripts, docs, and operators who already know the target host.

### Supported Clients
- `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` must all participate in the same product story.
- Claude CLI keeps its scope selection semantics.
- No new host expansion, removal flows, or capability-preflight work lands in this phase.

### Output UX
- Preview output should show only the relevant registration command or config block for the selected client, not the user's full host configuration.
- Write and preview flows should both end with one clear next step that preserves `optimusctx run` as the canonical runtime handoff.
- Existing merge-safe write behavior must stay intact even when the preview becomes more focused.

### Documentation
- README, quickstart, and install/verify docs must describe the same-command `init` onboarding path truthfully.
- The explicit `--client ... [--write]` path remains documented as the non-interactive and direct-control fallback.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/cli/init.go` already owns repository bootstrap plus the direct `--client` onboarding path.
- `internal/app/install.go` centralizes per-client preview/write behavior, including Claude CLI command execution and config-file merges.
- `internal/repository/client_config.go` and `internal/repository/codex_config.go` already provide the render/merge helpers needed to separate preview snippets from applied file content.

### Established Patterns
- CLI commands parse flags locally, then pass typed requests into the app layer.
- Preview/write behavior is regression-locked with CLI and app tests rather than screenshots or snapshots only.
- Public docs stay close to the literal command surface and are backed by release/doc tests in `internal/release`.

### Integration Points
- `internal/cli/init.go` needs the interactive same-command onboarding flow.
- `internal/app/install.go` needs focused preview rendering for any client that still leaks unrelated host config.
- `README.md`, `docs/quickstart.md`, and `docs/install-and-verify.md` need to align with the smoother `init` path.

</code_context>

<specifics>
## Specific Ideas

- Offer a short numbered selection flow only when `init` runs in an interactive terminal and no `--client` was passed.
- Reuse the same backend request shape for both the interactive path and explicit flags so there is only one supported-client registration implementation.
- Keep the terminal prompt sequence short: choose client, choose Claude CLI scope when needed, then choose preview or write.

</specifics>

<deferred>
## Deferred Ideas

- Additional first-class hosts
- Host capability preflight checks before write-backed registration
- Lifecycle/remove/manage commands for existing host registrations
- Any TUI or full-screen onboarding wizard

</deferred>
