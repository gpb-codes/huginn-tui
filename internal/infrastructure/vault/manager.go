package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	domain "huginn/internal/domain/vault"
)

// FilesystemManager implements domain.Manager via filesystem.
type FilesystemManager struct {
	current *domain.Vault
}

func NewFilesystemManager() *FilesystemManager { return &FilesystemManager{} }

func (m *FilesystemManager) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m *FilesystemManager) IsInitialized(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".huginn", "vault.json"))
	return err == nil
}

func (m *FilesystemManager) Detect(startPath string) (*domain.Vault, bool) {
	abs, err := filepath.Abs(startPath)
	if err != nil {
		return nil, false
	}
	for {
		if m.IsInitialized(abs) {
			// load vault.json
			v, err := m.loadVault(abs)
			if err == nil {
				return v, true
			}
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return nil, false
}

func (m *FilesystemManager) Open(_ context.Context, path string) (*domain.Vault, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !m.Exists(abs) {
		return nil, fmt.Errorf("vault path does not exist: %s", abs)
	}
	var v *domain.Vault
	if m.IsInitialized(abs) {
		v, err = m.loadVault(abs)
		if err != nil {
			return nil, fmt.Errorf("vault.json corrupted: %w", err)
		}
	} else {
		// auto-initialize
		v, err = m.Initialize(context.Background(), abs)
		if err != nil {
			return nil, err
		}
	}
	v.UpdatedAt = time.Now()
	_ = m.saveVault(v)
	_ = m.AddRecent(*v)
	m.current = v
	_ = m.updateState(v.Path)
	return v, nil
}

func (m *FilesystemManager) Create(_ context.Context, parentDir, name string) (*domain.Vault, error) {
	if name == "" {
		return nil, fmt.Errorf("vault name required")
	}
	parentAbs, err := filepath.Abs(parentDir)
	if err != nil {
		return nil, err
	}
	if !m.Exists(parentAbs) {
		return nil, fmt.Errorf("parent does not exist: %s", parentAbs)
	}
	vaultPath := filepath.Join(parentAbs, name)
	if m.Exists(vaultPath) {
		return nil, fmt.Errorf("vault already exists: %s", vaultPath)
	}
	if err := os.MkdirAll(vaultPath, 0755); err != nil {
		return nil, err
	}
	return m.Initialize(context.Background(), vaultPath)
}

func (m *FilesystemManager) Initialize(_ context.Context, path string) (*domain.Vault, error) {
	abs, _ := filepath.Abs(path)
	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, err
	}
	huginnDir := filepath.Join(abs, ".huginn")
	if err := os.MkdirAll(huginnDir, 0755); err != nil {
		return nil, err
	}
	// create subdirs
	for _, dir := range []string{"agents", "memory", "plugins", "cache", "logs", "runtime"} {
		_ = os.MkdirAll(filepath.Join(huginnDir, dir), 0755)
	}
	for _, dir := range []string{"notes", "projects", "agents", "memory", "attachments"} {
		_ = os.MkdirAll(filepath.Join(abs, dir), 0755)
	}
	// README
	readmePath := filepath.Join(abs, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		_ = os.WriteFile(readmePath, []byte("# "+filepath.Base(abs)+"\n\nHuginn Vault — initialized "+time.Now().Format("2006-01-02")+"\n"), 0644)
	}
	// .gitignore
	gitignorePath := filepath.Join(abs, ".gitignore")
	if _, err := os.Stat(gitignorePath); err != nil {
		content := "# Huginn runtime/cache\n.huginn/cache/\n.huginn/runtime/\n.huginn/logs/\n.huginn/state.json\n"
		_ = os.WriteFile(gitignorePath, []byte(content), 0644)
	} else {
		// ensure entries exist
		b, _ := os.ReadFile(gitignorePath)
		s := string(b)
		for _, entry := range []string{".huginn/cache/", ".huginn/runtime/", ".huginn/logs/"} {
			if !contains(s, entry) {
				f, _ := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
				if f != nil {
					_, _ = f.WriteString("\n" + entry + "\n")
					f.Close()
				}
			}
		}
	}
	// vault.json
	v := &domain.Vault{
		ID:        generateID(),
		Name:      filepath.Base(abs),
		Path:      abs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}
	if err := m.saveVault(v); err != nil {
		return nil, err
	}
	// config.json
	cfg := domain.DefaultVaultConfig(v.Name)
	cfgPath := filepath.Join(huginnDir, "config.json")
	if _, err := os.Stat(cfgPath); err != nil {
		b, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(cfgPath, b, 0644)
	}
	// other jsons
	for _, name := range []string{"agents.json", "memory.jsonl", "plugins.json", "state.json"} {
		p := filepath.Join(huginnDir, name)
		if _, err := os.Stat(p); err != nil {
			_ = os.WriteFile(p, []byte("{}"), 0644)
		}
	}
	// vault state
	_ = m.updateState(abs)
	m.current = v
	_ = m.AddRecent(*v)
	return v, nil
}

func (m *FilesystemManager) Close(_ context.Context) error {
	m.current = nil
	return nil
}

func (m *FilesystemManager) GetCurrent() (*domain.Vault, bool) {
	if m.current == nil {
		return nil, false
	}
	return m.current, true
}

func (m *FilesystemManager) GetPath() string {
	if m.current == nil {
		return ""
	}
	return m.current.Path
}

func (m *FilesystemManager) Recent() ([]domain.Vault, error) {
	home, _ := os.UserHomeDir()
	recentPath := filepath.Join(home, ".huginn", "recent_vaults.json")
	b, err := os.ReadFile(recentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var list []domain.Vault
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *FilesystemManager) AddRecent(vault domain.Vault) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".huginn")
	_ = os.MkdirAll(dir, 0755)
	recentPath := filepath.Join(dir, "recent_vaults.json")
	var list []domain.Vault
	b, _ := os.ReadFile(recentPath)
	if len(b) > 0 {
		_ = json.Unmarshal(b, &list)
	}
	// dedup by path, move to front
	var newList []domain.Vault
	newList = append(newList, vault)
	for _, v := range list {
		if v.Path != vault.Path && v.ID != vault.ID {
			newList = append(newList, v)
		}
	}
	if len(newList) > 10 {
		newList = newList[:10]
	}
	out, _ := json.MarshalIndent(newList, "", "  ")
	return os.WriteFile(recentPath, out, 0644)
}

func (m *FilesystemManager) loadVault(path string) (*domain.Vault, error) {
	b, err := os.ReadFile(filepath.Join(path, ".huginn", "vault.json"))
	if err != nil {
		return nil, err
	}
	var v domain.Vault
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	// ensure Path is absolute
	if v.Path == "" {
		v.Path = path
	}
	return &v, nil
}

func (m *FilesystemManager) saveVault(v *domain.Vault) error {
	v.UpdatedAt = time.Now()
	b, _ := json.MarshalIndent(v, "", "  ")
	return os.WriteFile(filepath.Join(v.Path, ".huginn", "vault.json"), b, 0644)
}

func (m *FilesystemManager) updateState(vaultPath string) error {
	statePath := filepath.Join(vaultPath, ".huginn", "state.json")
	var state domain.VaultState
	b, _ := os.ReadFile(statePath)
	if len(b) > 0 {
		_ = json.Unmarshal(b, &state)
	}
	state.LastOpened = time.Now()
	state.OpenCount++
	out, _ := json.MarshalIndent(state, "", "  ")
	return os.WriteFile(statePath, out, 0644)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
