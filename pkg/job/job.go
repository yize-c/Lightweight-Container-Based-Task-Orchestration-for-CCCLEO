// Package job defines the simplified job description format used across
// the controller and worker agent. This replaces Kubernetes-style YAML
// with a handful of fields: image, command, node target, and retry rule.
package job

import "time"

// Status values a job can be in. Just three real states plus the start:
// pending -> running -> succeeded/failed. ("Claimed by a worker" and
// "actually executing" are treated as the same moment — a worker that
// just claimed a job is about to run it, so there's no useful distinction
// between an "assigned" and a "running" state here.)
const (
	StatusPending   = "pending"   // submitted, not yet picked up by a worker
	StatusRunning   = "running"   // claimed by a worker (pull) or handed to one (push)
	StatusSucceeded = "succeeded" // container exited 0
	StatusFailed    = "failed"    // container exited non-zero, or retries exhausted
)

// Spec is the simplified job description a client submits via REST.
// It intentionally has far fewer fields than a Kubernetes Pod spec.
type Spec struct {
	Image      string            `json:"image" yaml:"image"`             // container image to run
	Command    []string          `json:"command,omitempty" yaml:"command,omitempty"` // override CMD, optional
	Env        map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	NodeTarget string            `json:"node_target,omitempty" yaml:"node_target,omitempty"` // worker ID; empty = any worker
	MaxRetries int               `json:"max_retries" yaml:"max_retries"`                     // retry rule: how many times to retry on failure
}

// Job is a Spec plus controller-assigned bookkeeping. The controller holds
// the authoritative copy; workers see a trimmed view when they poll.
type Job struct {
	ID         string     `json:"id"`
	Spec       Spec       `json:"spec"`
	Status     string     `json:"status"`
	AssignedTo string     `json:"assigned_to,omitempty"` // worker ID that owns this job
	Attempts   int        `json:"attempts"`
	Result     string     `json:"result,omitempty"` // last exit message / error
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	AssignedAt *time.Time `json:"assigned_at,omitempty"` // when it left "pending" (claimed/pushed) -> scheduling delay = AssignedAt - CreatedAt
	FinishedAt *time.Time `json:"finished_at,omitempty"` // when it reached succeeded/failed -> total latency = FinishedAt - CreatedAt
}

// StatusReport is what a worker POSTs back once a job finishes.
type StatusReport struct {
	JobID    string `json:"job_id"`
	WorkerID string `json:"worker_id"`
	Status   string `json:"status"` // succeeded | failed
	Result   string `json:"result,omitempty"`
}

// Metrics is the controller-side aggregate the proposal's research
// questions need: completion rate, scheduling delay, duplicate
// executions avoided.
type Metrics struct {
	TotalJobs            int     `json:"total_jobs"`
	Pending              int     `json:"pending"`
	Running              int     `json:"running"`
	Succeeded            int     `json:"succeeded"`
	Failed               int     `json:"failed"`
	CompletionRate       float64 `json:"completion_rate"`         // succeeded / total_jobs
	AvgSchedulingDelayMs float64 `json:"avg_scheduling_delay_ms"` // mean(AssignedAt - CreatedAt) over jobs that have been assigned
	AvgTotalLatencyMs    float64 `json:"avg_total_latency_ms"`    // mean(FinishedAt - CreatedAt) over finished jobs
	StaleReportsIgnored  int     `json:"stale_reports_ignored"`   // duplicate-execution guard hit count
}

// RegisterRequest is what a worker sends once at startup to announce itself.
type RegisterRequest struct {
	WorkerID string `json:"worker_id"`
}
