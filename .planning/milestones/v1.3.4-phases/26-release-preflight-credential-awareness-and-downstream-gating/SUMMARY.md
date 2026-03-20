# Phase 26 Summary

Completed:

- `release prepare` now distinguishes canonical GitHub Release readiness from downstream publication-secret readiness.
- Homebrew and Scoop now become `blocked` when `gh secret list --repo <slug>` confirms the repo secrets are absent.
- The CLI output now prints channel summaries and details instead of only terse readiness labels.
- Fixed the hidden milestone parser bug so active patch milestones like `v1.3.4` still resolve the correct `v1.3.x` release lane.

Verification:

- `go test ./...`
- `go run ./cmd/optimusctx release prepare --json` in the real repo, which now reports Homebrew and Scoop blocked by missing repository secrets before tag creation.
