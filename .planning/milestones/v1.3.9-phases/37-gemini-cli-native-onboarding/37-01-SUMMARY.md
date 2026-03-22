# Summary 37-01: Gemini CLI Config and Guidance Writes

## Completed

- Added `gemini-cli` as a first-class supported host with explicit JSON config, repo/shared scope, managed guidance, and usage-evidence capabilities.
- Implemented merge-safe Gemini `settings.json` rendering and writes that preserve unrelated settings while adding or updating the `optimusctx` MCP entry.
- Wired repo-local and shared `GEMINI.md` guidance generation through the existing managed-block merge so repeated writes preserve user content.

## Files Changed

- `internal/repository/client_config.go`
- `internal/repository/gemini_config.go`
- `internal/repository/agent_guidance.go`
- `internal/app/install.go`
- `internal/cli/init_prompt.go`
- `internal/cli/onboarding_output.go`

## Outcome

`optimusctx init --client gemini-cli` can now preview and write the native Gemini CLI contract without forcing operators to translate from generic JSON snippets or lose existing Gemini settings.
