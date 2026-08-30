package mcp

type Transport string

const (
	Stdio     Transport = "stdio"
	SSE       Transport = "sse"
	WebSocket Transport = "websocket"
)

type Server struct {
	Name      string
	Command   string
	Transport Transport
	Status    string
	Tools     int
	Latency   string
	ErrorMsg  string
	Enabled   bool
}

var DefaultServers = []Server{
	{"filesystem", "npx @modelcontextprotocol/server-filesystem", Stdio, "Connected", 8, "12ms", "", true},
	{"github", "npx @modelcontextprotocol/server-github", Stdio, "Connected", 6, "34ms", "", true},
	{"memory", "npx @modelcontextprotocol/server-memory", Stdio, "Connected", 4, "8ms", "", true},
	{"playwright", "npx @modelcontextprotocol/server-playwright", Stdio, "Connected", 7, "45ms", "", true},
	{"sequential", "node sequential-thinking", Stdio, "Connected", 1, "9ms", "", true},
	{"notion", "npx @modelcontextprotocol/server-notion", SSE, "Needs auth", 0, "—", "Missing NOTION_API_KEY", true},
	{"vault-sync", "huginn-vault-sync", WebSocket, "Disabled", 3, "—", "", false},
}
