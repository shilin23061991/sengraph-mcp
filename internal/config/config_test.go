package config

import (
	"strings"
	"testing"
)

func TestLoadReadsRequiredEnv(t *testing.T) {
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
	for _, k := range []string{"ZEP_API_KEY", "ZEP_USER_ID", "SENTGRAPH_PROJECT_ID"} {
		t.Setenv(k, "")
	}
	if err := Load().Validate(); err == nil {
		t.Fatal("missing required env must fail validation")
	}
}

func TestToggleDefaults(t *testing.T) {
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
