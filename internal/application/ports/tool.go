package ports

import "context"

// Tool represents a capability an agent can use.
type Tool struct {
	Name        string
	Description string
	Permissions []string // read, write, execute, network, git
}

// ToolExecutor executes a tool call.
type ToolExecutor interface {
	Execute(ctx context.Context, tool string, args map[string]string) (string, error)
}

// PermissionPolicy decides if an agent can use a tool.
type PermissionPolicy interface {
	Can(agentID, tool string, permission string) bool
}

// ToolResult is the outcome of a tool execution.
type ToolResult struct {
	Tool    string
	Success bool
	Output  string
	Error   string
}
