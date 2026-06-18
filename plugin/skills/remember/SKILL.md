# Remember

Use when the user explicitly asks you to remember something, or when you learn a durable preference, project decision, invariant, or implementation fact that should survive future sessions.

## Workflow

1. Decide whether the memory is personal/user-level or project-level.
2. Call `memory_add` with:
   - `target: "user"` for personal preferences and cross-project rules.
   - `target: "project"` for architecture, repo conventions, decisions, and implementation facts.
3. Keep the saved text concise and factual.

Never save secrets, tokens, credentials, private keys, or raw sensitive logs.
