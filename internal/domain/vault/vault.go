// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package vault

import "time"

// Vault representa un espacio de trabajo Huginn: carpeta que contiene .huginn.
type Vault struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int       `json:"version"`
}

// VaultConfig es la configuración global en .huginn/config.json.
type VaultConfig struct {
	SchemaVersion int `json:"schemaVersion"`
	Version       int `json:"version"`
	Vault         struct {
		Name string `json:"name"`
	} `json:"vault"`
	Interface struct {
		Theme string `json:"theme"`
	} `json:"interface"`
	Agents struct {
		Enabled bool `json:"enabled"`
	} `json:"agents"`
	Memory struct {
		Enabled bool `json:"enabled"`
	} `json:"memory"`
	Plugins struct {
		Enabled bool `json:"enabled"`
	} `json:"plugins"`
}

func DefaultVaultConfig(name string) VaultConfig {
	var c VaultConfig
	c.SchemaVersion = 1
	c.Version = 1
	c.Vault.Name = name
	c.Interface.Theme = "default"
	c.Agents.Enabled = true
	c.Memory.Enabled = true
	c.Plugins.Enabled = true
	return c
}

// VaultState conserva el estado de ejecución (no versionado en git).
type VaultState struct {
	LastOpened time.Time `json:"lastOpened"`
	OpenCount  int       `json:"openCount"`
}
