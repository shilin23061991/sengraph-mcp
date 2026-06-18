// Package config resolves runtime configuration from the environment and an
// optional per-repo .sentgraph.toml file.
//
// Memory scoping: a single Zep user (the developer) holds personal,
// cross-project context; each project gets its own standalone graph. A
// "project" may span several repositories, so multiple repos can share one
// project_id and therefore one project graph.
package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const configFileName = ".sentgraph.toml"

// Config holds resolved settings. Build it with Load.
type Config struct {
	ZepAPIKey string
	UserID    string
	ProjectID string

	// Hook frequency / behavior toggles ("read more, write more").
	InjectEveryPrompt  bool
	ProjectAutocapture bool
	CaptureTools       bool
	ContextTokenBudget int
}

// Load resolves configuration. startDir is the repo/working directory used to
// locate .sentgraph.toml (searched upward) and to derive a fallback project id.
func Load(startDir string) Config {
	return Config{
		ZepAPIKey:          os.Getenv("ZEP_API_KEY"),
		UserID:             firstNonEmpty(os.Getenv("ZEP_USER_ID"), os.Getenv("USER")),
		ProjectID:          resolveProjectID(startDir),
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

// Validate reports whether the config is usable for talking to Zep.
func (c Config) Validate() error {
	switch {
	case c.ZepAPIKey == "":
		return errors.New("ZEP_API_KEY is required")
	case c.UserID == "":
		return errors.New("ZEP_USER_ID is required (or set $USER)")
	case c.ProjectID == "":
		return errors.New("project id could not be resolved (set SENTGRAPH_PROJECT_ID or add .sentgraph.toml)")
	default:
		return nil
	}
}

func resolveProjectID(startDir string) string {
	if v := os.Getenv("SENTGRAPH_PROJECT_ID"); v != "" {
		return v
	}
	if dir, ok := findUp(startDir, configFileName); ok {
		if v := readTOMLString(filepath.Join(dir, configFileName), "project_id"); v != "" {
			return v
		}
		return filepath.Base(dir)
	}
	if startDir != "" {
		return filepath.Base(filepath.Clean(startDir))
	}
	return ""
}

func findUp(start, name string) (string, bool) {
	dir := start
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", false
		}
		dir = wd
	}
	dir = filepath.Clean(dir)
	for {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// readTOMLString reads a single top-level string key from a minimal TOML file.
// Only the flat `key = "value"` form is supported, which is all sentgraph needs.
func readTOMLString(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(k) != key {
			continue
		}
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, `"`) {
			rest := v[1:]
			if i := strings.Index(rest, `"`); i >= 0 {
				return rest[:i]
			}
			return rest
		}
		if i := strings.Index(v, "#"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		return v
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
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
