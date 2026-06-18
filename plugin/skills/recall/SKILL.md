# Recall

Use when the user asks you to remember prior context, previous decisions, project facts, preferences, or "what did we decide".

## Workflow

1. Call `memory_context` with the current `thread_id` and a concise `query` when available.
2. If the answer needs a specific fact, call `memory_search` with `target: "project"` first, then `target: "user"` if needed.
3. Answer from the returned context. If memory is empty or uncertain, say so clearly.

Do not invent missing memories.
