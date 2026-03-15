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
