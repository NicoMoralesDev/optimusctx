# Phase 28 Context

Goal: close the gap between successful MCP registration and actual MCP value.

Observed trigger:

- Supported-client registration worked, but the product still implied that `optimusctx run` was a normal manual next step.
- The docs did not yet explain what the current `optimusctx.*` tool surface offers or how operators verify that a host is actually using it.

Key decisions:

- Make onboarding and status explicit that registered hosts should launch `optimusctx run` automatically.
- Reserve manual `run` for direct STDIO checks and debugging.
- Add one dedicated MCP guide that explains tool families, usage order, and verification of real host-side tool usage.
