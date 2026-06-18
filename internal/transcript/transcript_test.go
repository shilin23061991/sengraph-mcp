package transcript

import (
	"strings"
	"testing"
)

func TestParseExtractsRolesAndText(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"type":"user","message":{"role":"user","content":"hello there"}}`,
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hi, how can I help?"}]}}`,
		`{"type":"system","message":{"role":"system","content":"ignored"}}`,
		`not-json-noise`,
		`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"fix the bug"}]}}`,
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"done"},{"type":"tool_use","id":"x"}]}}`,
	}, "\n")

	entries, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 text entries, got %d: %+v", len(entries), entries)
	}
	if got := LastByRole(entries, "assistant"); got != "done" {
		t.Fatalf("last assistant = %q, want %q", got, "done")
	}
	if got := LastByRole(entries, "user"); got != "fix the bug" {
		t.Fatalf("last user = %q, want %q", got, "fix the bug")
	}
}

func TestLastByRoleEmpty(t *testing.T) {
	if got := LastByRole(nil, "assistant"); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}
