<!-- optimusctx:guidance-begin -->
## OptimusCtx MCP guidance

- When `optimusctx.*` tools are available, prefer them over broad repository exploration.
- Start with `optimusctx.symbol_lookup` or `optimusctx.structure_lookup` when you know the symbol, file, or path you need.
- Use `optimusctx.repository_map` only to orient yourself when the exact target is still unclear.
- After narrowing the target, use `optimusctx.targeted_context` to fetch bounded context instead of large reads.
- Use `optimusctx.layered_context_l0` or `optimusctx.layered_context_l1` for high-level exploration, not as the default when exact lookup is possible.
- Use `optimusctx.pack` only after you already know which files or artifacts need to be bundled together.
- If the runtime looks stale or a call fails unexpectedly, check `optimusctx.health` first and call `optimusctx.refresh` only when needed.
<!-- optimusctx:guidance-end -->
