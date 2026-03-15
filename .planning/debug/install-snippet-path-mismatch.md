---
status: investigating
trigger: "Diagnose the root cause of a UAT gap. Find root cause only, do not implement fixes."
created: 2026-03-15T00:00:00Z
updated: 2026-03-15T00:00:00Z
---

## Current Focus

hypothesis: `snippet` and `install` share the JSON contract renderer but intentionally source the executable path differently, causing `go run` to preview an ephemeral build-cache binary while `snippet` keeps a placeholder path.
test: Confirm the shared repository contract helper and reproduce `go run ./cmd/optimusctx install ...` output to verify that `os.Executable()` resolves to the Go build cache path at runtime.
expecting: Both paths will use `repository.NewServeCommand`, but `snippet` will pass a hardcoded placeholder and `install` under `go run` will emit a cache path from `os.Executable()`.
next_action: finalize diagnosis from the confirmed binary-path source mismatch and test expectations

## Symptoms

expected: Snippet/install should present one stable contract that a user can actually reuse.
actual: User reported: snippet prints `/absolute/path/to/optimusctx` while `go run ./cmd/optimusctx install --client claude-desktop` previews a Go build cache path under `/home/nico/.cache/go-build/.../optimusctx`.
errors: None reported
reproduction: Test 7 in Phase 05 UAT
started: Discovered during UAT

## Eliminated

## Evidence

- timestamp: 2026-03-15T00:00:00Z
  checked: .planning/phases/05-mcp-serving-and-integration-contracts/05-UAT.md
  found: UAT gap 7 reports `snippet` showing a placeholder path and `go run ... install` preview showing a Go build cache path.
  implication: The user-visible mismatch is reproduced at the contract-preview layer, not as a write-path side effect.

- timestamp: 2026-03-15T00:00:00Z
  checked: internal/app/snippet.go
  found: `SnippetGenerator.Render` merges the client config with `repository.NewServeCommand("/absolute/path/to/optimusctx")`.
  implication: The snippet path is intentionally hardcoded to a placeholder instead of deriving a reusable binary contract.

- timestamp: 2026-03-15T00:00:00Z
  checked: internal/cli/install.go and internal/app/install.go
  found: `runInstallCommand` fills an empty `BinaryPath` by calling `installExecutablePath`, which defaults to `os.Executable`, and `InstallService.Preview` passes that value into `repository.NewServeCommand(request.BinaryPath)`.
  implication: Install preview uses the concrete current process path, so `go run` will surface the temporary compiled binary location.

- timestamp: 2026-03-15T00:00:00Z
  checked: internal/cli/install_test.go and internal/app/snippet_test.go
  found: Tests assert `snippet` must contain `/absolute/path/to/optimusctx`, while install tests only compare the `args` contract and separately accept caller-specific command paths.
  implication: The mismatch is codified by tests, so the divergence is deliberate in the current implementation rather than an accidental regression.

- timestamp: 2026-03-15T00:00:00Z
  checked: internal/repository/client_config.go plus a live `go run ./cmd/optimusctx install --client claude-desktop --config /tmp/optimusctx-uat-claude.json`
  found: Both flows render through the same `NewServeCommand` and JSON merge helpers, and the live preview emitted `/home/nico/.cache/go-build/.../optimusctx` as the command path.
  implication: The contract drift is not in JSON rendering; it comes from feeding different command values into the shared renderer, with `go run` making the install default explicitly ephemeral.

## Resolution

root_cause: `snippet` and `install` do not share one stable executable-path contract. They only share the JSON rendering helper. `snippet` hardcodes `/absolute/path/to/optimusctx` in `SnippetGenerator.Render`, while `install` defaults `BinaryPath` from `os.Executable()` in the CLI layer. When invoked through `go run`, `os.Executable()` points at the temporary Go build-cache binary under `/home/nico/.cache/go-build/...`, so install preview emits an ephemeral path that differs from the snippet placeholder.
fix:
verification:
files_changed: []
