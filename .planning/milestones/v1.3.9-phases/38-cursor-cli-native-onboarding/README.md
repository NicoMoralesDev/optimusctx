# Phase 38: Cursor CLI Native Onboarding

## Goal

Add truthful Cursor CLI preview, write, and verification support using Cursor's documented shared `mcp.json` contract.

## Requirements

- `CUR-01`
- `CUR-02`

## Scope

- Add `cursor-cli` as a first-class supported host with repo-local and shared `mcp.json` targets.
- Reuse the JSON merge/write path so repeated writes preserve unrelated config and avoid duplicate server entries.
- Keep support wording precise: the first-class contract is Cursor CLI onboarding, even if the config file may also be used by Cursor editor surfaces.
- Detect Cursor registration in status/doctor without implying unsupported durable guidance.

## Success Criteria

1. `optimusctx init --client cursor-cli` previews and writes the native Cursor `mcp.json` contract without hand transcription.
2. `optimusctx status` can detect repo-local or shared Cursor registration and explain the shared CLI/editor config story accurately.
3. Repeated Cursor writes preserve unrelated config and avoid duplicate `optimusctx` entries.
