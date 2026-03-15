# Phase 05 Verification: MCP Serving and Integration Contracts

## Scope and Verification Standard

This report backfills the missing milestone-grade verification artifact for Phase 05.
It verifies the current shipped MCP contract for `CLI-02` and `MCP-01` through `MCP-04`
from current implementation and current automated evidence, rather than reusing plan
chronology as proof.

Phase 05 verification is requirement-driven:

- `MCP-01`: stdio MCP serving is the primary integration surface and exposes the
  complete current tool contract.
- `MCP-02`: tool results are machine-readable and include freshness and
  cache-versus-refresh metadata.
- `MCP-03`: the MCP surface includes repository map, context, lookup, token tree,
  refresh, health, and pack capabilities.
- `MCP-04`: handlers enforce bounded defaults and return explicit actionable
  failures.
- `CLI-02`: install registration is optional, explicit, consent-gated, and aligned
  with the same `optimusctx mcp serve` contract shown by `optimusctx snippet`.

The canonical current verification command path for this milestone backfill is
`/usr/local/go/bin/go`. Historical summary-era references to `/tmp/optimusctx-go`
remain useful as development history, but they are not treated as current proof in
this verification artifact.

## Evidence Inputs

This report draws from:

- `.planning/phases/05-mcp-serving-and-integration-contracts/05-01-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-02-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-03-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-04-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-05-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-06-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-07-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-08-SUMMARY.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-VALIDATION.md`
- `.planning/phases/05-mcp-serving-and-integration-contracts/05-UAT.md`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `internal/mcp/integration_test.go`
- `internal/mcp/query_tools.go`
- `internal/mcp/query_tools_test.go`
- `internal/app/token_tree_test.go`
- `internal/app/health_pack_test.go`
- `internal/cli/install_test.go`
- `internal/cli/mcp_test.go`
- `internal/app/snippet.go`
- `internal/app/snippet_test.go`
- `internal/repository/client_config.go`

## Requirement-Level Evidence Matrix

### CLI-02: explicit supported-client registration with consent

| Verification target | Current evidence anchors | Why the evidence is sufficient |
| --- | --- | --- |
| Registration remains optional and preview-first | `05-06-SUMMARY.md`, `internal/cli/install_test.go::TestInstallRegistrationDryRun` | Dry-run output proves preview mode is the default and no config file is created without explicit write mode. |
| Writes require explicit operator consent | `05-06-SUMMARY.md`, `internal/cli/install_test.go::TestInstallRegistrationConsent` | The write-path test shows config updates only occur behind `--write`, preserving explicit consent semantics. |
| Unsupported clients fail transparently | `05-06-SUMMARY.md`, `internal/cli/install_test.go::TestInstallCommandRejectsUnsupportedClient` | Unsupported-client rejection is asserted at the real CLI command boundary with a direct error. |
| Snippet and install previews share one MCP contract | `05-06-SUMMARY.md`, `05-08-SUMMARY.md`, `internal/app/snippet_test.go::TestSnippetInstallCommandAlignment`, `internal/cli/install_test.go::TestInstallRegistrationDryRun` | Shared command rendering is proven both at the app snippet level and the install preview command boundary. |
| Omitted `--binary` never persists an ephemeral runtime path | `05-08-SUMMARY.md`, `internal/cli/install_test.go::TestInstallNormalizesEphemeralExecutablePath`, `internal/cli/install_test.go::TestInstallWriteNormalizesEphemeralExecutablePath` | Preview and write flows both normalize unstable runtime paths onto the reusable `optimusctx mcp serve` contract. |

### MCP-01: stable stdio MCP serving as the primary integration mode

| Verification target | Current evidence anchors | Why the evidence is sufficient |
| --- | --- | --- |
| `optimusctx mcp serve` exists as the real CLI entrypoint | `05-01-SUMMARY.md`, `internal/cli/mcp_test.go::TestMCPServeCommand` | The real root command path delegates to the MCP server entrypoint and preserves stdout for transport output. |
| The server speaks over stdio, not an alternate transport | `05-01-SUMMARY.md`, `internal/mcp/server.go`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | The server reads framed input from stdin and writes framed responses to stdout in a real session test. |
| Tool discovery is deterministic and truthful | `05-01-SUMMARY.md`, `05-05-SUMMARY.md`, `internal/mcp/server_test.go::TestMCPServerBasicSession`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | `tools/list` is exercised against the live registry and validated for deterministic exposure of the shipped tool surface. |
| Unknown or unimplemented tool slots fail explicitly | `05-01-SUMMARY.md`, `internal/mcp/server_test.go::TestMCPServerRejectsUnknownTool`, `internal/mcp/server_test.go::TestMCPServerRejectsUnimplementedTool` | Server-boundary error handling is proven with structured errors rather than silent success or panics. |
| Manual readiness does not corrupt MCP stdout framing | `05-07-SUMMARY.md`, `internal/cli/mcp_test.go::TestMCPServeReadinessSignalUsesStderr`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | Readiness is emitted on stderr only, while the integration test confirms stdout stays reserved for framed MCP responses. |

### MCP-02: structured machine-readable payloads with freshness metadata

| Verification target | Current evidence anchors | Why the evidence is sufficient |
| --- | --- | --- |
| Read-only query tools return a shared structured envelope | `05-02-SUMMARY.md`, `internal/mcp/query_tools.go`, `internal/mcp/query_tools_test.go::TestMCPRepositoryQueries` | Query handlers wrap app-layer results in shared structured content instead of ad hoc transport payloads. |
| Freshness and cache-versus-refresh metadata remain consistent | `05-02-SUMMARY.md`, `05-05-SUMMARY.md`, `internal/mcp/query_tools_test.go::TestMCPRepositoryQueries`, `internal/mcp/integration_test.go::TestMCPRefreshPackHealth` | Repository, refresh, pack, and health results all expose metadata through one envelope contract. |
| Refresh reports refresh-attempt semantics rather than persisted-only cache semantics | `05-05-SUMMARY.md`, `internal/mcp/integration_test.go::TestMCPRefreshPackHealth`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | Integration tests assert that refresh calls set `cacheStatus` to `refresh_attempted`. |
| Structured content survives real stdio routing | `05-05-SUMMARY.md`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | The stdio session test decodes structured content after full framing and routing, proving the transport contract end to end. |

### MCP-03: complete promised MCP capability surface

| Verification target | Current evidence anchors | Why the evidence is sufficient |
| --- | --- | --- |
| Repository map, layered context, symbol lookup, structure lookup, and targeted context are exposed | `05-02-SUMMARY.md`, `internal/mcp/query_tools.go`, `internal/mcp/query_tools_test.go::TestMCPRepositoryQueries`, `internal/mcp/query_tools_test.go::TestMCPLookupQueries` | The read-only tool set is registered and exercised with real request normalization and structured payload decoding. |
| Token tree capability is present and deterministic | `05-03-SUMMARY.md`, `internal/app/token_tree_test.go::TestTokenTree`, `internal/app/token_tree_test.go::TestTokenTreeBounds`, `internal/mcp/integration_test.go::TestMCPRefreshPackHealth` | The persisted token-tree service and MCP exposure are verified for hierarchy, determinism, and bounded output. |
| Health and pack capabilities are present and machine-readable | `05-04-SUMMARY.md`, `internal/app/health_pack_test.go::TestHealthService`, `internal/app/health_pack_test.go::TestPackService`, `internal/mcp/integration_test.go::TestMCPRefreshPackHealth` | Health and pack are verified first at the service layer and again through the MCP surface. |
| Refresh is part of the same shipped registry | `05-05-SUMMARY.md`, `internal/mcp/integration_test.go::TestMCPRefreshPackHealth`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | Refresh is not a separate operator-only seam; it is exercised as part of the current MCP registry. |
| The full tool surface is discoverable through `tools/list` | `05-05-SUMMARY.md`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | The server-boundary session asserts the presence of operational tools such as refresh, pack, and health in the advertised registry. |

### MCP-04: bounded defaults and explicit actionable failures

| Verification target | Current evidence anchors | Why the evidence is sufficient |
| --- | --- | --- |
| Query handlers enforce field-specific bounds before service calls | `05-02-SUMMARY.md`, `internal/mcp/query_tools.go`, `internal/mcp/query_tools_test.go::TestMCPBoundedFailures` | Bounds normalization and validation happen at the MCP boundary and are covered by dedicated failure tests. |
| Structured failures identify the offending field | `05-02-SUMMARY.md`, `internal/mcp/query_tools_test.go::TestMCPStructuredErrors` | Failure payloads are asserted to carry machine-readable field and constraint details. |
| Token tree request bounds stay explicit | `05-03-SUMMARY.md`, `internal/app/token_tree_test.go::TestTokenTreeBounds`, `internal/mcp/integration_test.go::TestMCPServerStdioSession` | Oversized `maxNodes` requests produce explicit bounded failures instead of silent truncation. |
| Pack request scope remains bounded and deterministic | `05-04-SUMMARY.md`, `internal/app/health_pack_test.go::TestPackService` | Oversized pack section requests fail with a clear max-sections error while healthy results remain deterministic across repeated reads. |
| Unknown or missing MCP surface requests fail transparently | `05-01-SUMMARY.md`, `internal/mcp/server_test.go::TestMCPServerRejectsUnknownTool`, `internal/mcp/server_test.go::TestMCPServerRejectsUnimplementedTool` | The transport contract preserves explicit failures for unsupported registry states. |

## Current Command Truth

The current automated proof command for this report is:

```sh
env GOCACHE=/tmp/optimusctx-gocache \
  GOMODCACHE=/home/nico/go/pkg/mod \
  GOPROXY=off \
  /usr/local/go/bin/go test ./... -run 'TestMCPServeCommand|TestMCPServeReadinessSignalUsesStderr|TestMCPServerBasicSession|TestMCPServerRejectsUnknownTool|TestMCPServerRejectsUnimplementedTool|TestMCPRepositoryQueries|TestMCPLookupQueries|TestMCPBoundedFailures|TestMCPStructuredErrors|TestTokenTree|TestTokenTreeBounds|TestHealthService|TestPackService|TestMCPToolRegistry|TestMCPRefreshPackHealth|TestMCPServerStdioSession|TestSnippetGeneratorRender|TestSnippetInstallCommandAlignment|TestInstallRegistrationDryRun|TestInstallRegistrationConsent|TestInstallNormalizesEphemeralExecutablePath|TestInstallWriteNormalizesEphemeralExecutablePath|TestInstallCommandRejectsUnsupportedClient'
```

This command is intentionally written with the current `/usr/local/go/bin/go` path and
the existing offline module cache at `/home/nico/go/pkg/mod`. The cold-cache variant
that points `GOMODCACHE` at `/tmp/optimusctx-gomodcache` is not used as current proof
because `GOPROXY=off` prevents dependency resolution there when the cache starts
empty.

## Verified Current Behavior

### MCP-01 Passed: stdio serving and server-boundary contract

The current implementation in `internal/mcp/server.go` still exposes one stdio-based
server loop that reads header-framed requests from stdin and writes header-framed
JSON-RPC responses to stdout. `ServeStdio` remains the canonical server entrypoint,
and `internal/cli/mcp_test.go::TestMCPServeCommand` proves the CLI `optimusctx mcp
serve` path delegates to that server boundary without leaking output onto stdout.

`internal/mcp/server_test.go::TestMCPServerBasicSession` proves the basic session
contract from initialize through `tools/list`, while
`internal/mcp/integration_test.go::TestMCPServerStdioSession` verifies the end-to-end
stdio flow against a live repository-backed session. That test confirms:

- initialize succeeds over framed stdio transport
- `tools/list` exposes the actual tool registry
- repository-map and refresh requests can run through the same session
- invalid token-tree bounds fail with structured MCP errors

The readiness gap closed in Phase 05 is still verified in current code. The server
emits `optimusctx mcp: ready for stdio requests` on stderr, and both
`TestMCPServeReadinessSignalUsesStderr` and `TestMCPServerStdioSession` assert that
the readiness string stays off stdout. That matters because stdout is reserved for MCP
framing bytes; any operator-facing output there would corrupt the protocol stream.

Verdict: `MCP-01` is currently satisfied by a real stdio command path, deterministic
tool discovery, and server-boundary error handling that is proven by current tests.

### MCP-02 Passed: machine-readable envelopes with freshness metadata

The current MCP query layer in `internal/mcp/query_tools.go` still adapts app-layer
results into one shared structured envelope rather than creating per-tool prose
payloads. The evidence from `internal/mcp/query_tools_test.go::TestMCPRepositoryQueries`
shows repository queries carry machine-readable metadata for repository root,
generation, freshness, cache status, and bounds.

The current integration coverage also shows that refresh-style tools do not pretend to
be persisted-only reads. `internal/mcp/integration_test.go::TestMCPRefreshPackHealth`
asserts that refresh responses set `cacheStatus` to `refresh_attempted`, while pack
and health results continue to expose the same envelope structure. This is important
because Phase 05 promised not just structured payloads, but structured payloads whose
metadata correctly describes how the result was produced.

`internal/mcp/integration_test.go::TestMCPServerStdioSession` closes the last gap by
decoding structured content after the full stdio routing path. That makes the proof
stronger than unit-level adaptation checks alone: the structured envelope survives the
full MCP framing, dispatch, and response path used by real clients.

Verdict: `MCP-02` is currently satisfied. The live tool surface returns
machine-readable envelopes with freshness and cache-status metadata, and the metadata
changes appropriately for refresh-oriented calls.

### MCP-03 Passed: complete shipped tool surface

Phase 05 promised a complete MCP surface spanning repository map, layered context,
lookup, targeted context, token tree, refresh, health, and pack. The current evidence
shows those capabilities are present in one registry rather than split across
unrelated seams.

`internal/mcp/query_tools.go` defines the read-only capability set for repository map,
L0 context, L1 context, symbol lookup, structure lookup, and targeted context. The
tests `TestMCPRepositoryQueries` and `TestMCPLookupQueries` cover the query side of
that registry. `internal/app/token_tree_test.go::TestTokenTree` and
`TestTokenTreeBounds` prove the token-tree service still returns deterministic
hierarchical results with explicit truncation behavior. `internal/app/health_pack_test.go`
shows health and pack remain transport-neutral app services with deterministic outputs
before the transport adapter is applied.

The integration proof in `internal/mcp/integration_test.go::TestMCPRefreshPackHealth`
demonstrates that refresh, token tree, health, and pack are all callable through the
current MCP surface. `TestMCPServerStdioSession` then verifies that the same
operational tools appear in `tools/list`, confirming that discovery and invocation
agree on the shipped surface.

Verdict: `MCP-03` is currently satisfied. The current server exposes the full promised
Phase 05 capability set through one discoverable MCP contract.

### MCP-04 Passed: bounded defaults and structured failures

Phase 05 also promised that the MCP surface would stay bounded and fail transparently.
The current code still enforces that discipline at the transport edge. In
`internal/mcp/query_tools.go`, repository-map and other query handlers normalize
limits before delegating to app services. `internal/mcp/query_tools_test.go::TestMCPBoundedFailures`
and `TestMCPStructuredErrors` verify that failures carry field-specific metadata
instead of generic transport errors.

The token-tree and pack services preserve that same bounded behavior. On the service
side, `internal/app/token_tree_test.go::TestTokenTreeBounds` proves that token-tree
responses remain deterministic and explicitly marked as truncated when appropriate.
At the MCP boundary, `internal/mcp/integration_test.go::TestMCPServerStdioSession`
asserts that oversized `maxNodes` input fails with a structured bounds error naming
the offending field. `internal/app/health_pack_test.go::TestPackService` similarly
proves pack requests reject oversized section counts rather than silently stretching
scope.

The base server transport also keeps failure behavior explicit for unsupported tool
states. `TestMCPServerRejectsUnknownTool` and `TestMCPServerRejectsUnimplementedTool`
show that missing or placeholder tools do not produce false-positive success results.

Verdict: `MCP-04` is currently satisfied. Bounds are enforced at the MCP edge and the
resulting failures are explicit, machine-readable, and actionable.

### CLI-02 Passed: consent-gated registration and snippet parity

Phase 05 made install registration an optional wedge, not a silent mutation path.
That contract still holds in the current CLI and snippet code. `internal/app/snippet.go`
renders the manual integration snippet using the same `repository.NewServeCommand("")`
helper that install uses to build client configuration, and
`internal/app/snippet_test.go::TestSnippetGeneratorRender` plus
`TestSnippetInstallCommandAlignment` confirm the snippet advertises the canonical
`optimusctx mcp serve` contract rather than placeholder or stale command text.

At the command boundary, `internal/cli/install_test.go::TestInstallRegistrationDryRun`
proves preview mode is the default and that no config file is written during dry-run.
`TestInstallRegistrationConsent` proves writes happen only with `--write`, preserving
the explicit-consent requirement. `TestInstallCommandRejectsUnsupportedClient` proves
unsupported targets fail transparently rather than creating ambiguous config output.

The executable-path drift closed in Phase 05 also remains covered. Both
`TestInstallNormalizesEphemeralExecutablePath` and
`TestInstallWriteNormalizesEphemeralExecutablePath` prove that omitted `--binary`
flows normalize unstable runtime paths onto the canonical reusable `optimusctx`
command. That keeps install preview, install write, and snippet output aligned around
the same serve command and prevents a `go run` cache binary from being written into
client configuration.

Verdict: `CLI-02` is currently satisfied. Registration is preview-first, explicit,
consent-gated, and contract-aligned with the manual snippet guidance.

## Automated Verification Run

This backfill relies on the following current test anchors. The full targeted run
passed with `/usr/local/go/bin/go`, `GOCACHE=/tmp/optimusctx-gocache`,
`GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off`.

- `TestMCPServeCommand`
- `TestMCPServeReadinessSignalUsesStderr`
- `TestMCPServerBasicSession`
- `TestMCPServerRejectsUnknownTool`
- `TestMCPServerRejectsUnimplementedTool`
- `TestMCPRepositoryQueries`
- `TestMCPLookupQueries`
- `TestMCPBoundedFailures`
- `TestMCPStructuredErrors`
- `TestTokenTree`
- `TestTokenTreeBounds`
- `TestHealthService`
- `TestPackService`
- `TestMCPToolRegistry`
- `TestMCPRefreshPackHealth`
- `TestMCPServerStdioSession`
- `TestSnippetGeneratorRender`
- `TestSnippetInstallCommandAlignment`
- `TestInstallRegistrationDryRun`
- `TestInstallRegistrationConsent`
- `TestInstallNormalizesEphemeralExecutablePath`
- `TestInstallWriteNormalizesEphemeralExecutablePath`
- `TestInstallCommandRejectsUnsupportedClient`

These tests collectively prove:

- the current CLI entrypoint exists and delegates correctly
- the server preserves a clean stdio transport contract
- the live registry exposes the shipped query and operational tools
- structured machine-readable envelopes survive live routing
- bounds failures stay explicit and field-specific
- token tree, health, and pack remain deterministic and bounded
- install registration stays preview-first, consent-gated, and aligned with snippet

## Final Verdict

Phase 05 now has current milestone-grade verification evidence.

- `CLI-02`: passed
- `MCP-01`: passed
- `MCP-02`: passed
- `MCP-03`: passed
- `MCP-04`: passed

The current implementation and tests support one coherent MCP product story:
`optimusctx mcp serve` is a real stdio MCP server; `tools/list` reflects the actual
registry; query and operational tools return structured machine-readable envelopes;
bounded failures are explicit; and supported-client registration is opt-in, preview
first, and aligned with the same `optimusctx mcp serve` contract shown by
`optimusctx snippet`.
