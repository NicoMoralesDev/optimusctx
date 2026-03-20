# Phase 32 Context

Goal: make Homebrew and Scoop first publication to empty downstream repos actually commit and push the generated file, and make downstream publication status truthful.

Observed trigger:

- The `v1.3.5` release run reported Homebrew and Scoop as `published`.
- Direct verification showed `niccrow/homebrew-tap` and `niccrow/scoop-bucket` still contained only their initial README commits.
- Workflow logs showed the generated files were rendered, but `git diff --quiet -- <path>` exited cleanly because the files were still untracked in empty repos.

Key decisions:

- Move downstream repo update behavior into a reusable script so first-publish semantics can be tested directly.
- Detect pending publication through `git status --porcelain --untracked-files=all -- <path>` instead of `git diff --quiet`.
- Reserve `publication_status=published` for real downstream repo writes and use `already_current` for truthful no-op reruns.
