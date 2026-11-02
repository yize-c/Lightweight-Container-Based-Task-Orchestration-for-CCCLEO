// Controller: accepts job submissions over REST, hands jobs out to
// worker agents (pull or push), and collects status reports.
//
// Run modes (research question: push vs pull under latency/disconnection):
//
//	go run . -mode=pull   (default)
//	go run . -mode=push
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cccleo-orchestrator/pkg/job"
)

var store = NewStore()
var mode = "pull"

func main() {
	addr := ":8080"
	if v := os.Getenv("MODE"); v != "" {
		mode = v
	}
	for _, a := range os.Args[1:] {
		if a == "-mode=push" {
			mode = "push"
		}
		if a == "-mode=pull" {
			mode = "pull"
		}
	}
	log.Printf("controller starting in %s mode on %s", mode, addr)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", handleSubmitJob)
	mux.HandleFunc("GET /jobs", handleListJobs)
	mux.HandleFunc("GET /jobs/{id}", handleGetJob)
	mux.HandleFunc("POST /jobs/{id}/status", handleJobStatus)
	mux.HandleFunc("POST /workers/register", handleRegisterWorker)
	mux.HandleFunc("GET /workers", handleListWorkers)
	mux.HandleFunc("GET /poll", handlePoll) // pull mode: worker asks "got anything for me?"
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	if mode == "push" {
		go pushSchedulerLoop()
	}

	log.Fatal(http.ListenAndServe(addr, mux))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	var spec job.Spec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "bad job spec: "+err.Error(), http.StatusBadRequest)
		return
	}
	if spec.Image == "" {
		http.Error(w, "spec.image is required", http.StatusBadRequest)
		return
	}
	if spec.MaxRetries == 0 {
		spec.MaxRetries = 1
	}
	j := store.AddJob(spec)
	log.Printf("submitted %s image=%s target=%q", j.ID, spec.Image, spec.NodeTarget)
	writeJSON(w, http.StatusCreated, j)
}

func handleListJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, store.ListJobs())
}

func handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	j, ok := store.GetJob(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, j)
}

func handleJobStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var rep job.StatusReport
	if err := json.NewDecoder(r.Body).Decode(&rep); err != nil {
		http.Error(w, "bad status report: "+err.Error(), http.StatusBadRequest)
		return
	}
	rep.JobID = id
	j, ok := store.ApplyStatus(rep)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	log.Printf("status %s -> %s (worker=%s attempts=%d)", id, rep.Status, rep.WorkerID, j.Attempts)
	writeJSON(w, http.StatusOK, j)
}

func handleRegisterWorker(w http.ResponseWriter, r *http.Request) {
	var req job.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad register request: "+err.Error(), http.StatusBadRequest)
		return
	}
	callback := r.URL.Query().Get("callback_url")
	store.RegisterWorker(WorkerInfo{ID: req.WorkerID, Labels: req.Labels, CallbackURL: callback})
	log.Printf("registered worker=%s callback=%q labels=%v", req.WorkerID, callback, req.Labels)
	writeJSON(w, http.StatusOK, map[string]string{"status": "registered", "mode": mode})
}

func handleListWorkers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, store.ListWorkers())
}

// handlePoll implements pull-mode scheduling: a worker calls this
// whenever it is free. Returns 204 if nothing is available.
func handlePoll(w http.ResponseWriter, r *http.Request) {
	workerID := r.URL.Query().Get("worker_id")
	if workerID == "" {
		http.Error(w, "worker_id query param required", http.StatusBadRequest)
		return
	}
	j := store.ClaimNextPending(workerID)
	if j == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Printf("poll: handed %s to worker=%s (attempt %d)", j.ID, workerID, j.Attempts)
	writeJSON(w, http.StatusOK, j)
}

// pushSchedulerLoop implements push-mode scheduling: the controller
// actively ticks and shoves pending jobs at registered workers'
// callback URLs. This is the baseline being compared against pull.
func pushSchedulerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		pending := store.PendingJobsForPush()
		if len(pending) == 0 {
			continue
		}
		workers := store.ListWorkers()
		if len(workers) == 0 {
			continue
		}
		wi := 0
		for _, j := range pending {
			w := pickWorker(workers, j.Spec.NodeTarget, &wi)
			if w == nil || w.CallbackURL == "" {
				continue
			}
			store.MarkAssigned(j.ID, w.ID)
			go pushJobToWorker(w.CallbackURL, j)
		}
	}
}

func pickWorker(workers []*WorkerInfo, target string, rr *int) *WorkerInfo {
	if target != "" {
		for _, w := range workers {
			if w.ID == target {
				return w
			}
		}
		return nil
	}
	*rr = (*rr + 1) % len(workers)
	return workers[*rr]
}

func pushJobToWorker(callbackURL string, j *job.Job) {
	body, _ := json.Marshal(j)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(callbackURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("push: failed to reach worker at %s for job %s: %v (will retry next tick)", callbackURL, j.ID, err)
		// leave job assigned; a real implementation would time out and
		// requeue — left as future work per the proposal's risk section.
		return
	}
	defer resp.Body.Close()
}
