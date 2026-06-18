package config

import (
	"os"
	"path/filepath"
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

func TestProjectGraphID(t *testing.T) {
	if got := (Config{ProjectID: "alpha"}).ProjectGraphID(); got != "proj:alpha" {
		t.Fatalf("graph id = %q", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
