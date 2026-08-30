package lsp

type Server struct {
	Language    string
	ServerName  string
	Command     string
	Status      string
	Root        string
	Diagnostics int
	Version     string
	Enabled     bool
}

var DefaultServers = []Server{
	{"Go", "gopls", "gopls", "Running", "./", 0, "v0.15.2", true},
	{"TypeScript", "tsserver", "typescript-language-server --stdio", "Running", "web/", 2, "4.3.3", true},
	{"Python", "pyright", "pyright-langserver --stdio", "Stopped", "py/", 0, "1.1.350", false},
	{"Rust", "rust-analyzer", "rust-analyzer", "Running", "./", 1, "2024-08-28", true},
	{"Lua", "lua_ls", "lua-language-server", "Error", "lua/", 0, "3.7.4", true},
}
