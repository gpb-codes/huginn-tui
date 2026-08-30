package memory

// Memory and Knowledge are domain concepts.
// Agent Vault owns persistence; Huginn only depends on ports (interfaces).
type Entry struct {
	Text       string
	Project    string
	Importance string
	Type       string
}
