// Package transcript parses Claude Code transcript JSONL files. The Stop and
// SessionEnd hooks receive a transcript path; we extract the latest turns to
// persist them to Zep. The parser is deliberately tolerant of unknown or
// malformed lines so a single bad record never drops a whole turn.
package transcript

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// Entry is a single user or assistant turn with its plain-text content.
type Entry struct {
	Role string
	Text string
}

type record struct {
	Type    string   `json:"type"`
	Role    string   `json:"role"`
	Message *message `json:"message"`
}

type message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type block struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

const maxLine = 16 * 1024 * 1024

// Parse reads JSONL and returns the ordered user/assistant entries that carry
// text. Non-JSON lines and roles other than user/assistant are skipped.
func Parse(r io.Reader) ([]Entry, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxLine)

	var out []Entry
	for sc.Scan() {
		raw := strings.TrimSpace(sc.Text())
		if raw == "" {
			continue
		}
		var rec record
		if err := json.Unmarshal([]byte(raw), &rec); err != nil {
			continue
		}

		// Priority: Message.Role > rec.Role > rec.Type.
		var role string
		var content json.RawMessage
		if rec.Message != nil {
			if rec.Message.Role != "" {
				role = rec.Message.Role
			} else {
				role = rec.Role
			}
			content = rec.Message.Content
		} else {
			role = rec.Role
		}
		if role == "" {
			role = rec.Type
		}
		if role != "user" && role != "assistant" {
			continue
		}

		text := extractText(content)
		if text == "" {
			continue
		}
		out = append(out, Entry{Role: role, Text: text})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// extractText handles content that is either a plain string or an array of
// typed blocks (only "text" blocks contribute).
func extractText(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var blocks []block
	if err := json.Unmarshal(content, &blocks); err == nil {
		var b strings.Builder
		for _, bl := range blocks {
			if bl.Type == "text" && bl.Text != "" {
				if b.Len() > 0 {
					b.WriteByte('\n')
				}
				b.WriteString(bl.Text)
			}
		}
		return strings.TrimSpace(b.String())
	}
	return ""
}

// LastByRole returns the text of the last entry with the given role, or "".
func LastByRole(entries []Entry, role string) string {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Role == role {
			return entries[i].Text
		}
	}
	return ""
}
