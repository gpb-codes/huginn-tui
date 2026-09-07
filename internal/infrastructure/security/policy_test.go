// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package security

import (
	"strings"
	"testing"
)

func TestPolicy_ReadAlwaysAllowed(t *testing.T) {
	p := NewPolicy()
	if !p.Can("unknown-agent", "anything", "read") {
		t.Fatal("read must always be allowed")
	}
}

func TestPolicy_DefaultDenyWrite(t *testing.T) {
	p := NewPolicy()
	if p.Can("researcher-x", "filesystem", "write") {
		t.Fatal("write must default-deny")
	}
}

func TestPolicy_ExplicitAllow(t *testing.T) {
	p := NewPolicy()
	p.Allow("opencode", "*")
	if !p.Can("opencode", "filesystem", "write") {
		t.Fatal("explicit allow must grant")
	}
	if p.Can("ollama", "filesystem", "write") {
		t.Fatal("grant must not leak to other agents")
	}
}

func TestRedact_NoSecrets(t *testing.T) {
	in := " implement login form "
	if Redact(in) != in {
		t.Fatal("benign text must pass through")
	}
}

func TestRedact_Assignments(t *testing.T) {
	out := Redact("config:\n  api_key=supersecret123\n  user=bob")
	if strings.Contains(out, "supersecret123") {
		t.Fatalf("secret leaked: %q", out)
	}
	if !strings.Contains(out, "user=bob") {
		t.Fatalf("benign line altered: %q", out)
	}
}

func TestRedact_TokenShapes(t *testing.T) {
	out := Redact("key ghp_abcdefgh1234567890 here")
	if strings.Contains(out, "ghp_abcdefgh1234567890") {
		t.Fatalf("token leaked: %q", out)
	}
}
