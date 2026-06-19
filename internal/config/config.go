// Package config resolves runtime configuration from required environment
// variables, optionally seeded from a per-project .env.local file so each
// project can carry its own keys without a shared global environment.
//
// Memory scoping: a single Zep user (the developer) holds personal,
// cross-project context; each project gets its own standalone graph. A
// "project" may span several repositories by sharing the same
// SENTGRAPH_PROJECT_ID and therefore one project graph.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

const envFileName = ".env.local"

// Config holds resolved settings. Build it with Load.
type Config struct {
	ZepAPIKey string
	UserID    string
	ProjectID string

	// EnvFilePresent reports whether a .env.local was found for this project.
	// serve and doctor require it (RequireEnvFile) so a global (user-scope)
	// install does not silently run in projects that are not set up.
	EnvFilePresent bool

	// envFileErr is non-nil when a .env.local was found but could not be loaded
	// (syntax/permission error). RequireEnvFile surfaces it so the user sees the
	// real cause instead of a misleading "key is required".
	envFileErr error

	// Hook frequency / behavior toggles ("read more, write more").
	// TODO: Wire these into hooks and context assembly once runtime tuning is
	// exposed; defaults currently match the intended first release behavior.
	InjectEveryPrompt  bool
	ProjectAutocapture bool
	CaptureTools       bool
	ContextTokenBudget int
}

// Load resolves configuration from the environment. A per-project .env.local
// (searched upward from CLAUDE_PROJECT_DIR or the working directory) is loaded
// first and takes precedence, so each project supplies its own keys without a
// shared global environment. The three identity values are still required;
// Validate rejects any that are empty.
func Load() Config {
	found, envErr := loadEnvFile()
	return Config{
		ZepAPIKey:          os.Getenv("ZEP_API_KEY"),
		UserID:             os.Getenv("ZEP_USER_ID"),
		ProjectID:          os.Getenv("SENTGRAPH_PROJECT_ID"),
		InjectEveryPrompt:  boolEnv("SENTGRAPH_INJECT_EVERY_PROMPT", true),
		ProjectAutocapture: boolEnv("SENTGRAPH_PROJECT_AUTOCAPTURE", true),
		CaptureTools:       boolEnv("SENTGRAPH_CAPTURE_TOOLS", false),
		ContextTokenBudget: intEnv("SENTGRAPH_CONTEXT_TOKEN_BUDGET", 2000),
		EnvFilePresent:     found,
		envFileErr:         envErr,
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

// RequireEnvFile guards against global (user-scope) or accidental installs:
// without a .env.local in the project, serve and doctor refuse to run.
func (c Config) RequireEnvFile() error {
	if !c.EnvFilePresent {
		return errors.New(".env.local not found in project: sentgraph-mcp is project-scoped -- create .env.local in the project and install the plugin with --scope project")
	}
	if c.envFileErr != nil {
		return fmt.Errorf(".env.local found but could not be loaded: %w", c.envFileErr)
	}
	return nil
}

// loadEnvFile seeds the process environment from the nearest .env.local so each
// project can carry its own keys. It searches upward from CLAUDE_PROJECT_DIR
// (set by Claude Code for project-scoped servers) or the working directory and
// loads the file with godotenv (non-override): existing environment variables
// win, the file only fills in the ones that are unset. It returns whether a
// file was found and any load error (a found-but-unparsable file). Missing
// files are not an error.
func loadEnvFile() (bool, error) {
	base := os.Getenv("CLAUDE_PROJECT_DIR")
	if base == "" {
		wd, err := os.Getwd()
		if err != nil {
			return false, nil
		}
		base = wd
	}
	path, ok := findUp(base, envFileName)
	if !ok {
		return false, nil
	}
	if err := godotenv.Load(path); err != nil {
		return true, fmt.Errorf("load %s: %w", path, err)
	}
	return true, nil
}

// findUp returns the path to name in the nearest ancestor directory of start.
func findUp(start, name string) (string, bool) {
	dir := filepath.Clean(start)
	for {
		if p := filepath.Join(dir, name); fileExists(p) {
			return p, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
