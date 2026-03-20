# Phase 32 Plan

- `32-01` Extract Homebrew and Scoop downstream repo update logic into a tested reusable script.
- `32-02` Fix first-publish detection so untracked generated files in empty repos are treated as pending publication.
- `32-03` Tighten workflow summaries so downstream `publication_status` distinguishes `published`, `already_current`, `not_published`, and `failed`.
- `32-04` Add regression coverage for first-file publication and already-current reruns.
