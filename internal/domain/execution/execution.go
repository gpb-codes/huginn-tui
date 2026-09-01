package execution

import "time"

// Record — trazabilidad de cada ejecucion agente/provider
type Record struct {
	ExecutionID string    `json:"execution_id"`
	Agent       string    `json:"agent"`
	Provider    string    `json:"provider"`
	TaskID      string    `json:"task_id"`
	TaskType    string    `json:"task_type"`
	Input       string    `json:"input"`
	Status      string    `json:"status"`
	Result      any       `json:"result,omitempty"`
	Errors      []string  `json:"errors,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`
	Latency     int64     `json:"latency_ms"`
}
