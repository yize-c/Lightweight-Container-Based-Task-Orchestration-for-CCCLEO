// Package job defines the simplified job description format used across
// the controller and worker agent. This replaces Kubernetes-style YAML
// with a handful of fields: image, command, node target, and retry rule.
package job

import "time"

// Status values a job can be in.
const (
	StatusPending   = "pending"   // submitted, not yet picked up by a worker
	StatusAssigned  = "assigned"  // handed to a worker (push) or claimed by one (pull)
	StatusRunning   = "running"   // worker is executing the container
	StatusSucceeded = "succeeded" // container exited 0
	StatusFailed    = "failed"    // container exited non-zero, or retries exhausted
)

// Spec is the simplified job description a client submits via REST.
// It intentionally has far fewer fields than a Kubernetes Pod spec.
type Spec struct {
	Image      string            `json:"image" yaml:"image"`             // container image to run
	Command    []string          `json:"command,omitempty" yaml:"command,omitempty"` // override CMD, optional
	Env        map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	NodeTarget string            `json:"node_target,omitempty" yaml:"node_target,omitempty"` // worker ID or node label; empty = any worker
	MaxRetries int               `json:"max_retries" yaml:"max_retries"`                     // retry rule: how many times to retry on failure
}

// Job is a Spec plus controller-assigned bookkeeping. The controller holds
// the authoritative copy; workers see a trimmed view when they poll.
type Job struct {
	ID         string    `json:"id"`
	Spec       Spec      `json:"spec"`
	Status     string    `json:"status"`
	AssignedTo string    `json:"assigned_to,omitempty"` // worker ID that owns this job
	Attempts   int       `json:"attempts"`
	Result     string    `json:"result,omitempty"` // last exit message / error
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StatusReport is what a worker POSTs back after running (or failing to run) a job.
type StatusReport struct {
	JobID    string `json:"job_id"`
	WorkerID string `json:"worker_id"`
	Status   string `json:"status"` // running | succeeded | failed
	Result   string `json:"result,omitempty"`
}

// RegisterRequest is what a worker sends once at startup to announce itself.
type RegisterRequest struct {
	WorkerID string            `json:"worker_id"`
	Labels   map[string]string `json:"labels,omitempty"` // e.g. arch=arm64, region=sat-link-1
}
