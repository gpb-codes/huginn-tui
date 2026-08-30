package task

import "time"

// Result holds the outcome of a Task execution.
type Result struct {
	Success     bool
	Output      string
	Error       string
	Artifacts   []string
	Duration    time.Duration
	CompletedAt time.Time
}

func SuccessResult(output string, artifacts ...string) *Result {
	return &Result{Success: true, Output: output, Artifacts: artifacts, CompletedAt: time.Now()}
}

func FailedResult(err string) *Result {
	return &Result{Success: false, Error: err, CompletedAt: time.Now()}
}
