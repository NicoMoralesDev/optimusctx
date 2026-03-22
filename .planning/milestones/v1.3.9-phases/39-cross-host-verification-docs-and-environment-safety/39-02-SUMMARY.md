# Summary 39-02: Cross-Host Regression Coverage

## Completed

- Added explicit shared-path resolver coverage for Gemini CLI and Cursor CLI.
- Added repeated-write/idempotence coverage for the new JSON-backed hosts so server entries and managed guidance cannot silently duplicate.
- Re-ran the focused repository/app/CLI suite after the doc and test changes to keep the support story backed by executable evidence.

## Files Changed

- `internal/app/install_test.go`

## Outcome

Gemini CLI and Cursor CLI support claims are now backed by direct resolver and idempotence checks instead of only by broader onboarding behavior.
