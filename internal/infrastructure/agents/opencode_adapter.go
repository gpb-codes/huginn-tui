// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"huginn/internal/domain/agent"
)

// Errores tipados para distinguir la causa en llamantes y evaluador.
var (
	ErrNotInstalled     = errors.New("opencode: executable not found in PATH")
	ErrInvalidWorkspace = errors.New("opencode: invalid workspace directory")
	ErrTimeout          = errors.New("opencode: execution timed out")
	ErrProcessCrash     = errors.New("opencode: process failed")
)

// Config ajusta el adaptador; los valores cero seleccionan defectos.
type Config struct {
	// Bin sustituye la resolución del ejecutable (tests lo usan para dobles).
	Bin string
	// Model se pasa como --model; vacío usa el defecto de OpenCode.
	Model string
	// Agent se pasa como --agent (principal/subagente de .opencode/).
	Agent string
	// Timeout limita una ejecución; <=0 usa 90s.
	Timeout time.Duration
}

// OpenCodeAdapter ejecuta tareas de código con el CLI real de OpenCode.
// Usa `opencode run --format json …` (ver https://opencode.ai/docs/cli/).
// Huginn orquesta y OpenCode ejecuta; la planificación no depende de esto.
type OpenCodeAdapter struct {
	cfg  Config
	bin  string
	exec func(ctx context.Context, bin string, args []string, dir string) ([]byte, error)
}

// NewOpenCodeAdapter resuelve el binario desde PATH.
func NewOpenCodeAdapter() *OpenCodeAdapter {
	return NewOpenCodeAdapterWithConfig(Config{})
}

// NewOpenCodeAdapterWithConfig construye el adaptador con configuración explícita.
func NewOpenCodeAdapterWithConfig(cfg Config) *OpenCodeAdapter {
	bin := cfg.Bin
	if bin == "" {
		bin = resolveBin("opencode")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 90 * time.Second
	}
	return &OpenCodeAdapter{cfg: cfg, bin: bin, exec: realExec}
}

func realExec(ctx context.Context, bin string, args []string, dir string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	// Privilegio mínimo: entorno reducido sin heredar secretos del padre.
	cmd.Env = minimalEnv()
	return cmd.CombinedOutput()
}

func minimalEnv() []string {
	keep := []string{"PATH", "HOME", "USERPROFILE", "HOMEDRIVE", "HOMEPATH", "SYSTEMROOT",
		"TEMP", "TMP", "LANG", "LC_ALL", "TERM", "NO_COLOR", "OPENCODE_CONFIG", "OPENCODE_CONFIG_DIR"}
	var out []string
	for _, k := range keep {
		if v, ok := os.LookupEnv(k); ok {
			out = append(out, k+"="+v)
		}
	}
	return out
}

func (a *OpenCodeAdapter) ID() string   { return "opencode" }
func (a *OpenCodeAdapter) Name() string { return "OpenCode" }

// SetModel sustituye el modelo para ejecuciones posteriores.
func (a *OpenCodeAdapter) SetModel(m string) { a.cfg.Model = m }

// Capabilities declara las capacidades soportadas para el enrutado.
func (a *OpenCodeAdapter) Capabilities() []string {
	return []string{"code_generation", "code_editing", "terminal_execution", "file_operations", "git", "streaming", "tool_calling"}
}

// Detect indica si el binario es utilizable y su versión.
func (a *OpenCodeAdapter) Detect() (bool, string) {
	if a.bin == "" {
		return false, "NOT_INSTALLED"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := a.exec(ctx, a.bin, []string{"--version"}, "")
	if err != nil {
		return false, "ERROR"
	}
	return true, strings.TrimSpace(string(out))
}

// Execute ejecuta task.Input en el workspace (task.Context.ProjectPath si existe).
func (a *OpenCodeAdapter) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	fail := func(err error) (agent.AgentResult, error) {
		return agent.AgentResult{
			TaskID: task.ID, Agent: a.Name(), Provider: a.ID(),
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	if a.bin == "" {
		return fail(ErrNotInstalled)
	}
	dir := task.Context.ProjectPath
	if dir != "" {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return fail(fmt.Errorf("%w: %s", ErrInvalidWorkspace, dir))
		}
	}
	args := []string{"run", "--format", "json"}
	if a.cfg.Model != "" {
		args = append(args, "--model", a.cfg.Model)
	}
	if a.cfg.Agent != "" {
		args = append(args, "--agent", a.cfg.Agent)
	}
	args = append(args, task.Input)

	callCtx, cancel := context.WithTimeout(ctx, a.cfg.Timeout)
	defer cancel()
	out, err := a.exec(callCtx, a.bin, args, dir)
	raw := strings.TrimSpace(string(out))
	text := parseOpenCodeJSON(raw)
	if text == "" {
		text = raw
	}
	if callCtx.Err() == context.DeadlineExceeded {
		return fail(fmt.Errorf("%w after %s", ErrTimeout, a.cfg.Timeout))
	}
	if err != nil {
		// Salida no cero: conserva lo capturado para el evaluador.
		execErr := fmt.Errorf("%w: %v", ErrProcessCrash, err)
		res, _ := fail(execErr)
		res.Output = agent.CodeResult{Summary: text}
		return res, execErr
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: a.Name(), Provider: a.ID(), Status: "ok",
		Output:     agent.CodeResult{Summary: text},
		StartedAt:  start,
		FinishedAt: time.Now(),
		Latency:    time.Since(start),
	}, nil
}

func parseOpenCodeJSON(raw string) string {
	var text string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			if typ, ok := obj["type"].(string); ok && typ == "text" {
				if part, ok := obj["part"].(map[string]interface{}); ok {
					if t, ok := part["text"].(string); ok {
						text += t
					}
				}
			}
		}
	}
	return strings.TrimSpace(text)
}

func resolveBin(name string) string {
	if runtime.GOOS == "windows" {
		for _, cand := range []string{name + ".cmd", name + ".ps1", name + ".exe", name} {
			if _, err := exec.LookPath(cand); err == nil {
				return cand
			}
		}
		return ""
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}
