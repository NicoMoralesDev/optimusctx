# Phase 37: Gemini CLI Native Onboarding

## Goal

Add truthful Gemini CLI preview, write, and verification support using Gemini's documented `settings.json` and `mcpServers` model.

## Requirements

- `GEM-01`
- `GEM-02`

## Scope

- Add `gemini-cli` as an explicit supported client.
- Support repo-local `.gemini/settings.json` and shared `~/.gemini/settings.json`.
- Preserve unrelated Gemini settings while merging `mcpServers`.
- Register managed guidance into default `GEMINI.md` locations for repo and shared scopes.
- Detect Gemini CLI registration and guidance from `optimusctx status`.

## Success Criteria

1. `optimusctx init --client gemini-cli` can preview and write the native Gemini contract without manual translation.
2. `optimusctx status` can detect Gemini CLI registration and distinguish discovery from actual tool use when evidence exists.
3. Repeated Gemini writes preserve unrelated config and avoid duplicate server entries.
