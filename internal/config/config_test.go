package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveProjectIDEnvWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, configFileName), `project_id = "from-file"`)
	t.Setenv("SENTGRAPH_PROJECT_ID", "from-env")

	if got := resolveProjectID(dir); got != "from-env" {
		t.Fatalf("env should win, got %q", got)
	}
}

func TestResolveProjectIDFromFile(t *testing.T) {
	t.Setenv("SENTGRAPH_PROJECT_ID", "")
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, configFileName), "# project config\nproject_id = \"alpha\"  # inline\n")

	if got := resolveProjectID(dir); got != "alpha" {
		t.Fatalf("file value expected, got %q", got)
	}
}

func TestResolveProjectIDFromAncestor(t *testing.T) {
	t.Setenv("SENTGRAPH_PROJECT_ID", "")
	root := t.TempDir()
	writeFile(t, filepath.Join(root, configFileName), `project_id = "mono"`)
	child := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resolveProjectID(child); got != "mono" {
		t.Fatalf("ancestor file value expected, got %q", got)
	}
}

func TestResolveProjectIDFallbackDirName(t *testing.T) {
	t.Setenv("SENTGRAPH_PROJECT_ID", "")
	parent := t.TempDir()
	dir := filepath.Join(parent, "my-repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resolveProjectID(dir); got != "my-repo" {
		t.Fatalf("fallback dir base expected, got %q", got)
	}
}

func TestToggleDefaults(t *testing.T) {
	for _, k := range []string{"SENTGRAPH_INJECT_EVERY_PROMPT", "SENTGRAPH_PROJECT_AUTOCAPTURE", "SENTGRAPH_CAPTURE_TOOLS", "SENTGRAPH_CONTEXT_TOKEN_BUDGET", "SENTGRAPH_PROJECT_ID"} {
		t.Setenv(k, "")
	}
	c := Load(t.TempDir())
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
	t.Setenv("SENTGRAPH_PROJECT_ID", "")
	t.Setenv("SENTGRAPH_CONTEXT_TOKEN_BUDGET", "not-a-number")
	t.Setenv("SENTGRAPH_INJECT_EVERY_PROMPT", "maybe")

	c := Load(t.TempDir())
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
		{name: "missing project", cfg: Config{ZepAPIKey: "key", UserID: "user", ContextTokenBudget: 2000}, wantErr: "project id could not be resolved"},
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
