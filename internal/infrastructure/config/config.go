// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config persiste en ~/.huginn/config.json (ver Path).
// La versión 2 añade planner/agentes/timeouts/reintentos/permisos/MCP/espacio/logs.
// Load migra ficheros v1 rellenando valores por defecto.
type Config struct {
	Version int    `json:"version"`
	Theme   string `json:"theme"`
	Memory  struct {
		Enabled    bool `json:"enabled"`
		Learning   bool `json:"learning"`
		AutoSave   bool `json:"auto_save"`
		MaxResults int  `json:"max_results"`
	} `json:"memory"`
	Agents struct {
		Default string `json:"default"`
	} `json:"agents"`
	Planner struct {
		MaxReplans  int `json:"max_replans"`
		MaxParallel int `json:"max_parallel"`
	} `json:"planner"`
	Timeouts struct {
		TaskSeconds   int `json:"task_seconds"`
		DetectSeconds int `json:"detect_seconds"`
	} `json:"timeouts"`
	Retry struct {
		MaxAttempts int `json:"max_attempts"`
		BackoffSecs int `json:"backoff_secs"`
	} `json:"retry"`
	Permissions struct {
		// AutoApprove desactiva la confirmación humana; por defecto es falso.
		AutoApprove bool `json:"auto_approve"`
		// AllowedWriteAgents concede escritura en disco a estos runtimes.
		AllowedWriteAgents []string `json:"allowed_write_agents"`
	} `json:"permissions"`
	MCP struct {
		// EnabledServers limita los servidores MCP; vacío usa valores por capacidad.
		EnabledServers []string `json:"enabled_servers"`
	} `json:"mcp"`
	Workspace struct {
		Path string `json:"path"`
	} `json:"workspace"`
	Logging struct {
		Level         string `json:"level"`
		RedactSecrets bool   `json:"redact_secrets"`
	} `json:"logging"`
	Models struct {
		OllamaHost  string `json:"ollama_host"`
		OllamaModel string `json:"ollama_model"`
		// OllamaModels asocia cada capacidad más "chat" con un modelo Ollama.
		// Las entradas vacías usan OllamaModel como reserva.
		OllamaModels  map[string]string `json:"ollama_models,omitempty"`
		OpenCodeModel string            `json:"opencode_model"`
		OpenCodeAgent string            `json:"opencode_agent"`
		// HuggingFace remota (sin modelos locales). Vacío = deshabilitado.
		HuggingFaceToken  string            `json:"huggingface_token,omitempty"`
		HuggingFaceModel  string            `json:"huggingface_model,omitempty"`
		HuggingFaceModels map[string]string `json:"huggingface_models,omitempty"`
	} `json:"models"`
}

// ChatModelKey es la clave de OllamaModels para el chat conversacional.
// Evita importar models para mantener config solo con stdlib.
const ChatModelKey = "chat"

// DefaultOllamaModels define la matriz local: Qwen para código y Llama para texto y chat.
func DefaultOllamaModels() map[string]string {
	return map[string]string{
		"coding":        "qwen3",
		"testing":       "qwen3",
		"devops":        "qwen3",
		"filesystem":    "qwen3",
		"git":           "qwen3",
		"browser":       "qwen3",
		"research":      "llama3.1",
		"reasoning":     "llama3.1",
		"documentation": "llama3.1",
		"review":        "llama3.1",
		ChatModelKey:    "llama3.1",
	}
}

// CurrentVersion es la versión de esquema que escribe Save.
const CurrentVersion = 3

func Default() Config {
	var c Config
	c.Version = CurrentVersion
	c.Theme = "vikingpunk"
	c.Memory.Enabled = true
	c.Memory.Learning = true
	c.Memory.AutoSave = false
	c.Memory.MaxResults = 10
	c.Agents.Default = "planner"
	c.Planner.MaxReplans = 1
	c.Planner.MaxParallel = 4
	c.Timeouts.TaskSeconds = 90
	c.Timeouts.DetectSeconds = 10
	c.Retry.MaxAttempts = 2
	c.Retry.BackoffSecs = 2
	c.Permissions.AutoApprove = false
	c.Permissions.AllowedWriteAgents = []string{"opencode"}
	c.Models.OllamaModels = DefaultOllamaModels()
	c.Logging.Level = "info"
	c.Logging.RedactSecrets = true
	return c
}

func Path(baseDir string) string {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".huginn")
	}
	return filepath.Join(baseDir, "config.json")
}

func Load(baseDir string) (Config, error) {
	p := Path(baseDir)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Default(), err
	}
	if c.Version == 0 {
		c = Default()
		return c, nil
	}
	if c.Version < CurrentVersion {
		c = migrate(c)
	}
	return c, nil
}

// migrate rellena con defectos los campos nuevos sin perder valores de usuario.
func migrate(c Config) Config {
	d := Default()
	if c.Theme == "" {
		c.Theme = d.Theme
	}
	if c.Planner.MaxReplans == 0 {
		c.Planner.MaxReplans = d.Planner.MaxReplans
	}
	if c.Planner.MaxParallel == 0 {
		c.Planner.MaxParallel = d.Planner.MaxParallel
	}
	if c.Timeouts.TaskSeconds == 0 {
		c.Timeouts.TaskSeconds = d.Timeouts.TaskSeconds
	}
	if c.Timeouts.DetectSeconds == 0 {
		c.Timeouts.DetectSeconds = d.Timeouts.DetectSeconds
	}
	if c.Retry.MaxAttempts == 0 {
		c.Retry.MaxAttempts = d.Retry.MaxAttempts
	}
	if c.Retry.BackoffSecs == 0 {
		c.Retry.BackoffSecs = d.Retry.BackoffSecs
	}
	if c.Permissions.AllowedWriteAgents == nil {
		c.Permissions.AllowedWriteAgents = d.Permissions.AllowedWriteAgents
	}
	// Fusiona por clave: gana el valor de usuario y los faltantes usan defecto.
	merged := d.Models.OllamaModels
	for k, v := range c.Models.OllamaModels {
		merged[k] = v
	}
	c.Models.OllamaModels = merged
	if c.Logging.Level == "" {
		c.Logging.Level = d.Logging.Level
	}
	if c.Version == 1 {
		// v1 no tenía flag de redacción; v2 lo activa por defecto.
		c.Logging.RedactSecrets = true
	}
	c.Version = CurrentVersion
	return c
}

func Save(baseDir string, c Config) error {
	if err := os.MkdirAll(filepath.Dir(Path(baseDir)), 0755); err != nil {
		return err
	}
	c.Version = CurrentVersion
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(Path(baseDir), b, 0644)
}
