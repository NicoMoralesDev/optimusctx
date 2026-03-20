# Phase 27 Summary

Completed:

- GitHub Actions summaries for GitHub Release, npm, Homebrew, and Scoop now include `publication_status`.
- Homebrew and Scoop now say `publication_status=not_published` when their repo secret is missing.
- Operator release guide, release checklist, distribution strategy, and install docs now describe partial-publication recovery truthfully.

Verification:

- `go test ./...`
- Workflow contract assertions in `internal/release/release_test.go`
