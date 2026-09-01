package onboarding

import "testing"

func TestNeedsSecrets(t *testing.T) {
	if !NeedsSecrets("api_key=123") { t.Fatal("api_key") }
	if !NeedsSecrets("sk-proj-xxx") { t.Fatal("sk-") }
	if NeedsSecrets("Go and TypeScript") { t.Fatal("falso positivo") }
}

func TestResult_IsComplete(t *testing.T) {
	r := Result{VaultPath: "C:/vault"}
	if !r.IsComplete() { t.Fatal("debe ser completo") }
	if (Result{}).IsComplete() { t.Fatal("vacio no completo") }
}
