# Phase 25: Outcome-oriented results, docs, and verification - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Finish the onboarding UX refinement by making result output outcome-oriented, trimming avoidable config noise after successful writes, and aligning the public docs to the shipped conversation.

</domain>

<decisions>
## Implementation Decisions

### Result Output
- Configure-now output should summarize what was configured, where it was configured, and what the operator does next.
- Review-first output should clearly frame the snippet or command as the exact change under review.
- Write-backed results should stop dumping avoidable config blocks when the important outcome is already known.

### Documentation
- README and operator guides must describe the new destination-first, intent-led conversation truthfully.
- The direct `init --client <client> [--write]` path remains documented as the explicit fallback.

### Verification
- The milestone should close with both automated regression coverage and at least one real PTY walkthrough of the interactive flow.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/cli/onboarding_output.go` owns the user-facing onboarding result contract.
- `README.md`, `docs/quickstart.md`, and `docs/install-and-verify.md` already describe the interactive onboarding flow from `v1.3.2`.

### Integration Points
- `internal/cli/onboarding_output.go` needs destination and outcome summaries.
- CLI integration tests need to lock the new result contract.
- Docs need to match the new language and destination model precisely.

</code_context>

<specifics>
## Specific Ideas

- Distinguish destination labels from raw config-path output so the operator sees both the human meaning and the exact target.
- Only show the config snippet in review-first mode.
- Keep `optimusctx run` as the canonical post-onboarding handoff.

</specifics>

<deferred>
## Deferred Ideas

- Richer post-onboarding education or examples
- Host capability validation before apply
- Registration lifecycle management

</deferred>
