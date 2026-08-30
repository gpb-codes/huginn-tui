package config

// Settings tree — single source of truth for ⚙ Settings UI.
// Infrastructure layer: persistence of values (in-memory for now, file later).
type Section struct {
	Name  string
	Items []string
}

var Sections = []Section{
	{Name: "General", Items: []string{"Profile", "Language", "Timezone", "Startup behavior"}},
	{Name: "Appearance", Items: []string{"Theme", "Accent color", "Font size", "Chat density"}},
	{Name: "AI", Items: []string{"Providers", "Models", "Default model", "Reasoning", "Token limits"}},
	{Name: "Agents", Items: []string{"Default agent", "Agent permissions", "Agent routing", "Max agents", "Execution limits"}},
	{Name: "Memory", Items: []string{"Memory enabled", "Persistent memory", "Auto-save", "Context retrieval", "Knowledge Graph"}},
	{Name: "Tools", Items: []string{"Web", "Files", "GitHub", "Terminal", "MCP"}},
	{Name: "Vault", Items: []string{"Vault connection", "Default vault", "Sync", "Indexing", "Embeddings"}},
	{Name: "Advanced", Items: []string{"API", "Logs", "Debug", "Experimental"}},
}

var Values = map[string]string{
	"Profile": "gabriel", "Language": "Español", "Timezone": "UTC-3", "Startup behavior": "Restore session",
	"Theme": "Dark", "Accent color": "#33d9f2", "Font size": "14px", "Chat density": "Comfortable",
	"Providers": "OpenCode Zen", "Models": "5 configured", "Default model": "mimo-v2.5-free", "Reasoning": "Enabled", "Token limits": "8k",
	"Default agent": "OpenCode", "Agent permissions": "Ask", "Agent routing": "Auto", "Max agents": "4", "Execution limits": "90s",
	"Memory enabled": "On", "Persistent memory": "On", "Auto-save": "On", "Context retrieval": "Hybrid", "Knowledge Graph": "Enabled",
	"Web": "Enabled", "Files": "Enabled", "GitHub": "Connected", "Terminal": "Enabled", "MCP": "5 servers",
	"Vault connection": "Connected", "Default vault": "~/huginn-vault", "Sync": "Auto", "Indexing": "On", "Embeddings": "Enabled",
	"API": "Local", "Logs": "Verbose", "Debug": "Off", "Experimental": "Off",
}
