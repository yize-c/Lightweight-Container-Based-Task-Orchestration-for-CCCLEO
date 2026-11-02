package main

import (
	"sync"
	"time"

	"cccleo-orchestrator/pkg/job"
)

// Store is a thread-safe in-memory job/worker registry. It is the
// controller's "local job store" mentioned in the proposal: unique job
// IDs plus status tracking are what prevent duplicate execution after a
// worker reconnects.
type Store struct {
	mu      sync.Mutex
	jobs    map[string]*job.Job
	workers map[string]*WorkerInfo
}

// WorkerInfo tracks a registered worker. CallbackURL is only used in push
// mode (worker runs its own tiny HTTP server the controller can POST to).
type WorkerInfo struct {
	ID          string
	Labels      map[string]string
	CallbackURL string
	LastSeen    time.Time
}

func NewStore() *Store {
	return &Store{
		jobs:    make(map[string]*job.Job),
		workers: make(map[string]*WorkerInfo),
	}
}

func (s *Store) AddJob(spec job.Spec) *job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	j := &job.Job{
		ID:        newJobID(),
		Spec:      spec,
		Status:    job.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.jobs[j.ID] = j
	return j
}

func (s *Store) GetJob(id string) (*job.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	return j, ok
}

func (s *Store) ListJobs() []*job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*job.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, j)
	}
	return out
}

// ClaimNextPending is used by pull-mode polling: it atomically finds the
// first pending job eligible for workerID (matching NodeTarget, or no
// target set) and marks it assigned so no other poll can also claim it.
func (s *Store) ClaimNextPending(workerID string) *job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range s.jobs {
		if j.Status != job.StatusPending {
			continue
		}
		if j.Spec.NodeTarget != "" && j.Spec.NodeTarget != workerID {
			continue
		}
		j.Status = job.StatusAssigned
		j.AssignedTo = workerID
		j.Attempts++
		j.UpdatedAt = time.Now()
		return j
	}
	return nil
}

// ApplyStatus updates a job from a worker's status report. If the job
// failed and retries remain, it is reset to pending so it can be
// re-claimed (by pull) or re-pushed (by push). Reports for a job that is
// no longer in-flight (already succeeded, or reassigned elsewhere) are
// ignored rather than overwriting a result — this is the duplicate-
// execution guard after a worker reconnects late.
func (s *Store) ApplyStatus(r job.StatusReport) (*job.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[r.JobID]
	if !ok {
		return nil, false
	}
	// Stale report from a worker that no longer owns this job attempt.
	if j.Status == job.StatusSucceeded {
		return j, true
	}
	if j.AssignedTo != "" && j.AssignedTo != r.WorkerID && r.Status != job.StatusFailed {
		return j, true
	}

	j.UpdatedAt = time.Now()
	switch r.Status {
	case job.StatusRunning:
		j.Status = job.StatusRunning
		j.AssignedTo = r.WorkerID
	case job.StatusSucceeded:
		j.Status = job.StatusSucceeded
		j.Result = r.Result
	case job.StatusFailed:
		j.Result = r.Result
		if j.Attempts < j.Spec.MaxRetries {
			// retry: make it eligible to be claimed/pushed again
			j.Status = job.StatusPending
			j.AssignedTo = ""
		} else {
			j.Status = job.StatusFailed
		}
	}
	return j, true
}

func (s *Store) RegisterWorker(w WorkerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.LastSeen = time.Now()
	s.workers[w.ID] = &w
}

func (s *Store) ListWorkers() []*WorkerInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*WorkerInfo, 0, len(s.workers))
	for _, w := range s.workers {
		out = append(out, w)
	}
	return out
}

// PendingJobsForPush returns pending jobs eligible for push delivery
// (used by the push-mode scheduler loop).
func (s *Store) PendingJobsForPush() []*job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*job.Job{}
	for _, j := range s.jobs {
		if j.Status == job.StatusPending {
			out = append(out, j)
		}
	}
	return out
}

func (s *Store) MarkAssigned(jobID, workerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[jobID]; ok && j.Status == job.StatusPending {
		j.Status = job.StatusAssigned
		j.AssignedTo = workerID
		j.Attempts++
		j.UpdatedAt = time.Now()
	}
}
