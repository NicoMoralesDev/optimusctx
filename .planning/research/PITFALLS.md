# v1.1 Pitfalls Research

## Main Risks

### 1. Benchmarks that prove the script, not the product

Risk:

- Control and treatment paths diverge in prompts, stop conditions, or repo state, making the reported savings meaningless.

Prevention:

- Fix scenario definitions, repo snapshots, stopping rules, and metric collection before running comparisons.
- Keep baseline and OptimusCtx paths comparable enough that the only intended difference is use of the runtime.

### 2. Measuring only command latency instead of workflow value

Risk:

- A benchmark shows one command is fast, but ignores the real user question: whether the agent needed fewer broad searches, fewer file reads, or less context assembly work.

Prevention:

- Measure end-to-end workflow steps, not just single command durations.
- Separate discovery time, context assembly time, refresh-after-change time, and completion time.

### 3. Token claims that change the measuring stick mid-run

Risk:

- The milestone mixes different token estimators, different model assumptions, or different artifact scopes across runs.

Prevention:

- Use one required estimator for milestone claims: the shipped `bytes_div_4_ceiling` policy.
- If later tokenizer-specific numbers are added, report them as secondary metrics, not as the milestone truth.

### 4. Overbuilding an evaluation platform

Risk:

- v1.1 turns into a generic telemetry or analytics system instead of proving the value of the shipped runtime.

Prevention:

- Keep the scope narrow: scenario runner, fixture repos, result capture, summary/export.
- Avoid hosted dashboards, remote collection, or large reporting surfaces this milestone.

### 5. Distribution planning that silently changes the product category

Risk:

- The distribution work pulls the project toward cloud services, daemon requirements, or vendor-specific integrations.

Prevention:

- Keep distribution aligned with the current single-binary, local-first, MCP-first story.
- Treat release automation, package-manager support, docs, and install verification as the milestone core.

### 6. Packaging breadth outruns adoption proof

Risk:

- The team spends time on many package managers before proving that archives plus one or two primary channels are enough.

Prevention:

- Prioritize GitHub Releases, Homebrew, and Scoop first.
- Defer Linux native packages, signing, SBOMs, and extra channels until the base path is stable.

### 7. Functional coverage that misses degraded and stale states

Risk:

- The test plan validates only healthy repos and ignores the actual operational edge cases that matter for trust.

Prevention:

- Include healthy, stale, degraded, and recovery scenarios.
- Reuse the current doctor, refresh, watch, and pack flows rather than mocking the failure paths away.

### 8. Distribution docs that drift from real commands

Risk:

- Install and usage docs start advertising flows that differ from what the shipped CLI actually does.

Prevention:

- Derive docs from the real command surface and verify them with runnable examples.
- Keep `doctor` and `snippet` part of the install verification story.

## Recommended Guardrails

- Freeze a small benchmark corpus early and version it.
- Publish the measurement rules before the first milestone claim.
- Use current shipped surfaces as the only supported treatment path.
- Keep result capture local and deterministic.
- Treat distribution as a reproducible release pipeline, not a marketing-only document.
