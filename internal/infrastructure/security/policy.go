package security

import "strings"

// Policy implements ports.PermissionPolicy
type Policy struct {
	// Default deny for write/execute/network unless explicitly allowed
	Allowed map[string]map[string]bool // agentID -> tool -> allowed
}

func NewPolicy() *Policy {
	return &Policy{Allowed: make(map[string]map[string]bool)}
}

func (p *Policy) Allow(agentID, tool string) {
	if p.Allowed[agentID] == nil {
		p.Allowed[agentID] = make(map[string]bool)
	}
	p.Allowed[agentID][tool] = true
}

func (p *Policy) Can(agentID, tool, permission string) bool {
	// read is always allowed
	if permission == "read" {
		return true
	}
	// check explicit allow
	if m, ok := p.Allowed[agentID]; ok {
		if m[tool] {
			return true
		}
		// wildcard
		if m["*"] {
			return true
		}
	}
	// default: planner can only read, coder can read/write, researcher read
	switch agentID {
	case "planner":
		return permission == "read"
	case "coder", "developer":
		return permission == "read" || permission == "write"
	case "researcher":
		return permission == "read" || permission == "network"
	default:
		return permission == "read"
	}
}

// DetectSecret is a helper for memory and logs
func DetectSecret(s string) bool {
	lower := strings.ToLower(s)
	indicators := []string{"api_key", "apikey", "sk-", "ghp_", "akia", "private_key", "password="}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	return false
}
