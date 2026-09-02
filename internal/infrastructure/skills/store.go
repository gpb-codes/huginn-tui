package skills

import (
	"os"
	"path/filepath"
	"strings"

	"huginn/internal/domain/skill"
)

// Store — filesystem ~/.huginn/skills/<name>/SKILL.md
type Store struct{ base string }

func New(base string) *Store { return &Store{base: base} }

func (s *Store) List() ([]skill.Skill, error) {
	entries, err := os.ReadDir(s.base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []skill.Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.base, e.Name(), "SKILL.md"))
		if err != nil {
			continue
		}
		sk := skill.Skill{Name: e.Name(), Description: strings.Split(string(b), "\n")[0]}
		out = append(out, sk)
	}
	return out, nil
}

func (s *Store) Save(sk skill.Skill, body string) error {
	dir := filepath.Join(s.base, sk.Name)
	_ = os.MkdirAll(dir, 0755)
	return os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0644)
}
