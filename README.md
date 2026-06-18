# sengraph-mcp

Go MCP server for long-term coding-agent memory backed by Zep Cloud.

Sentgraph keeps the local layer thin:

- Zep Cloud owns graph construction, deduplication, embeddings, and retrieval.
- Sentgraph exposes six core MCP tools for agents.
- Native Go hooks read/write memory frequently without a local daemon.
- Project memory is scoped by `project_id`, so one project can span several repositories.

## Status

Implemented:

- Go CLI skeleton: `sentgraph serve`, `sentgraph hook <event>`, `sentgraph doctor`.
- Config resolution from env and `.sentgraph.toml`.
- Secret redaction before cloud writes.
- Claude transcript parsing for Stop/SessionEnd hooks.
- Core memory service with Zep limits: 30 messages/call, 4096 chars/message, 10000 chars/graph payload chunk.
- Zep Cloud adapter using `github.com/getzep/zep-go/v3`.
- MCP server using the official `github.com/modelcontextprotocol/go-sdk`.
- Claude hook config and skill documentation.

## Install

```bash
go install github.com/shilin23061991/sengraph-mcp/cmd/sentgraph@latest
```

For local development:

```bash
go build ./cmd/sentgraph
```

## Configuration

Required:

```bash
export ZEP_API_KEY="..."
export ZEP_USER_ID="mrshikadancer"
```

Project scope can be shared across many repos by adding `.sentgraph.toml`:

```toml
project_id = "sentoke"
```

Optional:

```bash
export SENTGRAPH_PROJECT_ID="sentoke"
export SENTGRAPH_INJECT_EVERY_PROMPT=true
export SENTGRAPH_PROJECT_AUTOCAPTURE=true
export SENTGRAPH_CAPTURE_TOOLS=false
export SENTGRAPH_CONTEXT_TOKEN_BUDGET=2000
```

## Commands

```bash
sentgraph doctor
sentgraph doctor --online
sentgraph serve
sentgraph serve --http :8080
sentgraph hook SessionStart
```

## MCP tools

- `memory_context` -- assembled user + project context.
- `memory_search` -- search user or project graph memory.
- `memory_history` -- get recent thread messages.
- `memory_add_messages` -- persist conversation turns.
- `memory_add` -- persist project/user facts or data.
- `memory_forget` -- delete an edge, node, or episode by UUID.

## Hooks

The Claude plugin hook config is in `plugin/hooks/hooks.json`.

Default events:

- `SessionStart` -- read context and inject it.
- `UserPromptSubmit` -- write the user prompt and inject fresh context.
- `PreCompact` -- re-inject context before compaction.
- `Stop` -- persist the latest assistant turn from the transcript.
- `SessionEnd` -- final persist pass.

## Design

See `docs/implementation-plan.md` for the full plan and `zep-memory.md` for the architecture notes.
