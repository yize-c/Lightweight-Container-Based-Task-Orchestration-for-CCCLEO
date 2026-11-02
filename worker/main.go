// Worker agent: registers with the controller, then either polls for
// work (pull mode) or runs a tiny HTTP server the controller can push
// jobs to (push mode). Either way, it executes the job as a Docker
// container and reports status back.
//
//	go run . -id=w1 -controller=http://localhost:8080 -mode=pull
//	go run . -id=w2 -controller=http://localhost:8080 -mode=push -addr=:9001
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"cccleo-orchestrator/pkg/job"
)

func main() {
	id := flag.String("id", "worker-1", "worker ID")
	controllerURL := flag.String("controller", "http://localhost:8080", "controller base URL")
	mode := flag.String("mode", "pull", "pull | push")
	addr := flag.String("addr", ":9001", "address this worker listens on (push mode only)")
	pollInterval := flag.Duration("poll-interval", 2*time.Second, "how often to poll in pull mode")
	flag.Parse()

	w := &Worker{ID: *id, ControllerURL: *controllerURL}

	callback := ""
	if *mode == "push" {
		callback = fmt.Sprintf("http://localhost%s/run", *addr)
	}
	if err := w.register(callback); err != nil {
		log.Fatalf("register failed: %v", err)
	}
	log.Printf("worker %s registered with controller %s (mode=%s)", w.ID, w.ControllerURL, *mode)

	if *mode == "push" {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /run", func(wr http.ResponseWriter, r *http.Request) {
			var j job.Job
			if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
				http.Error(wr, err.Error(), http.StatusBadRequest)
				return
			}
			log.Printf("push: received job %s", j.ID)
			go w.runJob(&j)
			wr.WriteHeader(http.StatusAccepted)
		})
		log.Printf("worker %s listening for pushed jobs on %s", w.ID, *addr)
		log.Fatal(http.ListenAndServe(*addr, mux))
	} else {
		w.pollLoop(*pollInterval)
	}
}

type Worker struct {
	ID            string
	ControllerURL string
}

func (w *Worker) register(callbackURL string) error {
	body, _ := json.Marshal(job.RegisterRequest{WorkerID: w.ID, Labels: map[string]string{}})
	url := w.ControllerURL + "/workers/register"
	if callbackURL != "" {
		url += "?callback_url=" + callbackURL
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// pollLoop is the pull-mode heart of the worker: ask the controller for
// work whenever idle. This is the behavior the proposal's research
// question is testing against push — a worker that is briefly
// disconnected simply resumes polling once the link comes back, with no
// job "lost in flight" the way a push send to an unreachable worker can be.
func (w *Worker) pollLoop(interval time.Duration) {
	for {
		j, err := w.poll()
		if err != nil {
			log.Printf("poll error: %v (will retry)", err)
			time.Sleep(interval)
			continue
		}
		if j == nil {
			time.Sleep(interval)
			continue
		}
		w.runJob(j)
	}
}

func (w *Worker) poll() (*job.Job, error) {
	url := fmt.Sprintf("%s/poll?worker_id=%s", w.ControllerURL, w.ID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var j job.Job
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return nil, err
	}
	return &j, nil
}

// runJob executes the job as a Docker container and reports status back
// at each transition (running -> succeeded/failed).
func (w *Worker) runJob(j *job.Job) {
	log.Printf("running job %s (image=%s attempt=%d)", j.ID, j.Spec.Image, j.Attempts)
	w.reportStatus(j.ID, job.StatusRunning, "")

	args := []string{"run", "--rm"}
	for k, v := range j.Spec.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, j.Spec.Image)
	args = append(args, j.Spec.Command...)

	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if len(result) > 2000 {
		result = result[:2000] + "...(truncated)"
	}

	if err != nil {
		log.Printf("job %s failed: %v", j.ID, err)
		w.reportStatus(j.ID, job.StatusFailed, fmt.Sprintf("error: %v; output: %s", err, result))
		return
	}
	log.Printf("job %s succeeded", j.ID)
	w.reportStatus(j.ID, job.StatusSucceeded, result)
}

func (w *Worker) reportStatus(jobID, status, result string) {
	rep := job.StatusReport{JobID: jobID, WorkerID: w.ID, Status: status, Result: result}
	body, _ := json.Marshal(rep)
	url := fmt.Sprintf("%s/jobs/%s/status", w.ControllerURL, jobID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("failed to report status for %s: %v", jobID, err)
		return
	}
	defer resp.Body.Close()
}
