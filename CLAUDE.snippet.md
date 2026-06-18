## Memory (sentgraph MCP)

You have long-term memory through the `sentgraph` MCP server, backed by Zep Cloud.

- At the start of work and when the topic changes, call `memory_context` and use the returned context.
- If you need a specific past fact, call `memory_search` with a short focused query.
- When you learn a durable preference, decision, project fact, or important lesson, save it with `memory_add` (or the `remember` skill). Do not wait until the end of the session.
- The user controls history: use `session-history` to inspect stored turns and `forget` to delete memory items.
- Routine conversation capture is handled automatically by hooks. Do not duplicate routine turns; save only facts that should last.
- Never store secrets, tokens, API keys, credentials, or private keys in memory.
