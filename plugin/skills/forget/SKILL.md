# Forget

Use when the user asks to delete, remove, or forget a stored memory.

## Workflow

1. Find the memory item with `memory_search`.
2. Confirm the exact item to delete if there is any ambiguity.
3. Call `memory_forget` with the returned `kind` (`edge`, `node`, or `episode`) and `uuid`.
4. Report only what was deleted.

Deletion is destructive. Do not delete broad memory areas without explicit user confirmation.
