package redact

import (
	"strings"
	"testing"
)

func TestSecretsRedactsKnownTokens(t *testing.T) {
	// Secrets are assembled from fragments so this source file contains no
	// contiguous secret-shaped literal (keeps secret scanners meaningful).
	cases := []struct {
		name   string
		secret string
	}{
		{"openai", "sk" + "-" + strings.Repeat("aA1bB2cC", 3)},
		{"github", "gh" + "p_" + strings.Repeat("a", 36)},
		{"github_pat", "github" + "_pat_" + strings.Repeat("b", 30)},
		{"aws", "AK" + "IA" + strings.Repeat("Q", 16)},
		{"google", "AI" + "za" + strings.Repeat("c", 35)},
		{"slack", "xo" + "xb-" + strings.Repeat("9", 12) + "-abcdefghij"},
		{"jwt", "ey" + "J0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9" + "." + strings.Repeat("a", 20) + "." + strings.Repeat("b", 20)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := "before " + c.secret + " after"
			got := Secrets(in)
			if strings.Contains(got, c.secret) {
				t.Fatalf("secret %q not redacted: %q", c.secret, got)
			}
			if !strings.HasPrefix(got, "before ") || !strings.HasSuffix(got, " after") {
				t.Fatalf("surrounding text damaged: %q", got)
			}
		})
	}
}

func TestSecretsRedactsBearer(t *testing.T) {
	token := strings.Repeat("xy7Z", 5)
	in := "Authorization: Bearer " + token
	if got := Secrets(in); strings.Contains(got, token) {
		t.Fatalf("bearer token not redacted: %q", got)
	}
}

func TestSecretsRedactsMultiple(t *testing.T) {
	openAI := "sk" + "-" + strings.Repeat("aA1bB2cC", 3)
	aws := "AK" + "IA" + strings.Repeat("Q", 16)
	in := "first " + openAI + " second " + aws
	got := Secrets(in)
	for _, secret := range []string{openAI, aws} {
		if strings.Contains(got, secret) {
			t.Fatalf("secret %q not redacted: %q", secret, got)
		}
	}
}

func TestSecretsRedactsAtBoundaries(t *testing.T) {
	secret := "gh" + "p_" + strings.Repeat("a", 36)
	cases := []string{
		secret + " at start",
		"at end " + secret,
		secret,
		secret + "sk" + "-" + strings.Repeat("z9", 12),
	}
	for _, in := range cases {
		if got := Secrets(in); strings.Contains(got, secret) {
			t.Fatalf("boundary secret not redacted: %q", got)
		}
	}
}

func TestSecretsKeepsNormalText(t *testing.T) {
	in := "The user prefers tabs over spaces and targets Go 1.25."
	if got := Secrets(in); got != in {
		t.Fatalf("normal text altered: %q", got)
	}
}
