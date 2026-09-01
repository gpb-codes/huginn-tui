package runtime

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"

	"huginn/internal/application/ports"
	"huginn/internal/domain/task"
)

// ProcessRuntime executes tasks as external processes.
// It is provider-agnostic: command is determined by task.AgentID mapping.
type ProcessRuntime struct {
	// AgentCommand maps agentID -> command (e.g., "opencode", "claude", "ollama")
	AgentCommand map[string]string
}

func NewProcessRuntime(mapping map[string]string) *ProcessRuntime {
	if mapping == nil {
		mapping = map[string]string{
			"planner":    "echo",
			"coder":      "echo",
			"researcher": "echo",
			"reviewer":   "echo",
		}
	}
	return &ProcessRuntime{AgentCommand: mapping}
}

func (r *ProcessRuntime) Run(ctx context.Context, t task.Task) (<-chan ports.AgentEvent, error) {
	cmdStr, ok := r.AgentCommand[t.AgentID]
	if !ok {
		cmdStr = "echo"
	}
	ch := make(chan ports.AgentEvent, 16)
	cmd := exec.CommandContext(ctx, cmdStr, t.Title)
	go func() {
		defer close(ch)
		ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentStarted, Content: fmt.Sprintf("starting %s", t.Title), Timestamp: time.Now().Unix()}
		// If echo, just emit tool result
		if cmdStr == "echo" {
			ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentMessage, Content: "mock output for " + t.Title, Timestamp: time.Now().Unix()}
			ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentCompleted, Content: "done", Timestamp: time.Now().Unix()}
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentFailed, Content: err.Error(), Timestamp: time.Now().Unix()}
			return
		}
		if err := cmd.Start(); err != nil {
			ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentFailed, Content: err.Error(), Timestamp: time.Now().Unix()}
			return
		}
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentMessage, Content: scanner.Text(), Timestamp: time.Now().Unix()}
			}
		}
		_ = cmd.Wait()
		ch <- ports.AgentEvent{TaskID: t.ID, AgentID: t.AgentID, Type: ports.EventAgentCompleted, Content: "completed", Timestamp: time.Now().Unix()}
	}()
	return ch, nil
}
