package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"huginn/internal/domain/memory"
)

// MarkdownStore implements ports.MemoryPort using Markdown + frontmatter on filesystem.
// Source of truth: Markdown files in ~/.huginn/memory/*.md
// Index: ~/.huginn/memory/index.jsonl (rebuildable)
// Events: ~/.huginn/events.jsonl (append-only)
type MarkdownStore struct {
	BaseDir string // ~/.huginn
}

func NewMarkdownStore(baseDir string) *MarkdownStore {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".huginn")
	}
	return &MarkdownStore{BaseDir: baseDir}
}

func (s *MarkdownStore) memoryDir() string  { return filepath.Join(s.BaseDir, "memory") }
func (s *MarkdownStore) indexPath() string  { return filepath.Join(s.memoryDir(), "index.jsonl") }
func (s *MarkdownStore) eventsPath() string { return filepath.Join(s.BaseDir, "events.jsonl") }

func (s *MarkdownStore) ensureDirs() error {
	return os.MkdirAll(s.memoryDir(), 0755)
}

// Save writes a Memory as Markdown with frontmatter.
func (s *MarkdownStore) Save(ctx context.Context, m memory.Memory) error {
	if err := s.ensureDirs(); err != nil {
		return err
	}
	if err := assertNoSecrets(m.Content); err != nil {
		return err
	}
	if m.ID == "" {
		return fmt.Errorf("memory id required: %w", ErrInvalidInput)
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.UpdatedAt = time.Now()
	path := filepath.Join(s.memoryDir(), m.ID+".md")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	// frontmatter
	fmt.Fprintln(f, "---")
	fmt.Fprintf(f, "id: %s\n", m.ID)
	fmt.Fprintf(f, "type: %s\n", m.Type)
	fmt.Fprintf(f, "importance: %v\n", m.Importance)
	fmt.Fprintf(f, "confidence: %v\n", m.Confidence)
	fmt.Fprintf(f, "created: %s\n", m.CreatedAt.Format("2006-01-02"))
	fmt.Fprintf(f, "updated: %s\n", m.UpdatedAt.Format("2006-01-02"))
	if len(m.Tags) > 0 {
		fmt.Fprintln(f, "tags:")
		for _, t := range m.Tags {
			fmt.Fprintf(f, "  - %s\n", t)
		}
	}
	fmt.Fprintln(f, "---")
	fmt.Fprintln(f, "")
	fmt.Fprintf(f, "# %s\n\n", m.Title)
	fmt.Fprintln(f, m.Content)
	// update index
	_ = s.appendIndex(m)
	_ = s.appendEvent("memory.created", m.ID)
	return nil
}

func (s *MarkdownStore) Get(_ context.Context, id string) (*memory.Memory, error) {
	path := filepath.Join(s.memoryDir(), id+".md")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m, err := parseMarkdown(id, string(b))
	if err != nil {
		return nil, err
	}
	m.File = path
	return m, nil
}

func (s *MarkdownStore) List(_ context.Context, memoryType string) ([]memory.Memory, error) {
	entries, err := os.ReadDir(s.memoryDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []memory.Memory
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, _ := os.ReadFile(filepath.Join(s.memoryDir(), e.Name()))
		m, _ := parseMarkdown(strings.TrimSuffix(e.Name(), ".md"), string(b))
		if m == nil {
			continue
		}
		if memoryType != "" && m.Type != memoryType {
			continue
		}
		out = append(out, *m)
	}
	return out, nil
}

func (s *MarkdownStore) Search(_ context.Context, query string) ([]memory.Memory, error) {
	all, err := s.List(context.Background(), "")
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var res []memory.Memory
	for _, m := range all {
		if strings.Contains(strings.ToLower(m.Title), q) ||
			strings.Contains(strings.ToLower(m.Content), q) ||
			strings.Contains(strings.ToLower(strings.Join(m.Tags, " ")), q) ||
			strings.Contains(strings.ToLower(m.Type), q) {
			res = append(res, m)
		}
		// limit
		if len(res) >= 10 {
			break
		}
	}
	return res, nil
}

func (s *MarkdownStore) Delete(_ context.Context, id string) error {
	path := filepath.Join(s.memoryDir(), id+".md")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	_ = s.appendEvent("memory.deleted", id)
	// rebuild index lazily
	_ = s.RebuildIndex()
	return nil
}

// IndexEntry is a line in index.jsonl
type IndexEntry struct {
	ID         string  `json:"id"`
	File       string  `json:"file"`
	Type       string  `json:"type"`
	Importance float64 `json:"importance"`
	UpdatedAt  string  `json:"updated_at"`
}

func (s *MarkdownStore) appendIndex(m memory.Memory) error {
	entry := IndexEntry{
		ID:         m.ID,
		File:       filepath.Join("memory", m.ID+".md"),
		Type:       m.Type,
		Importance: m.Importance,
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
	}
	f, err := os.OpenFile(s.indexPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	line, _ := json.Marshal(entry)
	_, err = f.Write(append(line, '\n'))
	return err
}

func (s *MarkdownStore) appendEvent(event, memoryID string) error {
	f, err := os.OpenFile(s.eventsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	rec := map[string]string{
		"event":     event,
		"memory_id": memoryID,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	line, _ := json.Marshal(rec)
	_, err = f.Write(append(line, '\n'))
	return err
}

// RebuildIndex reconstructs index.jsonl from Markdown files.
func (s *MarkdownStore) RebuildIndex() error {
	all, _ := s.List(context.Background(), "")
	_ = os.Remove(s.indexPath())
	for _, m := range all {
		_ = s.appendIndex(m)
	}
	return nil
}

// Errors
var (
	ErrNotFound       = fmt.Errorf("not found")
	ErrInvalidInput   = fmt.Errorf("invalid input")
	ErrMemoryRejected = fmt.Errorf("memory rejected: potential secret")
)

var secretPattern = regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token|private[_-]?key)\s*[:=]\s*\S{8,}`)
var knownSecretTokens = regexp.MustCompile(`(sk-[A-Za-z0-9]{20,}|ghp_[A-Za-z0-9]{20,}|AKIA[0-9A-Z]{16}|-----BEGIN (RSA )?PRIVATE KEY-----)`)

func assertNoSecrets(content string) error {
	if secretPattern.MatchString(content) {
		return ErrMemoryRejected
	}
	if knownSecretTokens.MatchString(content) {
		return ErrMemoryRejected
	}
	return nil
}

func parseMarkdown(id, content string) (*memory.Memory, error) {
	// very simple frontmatter parser: between --- ... ---
	m := &memory.Memory{ID: id}
	lines := strings.Split(content, "\n")
	inFront := false
	var fmLines []string
	var bodyLines []string
	for i, l := range lines {
		if i == 0 && strings.TrimSpace(l) == "---" {
			inFront = true
			continue
		}
		if inFront && strings.TrimSpace(l) == "---" {
			inFront = false
			continue
		}
		if inFront {
			fmLines = append(fmLines, l)
		} else {
			bodyLines = append(bodyLines, l)
		}
	}
	for _, l := range fmLines {
		if strings.HasPrefix(l, "type:") {
			m.Type = strings.TrimSpace(strings.TrimPrefix(l, "type:"))
		} else if strings.HasPrefix(l, "importance:") {
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(l, "importance:")), "%f", &m.Importance)
		} else if strings.HasPrefix(l, "confidence:") {
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(l, "confidence:")), "%f", &m.Confidence)
		}
	}
	// title is first # line
	for _, l := range bodyLines {
		if strings.HasPrefix(strings.TrimSpace(l), "# ") {
			m.Title = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "# "))
			break
		}
	}
	m.Content = strings.Join(bodyLines, "\n")
	// try to get updated from file? use now
	m.UpdatedAt = time.Now()
	m.CreatedAt = m.UpdatedAt
	// tags not parsed fully for brevity
	return m, nil
}

// Ensure implements ports.MemoryPort
var _ = (*MarkdownStore)(nil)

// Helper to read index without loading all markdown
func (s *MarkdownStore) ReadIndex() ([]IndexEntry, error) {
	f, err := os.Open(s.indexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []IndexEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e IndexEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			out = append(out, e)
		}
	}
	return out, scanner.Err()
}
