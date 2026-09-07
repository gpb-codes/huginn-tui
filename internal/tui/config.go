// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TUIConfig refleja tui.json / tui.jsonc de opencode (solo lo que Huginn usa).
// Archivo: ~/.config/huginn/tui.jsonc  o  .huginn/tui.json  (proyecto).
type TUIConfig struct {
	Theme          string  `json:"theme"`           // "huginn" | "opencode" | "dark"
	ScrollSpeed    float64 `json:"scroll_speed"`    // 0.001..10, default 3
	DiffStyle      string  `json:"diff_style"`      // "auto" | "stacked"
	Mouse          bool    `json:"mouse"`
	LeaderTimeout  int     `json:"leader_timeout"`  // ms, default 2000
	KeybindsLeader string  `json:"keybinds.leader"` // default "ctrl+x"
}

func DefaultTUIConfig() TUIConfig {
	return TUIConfig{
		Theme:          "huginn",
		ScrollSpeed:    3,
		DiffStyle:      "auto",
		Mouse:          true,
		LeaderTimeout:  2000,
		KeybindsLeader: "ctrl+x",
	}
}

func tuiConfigPath() string {
	if home, _ := os.UserHomeDir(); home != "" {
		p := filepath.Join(home, ".config", "huginn", "tui.jsonc")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		q := filepath.Join(home, ".config", "huginn", "tui.json")
		if _, err := os.Stat(q); err == nil {
			return q
		}
		// por defecto, el de user
		return p
	}
	return "tui.jsonc"
}

func LoadTUIConfig() TUIConfig {
	cfg := DefaultTUIConfig()
	p := tuiConfigPath()
	b, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	// Permite comentarios // y /* */ estilo jsonc: los ignora de forma simple
	// Si falla el parseo, devuelve defaults (no rompe la TUI)
	var raw TUIConfig
	if err := json.Unmarshal(b, &raw); err != nil {
		return cfg
	}
	if raw.Theme != "" {
		cfg.Theme = raw.Theme
	}
	if raw.ScrollSpeed > 0 {
		cfg.ScrollSpeed = raw.ScrollSpeed
	}
	if raw.DiffStyle != "" {
		cfg.DiffStyle = raw.DiffStyle
	}
	// Mouse y LeaderTimeout se dejan por defecto si no vienen
	return cfg
}
