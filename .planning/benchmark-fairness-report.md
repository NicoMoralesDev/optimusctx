# Benchmark Fairness and Product Diagnosis

Date: 2026-03-16

## Executive summary

Phase 14 changes the benchmark answer materially.

The active benchmark corpus is now the committed `go-benchmark-discovery-v1` and `go-benchmark-refresh-v1` suites under `optimusctx/benchmark-suite@v2`. Those suites now:

- count only declared agent-input projections
- keep raw CLI and MCP payloads as provenance
- require comparable final-artifact validation for lane success

Pre-Phase-14 attribution-first evidence is superseded. The current benchmark answer comes from the repaired v2 counted-input contract and the reruns documented here.

Under that corrected boundary, the rerun result is no longer negative overall.

What the corrected counted-input reruns show:

- `go-benchmark-discovery-v1`
  - Discovery lane: baseline `39` vs OptimusCtx `10`, delta `29`
  - Context assembly lane: baseline `126` vs OptimusCtx `35`, delta `91`
- `go-benchmark-refresh-v1`
  - Refresh-after-change lane: baseline `33` vs OptimusCtx `3`, delta `30`
  - Task-completion lane: baseline `23` vs OptimusCtx `23`, delta `0`

Conclusion:

- The repaired methodology now shows OptimusCtx winning three counted lanes and tying one in the committed corpus.
- The earlier strongly negative result was mostly a boundary problem, not a stable statement about counted agent-facing benchmark cost.
- Raw system provenance is still large in several places and remains useful product-diagnosis evidence, but it is no longer conflated with counted agent input.

## What changed in Phase 14

The key correction is not "OptimusCtx got faster." The correction is "the benchmark now counts the right thing."

Before this repair:

- treatment totals could absorb large raw system outputs
- final-artifact equivalence was not a real lane gate
- rerun evidence could look negative because the benchmark counted transport and operational payloads that were not explicitly declared as agent input

After this repair:

- counted totals come only from suite-declared agent-input projections
- raw CLI stdout and MCP payloads stay exported as provenance only
- lane success requires both the stop condition and a passing final-artifact contract

That means the benchmark now answers a narrower and more defensible question:

`How much declared agent-facing input did each workflow consume while still producing the committed comparable artifact?`

It does not answer:

- provider billing truth
- end-user answer quality in the general case
- whether raw internal payloads are acceptably small

## Current committed rerun evidence

## 1. Discovery now favors OptimusCtx on counted input

`go-benchmark-discovery-v1` now uses these counted inputs:

- baseline discovery: internal tree listing plus exact symbol-match paths
- treatment discovery: one declared repository-map hint plus one declared symbol-lookup path projection
- baseline context assembly: bounded reads from the rollout handler and loader
- treatment context assembly: bounded targeted-context text

The counted rerun results are:

- Discovery lane: baseline `39`, OptimusCtx `10`
- Context assembly lane: baseline `126`, OptimusCtx `35`

Interpretation:

- once the benchmark counts only the declared discovery hint plus exact lookup projection, the treatment discovery lane is materially cheaper
- bounded targeted context is still cheaper than the baseline’s pair of manual file reads

## 2. Refresh readiness now favors OptimusCtx on counted input

`go-benchmark-refresh-v1` now uses these counted inputs:

- baseline readiness: tracked-file hints plus the mutated-note grep result
- treatment readiness: declared `health` projections for freshness and generation
- baseline task completion: bounded read of `docs/notes.txt`
- treatment task completion: bounded targeted-context text

The counted rerun results are:

- Refresh-after-change lane: baseline `33`, OptimusCtx `3`
- Task-completion lane: baseline `23`, OptimusCtx `23`

Interpretation:

- once `refresh` stdout and the full `health` object stop being counted automatically, the treatment readiness lane is much cheaper on the counted boundary
- bounded task completion is currently a tie

## 3. Export, verify, and report now pass on both committed suites

The shipped CLI path succeeds for both suites:

- `eval benchmark export`
- `eval benchmark verify`
- `eval benchmark report`

That matters because the milestone evidence is no longer coming from helper-only or synthetic suite ids. The current benchmark diagnosis is grounded in the same committed suite files that later reviewers will rerun.

## Counted input versus provenance

The positive counted result does not mean the raw product payloads are small.

The exported evidence still shows large provenance records:

- discovery provenance
  - `repository_map`: about `1047-1048` estimated tokens per attempt
  - `symbol_lookup`: about `163-164` estimated tokens per attempt
  - `targeted_context`: `35` estimated tokens per attempt
- refresh provenance
  - `refresh` stdout: `53` estimated tokens per attempt
  - `health`: `367` estimated tokens per attempt
  - `targeted_context`: `23` estimated tokens per attempt

This is now the correct split:

- counted cost: what the suite explicitly says the agent consumed
- provenance: what the runtime actually emitted and what product work might still want to shrink

That distinction is the core fairness repair.

## Why the new result is believable

The rerun outcome now matches the contract the benchmark is supposed to enforce.

Why the current result makes sense:

- discovery no longer charges the full `repository_map` payload as if the agent had consumed the entire structure
- refresh readiness no longer charges the full operational payload by default
- the benchmark still refuses bogus wins because each lane must satisfy a comparable final artifact
- bounded context retrieval remains visible and still ties or wins when compared directly to the baseline’s bounded reads

This is a more coherent methodology than the attribution-first version because the counted totals now line up with what the suites explicitly declare as agent-facing inputs.

## What this does not prove

The repaired benchmark is better, but it is still bounded.

It does not prove:

- provider-billed token truth
- broad statistical significance
- general answer quality beyond the committed final-artifact contracts
- that large provenance payloads are acceptable for real users
- that every future task will show the same direction

Treat the current result as milestone evidence for the committed corpus, not a universal product claim.
Treat the older attribution-first result as historical context only, not as the active milestone answer.

## Product diagnosis that remains valid

The benchmark-framework repair and the product backlog are separate.

These product observations still matter:

- `repository_map` remains a large provenance payload
- `health` remains a large provenance payload
- `refresh` stdout is still operationally verbose compared with the bounded signal the suite actually counts

Those are good candidates for future product optimization if the goal is to reduce raw payload size or improve ergonomics.

But that work is outside Phase 14.

Phase 14 did not redesign:

- `repository_map`
- `health`
- repository payload schemas

It only repaired how the benchmark corpus defines counted input and comparable completion.

## Recommended reading order for reviewers

If a reviewer wants the truthful milestone story, read the artifacts in this order:

1. `eval benchmark report` output for the counted-lane summary
2. exported evidence bundles for provenance inspection
3. the committed suite JSON files for the exact counted-input and final-artifact contracts

That order prevents the old mistake of treating raw provenance volume as the counted benchmark answer.

## Final judgment

The current Phase 14 benchmark evidence says:

- the methodology repair is successful
- the counted benchmark result is now positive or neutral across the committed corpus
- the raw product payloads can still be large and remain legitimate future optimization work

That is a much stronger and more defensible position than the pre-repair benchmark, and it keeps benchmark-framework repair clearly separated from later product redesign work.
