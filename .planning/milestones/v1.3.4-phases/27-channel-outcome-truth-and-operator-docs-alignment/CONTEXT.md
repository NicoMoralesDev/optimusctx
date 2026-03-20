# Phase 27 Context

Goal: make post-release truth as explicit as pre-release truth.

Observed trigger:

- The workflow reported `skipped` for Homebrew and Scoop, but that was easy to misread as a harmless CI detail instead of "channel did not publish".

Key decisions:

- Add one explicit `publication_status` field to workflow summaries.
- Use `not_published` when credentials are absent so the operator cannot confuse that outcome with success.
- Align release docs to the exact workflow contract instead of softer wording.
