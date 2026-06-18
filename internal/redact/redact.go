// Package redact removes high-confidence secrets from text before it is sent
// to the Zep Cloud memory backend. Zep builds the knowledge graph and dedups
// on its side, but secret stripping must happen locally, before anything
// leaves the machine.
package redact

import "regexp"

const placeholder = "[REDACTED]"

// patterns are high-confidence secret shapes. Each is applied independently;
// order is irrelevant.
var patterns = []*regexp.Regexp{
	// JWT (header.payload.signature)
	regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`),
	// OpenAI / Stripe-style sk- keys
	regexp.MustCompile(`sk-[A-Za-z0-9_-]{16,}`),
	// GitHub tokens (ghp_/gho_/ghu_/ghs_/ghr_) and fine-grained PATs
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`),
	// AWS access key id
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	// Google API key
	regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),
	// Slack tokens
	regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{10,}`),
	// Bearer tokens in Authorization headers
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]{12,}`),
}

// Secrets replaces any high-confidence secret found in s with a placeholder
// and returns the cleaned string. It is safe to call on arbitrary text.
func Secrets(s string) string {
	for _, re := range patterns {
		s = re.ReplaceAllString(s, placeholder)
	}
	return s
}
