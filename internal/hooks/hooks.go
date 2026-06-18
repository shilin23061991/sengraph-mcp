package hooks

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/shilin23061991/sengraph-mcp/internal/memory"
	"github.com/shilin23061991/sengraph-mcp/internal/transcript"
)

type Service interface {
	EnsureIdentity(ctx context.Context, threadID string) error
	AddTurn(ctx context.Context, threadID string, messages []memory.Message, returnContext bool) (string, error)
	GetContext(ctx context.Context, opts memory.ContextOptions) (string, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

type Payload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	Prompt         string `json:"prompt"`
	HookEventName  string `json:"hook_event_name"`
}

type Response struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func (h *Handler) Handle(ctx context.Context, event string, r io.Reader, w io.Writer) error {
	var p Payload
	if err := json.NewDecoder(r).Decode(&p); err != nil && err != io.EOF {
		return err
	}
	threadID := firstNonEmpty(p.SessionID, "sentgraph-default")
	if err := h.service.EnsureIdentity(ctx, threadID); err != nil {
		return err
	}

	switch event {
	case "SessionStart", "PreCompact":
		return h.injectContext(ctx, w, event, threadID, "")
	case "UserPromptSubmit":
		prompt := strings.TrimSpace(p.Prompt)
		var contextBlock string
		if prompt != "" {
			var err error
			contextBlock, err = h.service.AddTurn(ctx, threadID, []memory.Message{{Role: "user", Content: prompt}}, true)
			if err != nil {
				return err
			}
		}
		if contextBlock == "" {
			return h.injectContext(ctx, w, event, threadID, prompt)
		}
		return writeContext(w, event, contextBlock)
	case "Stop", "SessionEnd":
		return h.persistLatestAssistant(ctx, p.TranscriptPath, threadID)
	default:
		return nil
	}
}

func (h *Handler) injectContext(ctx context.Context, w io.Writer, event, threadID, query string) error {
	contextBlock, err := h.service.GetContext(ctx, memory.ContextOptions{ThreadID: threadID, Query: query, Limit: 5})
	if err != nil {
		return err
	}
	if contextBlock == "" {
		return nil
	}
	return writeContext(w, event, contextBlock)
}

func (h *Handler) persistLatestAssistant(ctx context.Context, transcriptPath, threadID string) error {
	if transcriptPath == "" {
		return nil
	}
	f, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer f.Close()

	entries, err := transcript.Parse(f)
	if err != nil {
		return err
	}
	text := transcript.LastByRole(entries, "assistant")
	if strings.TrimSpace(text) == "" {
		return nil
	}
	_, err = h.service.AddTurn(ctx, threadID, []memory.Message{{Role: "assistant", Content: text}}, false)
	return err
}

func writeContext(w io.Writer, event, contextBlock string) error {
	return json.NewEncoder(w).Encode(Response{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:     event,
			AdditionalContext: contextBlock,
		},
	})
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
