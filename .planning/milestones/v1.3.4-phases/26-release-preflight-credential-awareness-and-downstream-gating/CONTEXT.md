# Phase 26 Context

Goal: make `release prepare` truthful about downstream publication readiness before a tag push.

Observed trigger:

- `v1.3.2` and `v1.3.3` both reported successful release workflows while Homebrew and Scoop were actually skipped because repo secrets were missing.
- The pre-existing `release prepare` surface could not tell the operator that before tagging.

Key decisions:

- Use the GitHub repository itself as the source of truth for Homebrew/Scoop secret presence when `gh` can inspect it.
- Keep `release prepare` non-mutating and compatible with selective-channel review.
- Treat missing downstream publication secrets as blockers for the selected all-channel plan, not as a footnote discovered after the tag.
