package vault

import "time"

// Vault represents a Huginn workspace — a folder that contains .huginn
type Vault struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int       `json:"version"`
}

// VaultConfig is the global config inside .huginn/config.json
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

// VaultState holds runtime state (not versioned in git)
type VaultState struct {
	LastOpened time.Time `json:"lastOpened"`
	OpenCount  int       `json:"openCount"`
}
