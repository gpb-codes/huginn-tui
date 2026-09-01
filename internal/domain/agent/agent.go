package agent

import "os/exec"

// Status represents agent execution state.
type Status int

const (
	StatusDone Status = iota
	StatusWorking
	StatusWaiting
	StatusTesting
)

type Agent struct {
	ID           string
	Name         string
	Role         string
	Description  string
	Capabilities []string
	Status       Status
	Pct          int
}

func (s Status) Label() string {
	switch s {
	case StatusDone:
		return "Completed"
	case StatusWorking:
		return "Working"
	case StatusTesting:
		return "Running Tests"
	default:
		return "Queued"
	}
}

// BackendAgent is the external tool that backs an agent (opencode, kilocode, etc.)
type BackendAgent struct {
	Name        string
	Description string
	Command     string // empty = always online
}

var BackendAgents = []BackendAgent{
	{"ChatGPT", "central intelligence manager", ""},
	{"OpenCode", "terminal AI coding agent", "opencode"},
	{"KiloCode", "AI coding agent and workflow", "kilocode"},
}

func CommandAvailable(cmd string) bool {
	if cmd == "" {
		return true
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}
