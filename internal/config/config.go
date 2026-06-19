// Package config resolves runtime configuration from required environment
// variables. The three identity values (API key, user, project) must be set
// explicitly; there are no silent fallbacks.
//
// Memory scoping: a single Zep user (the developer) holds personal,
// cross-project context; each project gets its own standalone graph. A
// "project" may span several repositories by sharing the same
// SENTGRAPH_PROJECT_ID and therefore one project graph.
package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds resolved settings. Build it with Load.
type Config struct {
	ZepAPIKey string
	UserID    string
	ProjectID string

	// Hook frequency / behavior toggles ("read more, write more").
	// TODO: Wire these into hooks and context assembly once runtime tuning is
	// exposed; defaults currently match the intended first release behavior.
	InjectEveryPrompt  bool
	ProjectAutocapture bool
	CaptureTools       bool
	ContextTokenBudget int
}

// Load resolves configuration strictly from environment variables. The three
// identity values have no fallbacks; Validate rejects any that are empty.
func Load() Config {
	return Config{
		ZepAPIKey:          os.Getenv("ZEP_API_KEY"),
		UserID:             os.Getenv("ZEP_USER_ID"),
		ProjectID:          os.Getenv("SENTGRAPH_PROJECT_ID"),
		InjectEveryPrompt:  boolEnv("SENTGRAPH_INJECT_EVERY_PROMPT", true),
		ProjectAutocapture: boolEnv("SENTGRAPH_PROJECT_AUTOCAPTURE", true),
		CaptureTools:       boolEnv("SENTGRAPH_CAPTURE_TOOLS", false),
		ContextTokenBudget: intEnv("SENTGRAPH_CONTEXT_TOKEN_BUDGET", 2000),
	}
}

// ProjectGraphID is the Zep graph_id for this project's standalone graph.
func (c Config) ProjectGraphID() string {
	return "proj:" + c.ProjectID
}

// Validate reports whether the config is usable for talking to Zep. All three
// identity keys are required.
func (c Config) Validate() error {
	switch {
	case c.ZepAPIKey == "":
		return errors.New("ZEP_API_KEY is required")
	case c.UserID == "":
		return errors.New("ZEP_USER_ID is required")
	case c.ProjectID == "":
		return errors.New("SENTGRAPH_PROJECT_ID is required")
	case c.ContextTokenBudget <= 0:
		return errors.New("SENTGRAPH_CONTEXT_TOKEN_BUDGET must be greater than zero")
	default:
		return nil
	}
}

func boolEnv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func intEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
