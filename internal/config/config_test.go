package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolate points config resolution at an empty directory so a stray .env.local
// (in the repo or a developer's home tree) cannot leak into env-based tests.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(t.TempDir())
}

func TestLoadReadsRequiredEnv(t *testing.T) {
	isolate(t)
	t.Setenv("ZEP_API_KEY", "key-123")
	t.Setenv("ZEP_USER_ID", "dev-7")
	t.Setenv("SENTGRAPH_PROJECT_ID", "sentoke")

	c := Load()
	if c.ZepAPIKey != "key-123" || c.UserID != "dev-7" || c.ProjectID != "sentoke" {
		t.Fatalf("env not read into config: %+v", c)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("fully configured env should validate: %v", err)
	}
}

func TestLoadMissingRequiredFailsValidate(t *testing.T) {
	isolate(t)
	for _, k := range []string{"ZEP_API_KEY", "ZEP_USER_ID", "SENTGRAPH_PROJECT_ID"} {
		t.Setenv(k, "")
	}
	if err := Load().Validate(); err == nil {
		t.Fatal("missing required env must fail validation")
	}
}

func TestLoadEnvFileFillsUnsetAndKeepsExisting(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env.local"),
		"# project keys\nZEP_API_KEY=\"local-key\"\nexport ZEP_USER_ID=local-user\nSENTGRAPH_PROJECT_ID=local-proj\n")
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	// An existing env var must win (non-override, standard godotenv behavior).
	t.Setenv("ZEP_API_KEY", "env-key")
	// Unset vars get filled from .env.local.
	unset(t, "ZEP_USER_ID")
	unset(t, "SENTGRAPH_PROJECT_ID")

	c := Load()
	if c.ZepAPIKey != "env-key" {
		t.Fatalf("ZEP_API_KEY = %q, want env-key (existing env must win)", c.ZepAPIKey)
	}
	if c.UserID != "local-user" {
		t.Fatalf("ZEP_USER_ID = %q, want local-user (filled from .env.local)", c.UserID)
	}
	if c.ProjectID != "local-proj" {
		t.Fatalf("SENTGRAPH_PROJECT_ID = %q, want local-proj (filled from .env.local)", c.ProjectID)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("should validate: %v", err)
	}
}

func TestLoadEnvFileFoundUpward(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env.local"), "SENTGRAPH_PROJECT_ID=mono\n")
	child := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", child)
	unset(t, "SENTGRAPH_PROJECT_ID")

	if got := Load().ProjectID; got != "mono" {
		t.Fatalf("ProjectID = %q, want mono from ancestor .env.local", got)
	}
}

func TestToggleDefaults(t *testing.T) {
	isolate(t)
	for _, k := range []string{"SENTGRAPH_INJECT_EVERY_PROMPT", "SENTGRAPH_PROJECT_AUTOCAPTURE", "SENTGRAPH_CAPTURE_TOOLS", "SENTGRAPH_CONTEXT_TOKEN_BUDGET", "SENTGRAPH_PROJECT_ID"} {
		t.Setenv(k, "")
	}
	c := Load()
	if !c.InjectEveryPrompt || !c.ProjectAutocapture {
		t.Fatalf("read/write toggles should default on: %+v", c)
	}
	if c.CaptureTools {
		t.Fatalf("tool capture should default off")
	}
	if c.ContextTokenBudget != 2000 {
		t.Fatalf("token budget default = %d, want 2000", c.ContextTokenBudget)
	}
}

func TestInvalidEnvFallsBackToDefaults(t *testing.T) {
	isolate(t)
	t.Setenv("SENTGRAPH_PROJECT_ID", "")
	t.Setenv("SENTGRAPH_CONTEXT_TOKEN_BUDGET", "not-a-number")
	t.Setenv("SENTGRAPH_INJECT_EVERY_PROMPT", "maybe")

	c := Load()
	if c.ContextTokenBudget != 2000 {
		t.Fatalf("token budget = %d, want default 2000", c.ContextTokenBudget)
	}
	if !c.InjectEveryPrompt {
		t.Fatalf("invalid bool should fall back to default true")
	}
}

func TestProjectGraphID(t *testing.T) {
	if got := (Config{ProjectID: "alpha"}).ProjectGraphID(); got != "proj:alpha" {
		t.Fatalf("graph id = %q", got)
	}
}

func TestValidate(t *testing.T) {
	valid := Config{
		ZepAPIKey:          "key",
		UserID:             "user",
		ProjectID:          "project",
		ContextTokenBudget: 2000,
	}
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{name: "valid", cfg: valid},
		{name: "missing api key", cfg: Config{UserID: "user", ProjectID: "project", ContextTokenBudget: 2000}, wantErr: "ZEP_API_KEY is required"},
		{name: "missing user", cfg: Config{ZepAPIKey: "key", ProjectID: "project", ContextTokenBudget: 2000}, wantErr: "ZEP_USER_ID is required"},
		{name: "missing project", cfg: Config{ZepAPIKey: "key", UserID: "user", ContextTokenBudget: 2000}, wantErr: "SENTGRAPH_PROJECT_ID is required"},
		{name: "invalid budget", cfg: Config{ZepAPIKey: "key", UserID: "user", ProjectID: "project"}, wantErr: "SENTGRAPH_CONTEXT_TOKEN_BUDGET must be greater than zero"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestEnvFilePresentReported(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ".env.local"), "SENTGRAPH_PROJECT_ID=p\n")
		t.Setenv("CLAUDE_PROJECT_DIR", dir)
		unset(t, "SENTGRAPH_PROJECT_ID")
		if !Load().EnvFilePresent {
			t.Fatal("EnvFilePresent should be true when .env.local exists")
		}
	})
	t.Run("missing", func(t *testing.T) {
		isolate(t)
		if Load().EnvFilePresent {
			t.Fatal("EnvFilePresent should be false without .env.local")
		}
	})
}

func TestRequireEnvFile(t *testing.T) {
	if err := (Config{EnvFilePresent: true}).RequireEnvFile(); err != nil {
		t.Fatalf("present should pass: %v", err)
	}
	if err := (Config{EnvFilePresent: false}).RequireEnvFile(); err == nil {
		t.Fatal("absent .env.local should error")
	}
	// A found-but-unparsable .env.local must surface the load error, not pass.
	err := (Config{EnvFilePresent: true, envFileErr: errors.New("boom")}).RequireEnvFile()
	if err == nil || !strings.Contains(err.Error(), "could not be loaded") {
		t.Fatalf("load error should be surfaced, got %v", err)
	}
}

func TestLoadEnvFileSurfacesParseError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env.local"), "not a valid env line\n")
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	cfg := Load()
	if !cfg.EnvFilePresent {
		t.Fatal("file should be reported present")
	}
	if err := cfg.RequireEnvFile(); err == nil {
		t.Fatal("malformed .env.local must make RequireEnvFile error")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// unset removes an env var for the duration of the test (restoring it after) so
// non-override loading can fill it from .env.local. t.Setenv cannot express an
// unset variable; it only sets values.
func unset(t *testing.T, key string) {
	t.Helper()
	orig, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, orig)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}
