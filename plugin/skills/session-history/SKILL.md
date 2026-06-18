# Session History

Use when the user asks what has been recorded in the current conversation/session, or wants to inspect recent stored turns.

## Workflow

1. Call `memory_history` with the current `thread_id`.
2. Summarize the stored messages briefly.
3. If the user asks to remove an item, switch to the `forget` workflow.

This skill is for transparency. It should not add new memory.
