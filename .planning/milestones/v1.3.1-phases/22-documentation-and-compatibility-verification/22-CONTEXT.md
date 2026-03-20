# Phase 22: Documentation and Compatibility Verification - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Update the public and release-facing documentation so it matches the corrected init-led supported-client onboarding contract, then close the milestone with regression and operator evidence that the supported clients still work with `optimusctx run`.

</domain>

<decisions>
## Implementation Decisions

### Documentation Contract
- Public docs must treat `optimusctx init --client <client> [--write]` as the canonical supported-client preview/write surface.
- Public docs must describe `optimusctx status` as a read-only runtime/health command.
- Examples should continue to name the four supported native clients explicitly and preserve `optimusctx run` as the runtime handoff.

### Verification Shape
- Regression evidence should reuse existing Go test coverage wherever possible instead of inventing new ad hoc verification tools.
- Release/operator docs and the internal release distribution policy must agree on the init-led contract and the read-only status surface.
- Real Claude CLI write validation stays a named manual evidence item if the `claude` binary is unavailable in this environment.

### Scope Guardrails
- This phase updates docs, tests, and release/operator evidence only; it does not reopen the command contract established in Phase 21.1.
- No new MCP hosts, onboarding modes, or release channels are added here.

### Claude's Discretion
The exact split between docs edits, release-policy test updates, and evidence artifacts is at Claude's discretion as long as the final milestone story is truthful and verifiable.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/cli/init.go`, `internal/cli/status.go`, `internal/app/snippet.go`, and `internal/app/doctor.go` now define the authoritative onboarding/diagnostic contract.
- `internal/release/distribution_plan.go` and `internal/release/distribution_plan_test.go` already lock release-facing doc and policy language for the public distribution story.
- Phase 21 and 21.1 verification artifacts already contain concrete evidence for Codex host consumption and the remaining Claude CLI manual gap.

### Established Patterns
- Release/operator docs are kept in sync through exact-string assertions in `internal/release/distribution_plan_test.go`.
- CLI/operator behavior changes are verified with normal `go test ./...` coverage plus one explicit verification report per phase.
- User-facing docs prefer short copy-paste command blocks with the canonical command surface rather than prose-only descriptions.

### Integration Points
- `README.md`, `docs/quickstart.md`, and `docs/install-and-verify.md` must align with the init-led onboarding contract.
- `docs/distribution-strategy.md`, `docs/operator-release-guide.md`, `docs/release-checklist.md`, and `internal/release/distribution_plan*.go` must align for release operators.
- Phase 22 verification should combine regression runs with a local operator walkthrough using the real `codex` binary where available.

</code_context>

<specifics>
## Specific Ideas

- Keep examples concrete by showing Claude Desktop onboarding plus at least one Codex example.
- Make the docs explicit that `status` is still part of the verification path, but not the registration owner.
- Carry the real Claude CLI validation item forward as an explicit manual check instead of pretending the environment proved it.

</specifics>

<deferred>
## Deferred Ideas

- Additional host-capability preflight, uninstall/remove flows, and new MCP hosts remain out of scope.
- Any future release-channel expansion stays deferred beyond this milestone.

</deferred>
