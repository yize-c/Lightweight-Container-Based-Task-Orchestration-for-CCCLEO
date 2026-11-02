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
	mu           sync.Mutex
	jobs         map[string]*job.Job
	workers      map[string]*WorkerInfo
	staleReports int // duplicate-execution guard hit count, for metrics
}

// WorkerInfo tracks a registered worker. CallbackURL is only used in push
// mode (worker runs its own tiny HTTP server the controller can POST to).
type WorkerInfo struct {
	ID          string
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
// target set) and marks it running so no other poll can also claim it.
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
		j.Status = job.StatusRunning
		j.AssignedTo = workerID
		j.Attempts++
		now := time.Now()
		j.UpdatedAt = now
		if j.AssignedAt == nil {
			j.AssignedAt = &now
		}
		return j
	}
	return nil
}

// ApplyStatus updates a job from a worker's final status report
// (succeeded or failed — there's no separate "started running" report
// to handle). If it failed and retries remain, it's reset to pending so
// it can be re-claimed (by pull) or re-pushed (by push). A report for a
// job that is no longer in-flight (already succeeded, or reassigned to
// a different worker) is ignored rather than overwriting the result —
// this is the duplicate-execution guard after a worker reconnects late.
func (s *Store) ApplyStatus(r job.StatusReport) (*job.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[r.JobID]
	if !ok {
		return nil, false
	}
	if j.Status == job.StatusSucceeded {
		s.staleReports++
		return j, true
	}
	if j.AssignedTo != "" && j.AssignedTo != r.WorkerID {
		s.staleReports++
		return j, true
	}

	now := time.Now()
	j.UpdatedAt = now
	switch r.Status {
	case job.StatusSucceeded:
		j.Status = job.StatusSucceeded
		j.Result = r.Result
		j.FinishedAt = &now
	case job.StatusFailed:
		j.Result = r.Result
		if j.Attempts < j.Spec.MaxRetries {
			// retry: make it eligible to be claimed/pushed again
			j.Status = job.StatusPending
			j.AssignedTo = ""
		} else {
			j.Status = job.StatusFailed
			j.FinishedAt = &now
		}
	}
	return j, true
}

// Metrics computes the aggregate stats the proposal's research questions
// need, straight from the job list — no separate time-series store.
func (s *Store) Metrics() job.Metrics {
	s.mu.Lock()
	defer s.mu.Unlock()

	var m job.Metrics
	m.TotalJobs = len(s.jobs)
	m.StaleReportsIgnored = s.staleReports

	var schedSum, totalSum float64
	var schedN, totalN int

	for _, j := range s.jobs {
		switch j.Status {
		case job.StatusPending:
			m.Pending++
		case job.StatusRunning:
			m.Running++
		case job.StatusSucceeded:
			m.Succeeded++
		case job.StatusFailed:
			m.Failed++
		}
		if j.AssignedAt != nil {
			schedSum += j.AssignedAt.Sub(j.CreatedAt).Seconds() * 1000
			schedN++
		}
		if j.FinishedAt != nil {
			totalSum += j.FinishedAt.Sub(j.CreatedAt).Seconds() * 1000
			totalN++
		}
	}
	if m.TotalJobs > 0 {
		m.CompletionRate = float64(m.Succeeded) / float64(m.TotalJobs)
	}
	if schedN > 0 {
		m.AvgSchedulingDelayMs = schedSum / float64(schedN)
	}
	if totalN > 0 {
		m.AvgTotalLatencyMs = totalSum / float64(totalN)
	}
	return m
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
		j.Status = job.StatusRunning
		j.AssignedTo = workerID
		j.Attempts++
		now := time.Now()
		j.UpdatedAt = now
		if j.AssignedAt == nil {
			j.AssignedAt = &now
		}
	}
}
