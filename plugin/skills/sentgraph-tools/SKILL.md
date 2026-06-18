# Sentgraph MCP Tools

Reference for the six core Sentgraph tools. Use this when deciding which memory tool to call.

## memory_context

Read-only. Returns assembled user memory plus optional project memory.

Input:

```json
{"thread_id":"<session id>","query":"optional focused query","limit":5}
```

Use at the start of work, after topic changes, and before answering memory-dependent questions.

## memory_search

Read-only. Searches Zep graph memory.

Input:

```json
{"query":"short focused query","target":"project","scope":"edges","limit":10}
```

Targets: `project` (default) or `user`.
Scopes: `edges`, `nodes`, `episodes`, `auto`.

Use for exact recall and before deletion.

## memory_history

Read-only. Returns recent stored messages for a thread.

Input:

```json
{"thread_id":"<session id>","limit":20}
```

Use for transparency and user inspection of current-session history.

## memory_add_messages

Write, non-destructive. Persists conversation messages to a Zep thread.

Input:

```json
{
  "thread_id": "<session id>",
  "return_context": true,
  "messages": [{"role":"user","content":"..."}]
}
```

Limits enforced locally: max 30 messages per call, max 4096 characters per message.
Routine turn capture is normally handled by hooks.

## memory_add

Write, non-destructive. Persists facts, decisions, or project/user data to Zep graph memory.

Input:

```json
{"target":"project","type":"text","data":"Decision: ...","description":"optional"}
```

Targets: `project` (default) or `user`.
Types: `text` or `json`.
Payloads over 10000 characters are chunked locally.

Use for durable facts, not routine transcript duplication.

## memory_forget

Write, destructive. Deletes a Zep memory item by UUID.

Input:

```json
{"kind":"edge","uuid":"..."}
```

Kinds: `edge`, `node`, `episode`.
Use only after the user asks to delete/forget and the exact item is known.
