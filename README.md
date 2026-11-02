# Lightweight Container-Based Task Orchestration for CCCLEO

A minimal Go controller + worker agent for running measurement/experiment
jobs as Docker containers across heterogeneous, intermittently-connected
nodes (CCCLEO: satellite links, 100-200ms latency, low bandwidth,
frequent disconnection). Built as the course project for CSC 579.

## Layout

```
pkg/job/        shared job schema (Spec, Job, StatusReport, RegisterRequest)
controller/     REST API + in-memory job store + pull/push scheduler
worker/         polls (pull) or receives pushes (push), runs jobs via docker
testimg/        tiny `FROM scratch` Go binary used to test orchestration
                without needing a container registry
```

## Why pull by default

Research question: does pull-based scheduling (workers ask for work when
free) hold up better than push (controller sends work to workers) when
links are slow, lossy, or intermittently down? Pull is implemented as the
primary path; push is a working baseline so the two can be measured
against each other with `tc`/`netem`.

## Running it

**1. Build the test image** (no registry needed — avoids relying on
network access to Docker Hub during the demo):

```bash
cd testimg
CGO_ENABLED=0 GOOS=linux go build -o testjob main.go
docker build -t cccleo-testjob:local .
```

Swap in a real measurement image later by changing `-d '{"image": ...}'`
below to any image already pulled/cached on the demo machine.

**2. Start the controller** (pick a mode):

```bash
cd controller
go run . -mode=pull     # or: go run . -mode=push
```

**3. Start one or more workers:**

```bash
cd worker
# pull mode:
go run . -id=w1 -controller=http://localhost:8080 -mode=pull

# push mode (worker opens its own small HTTP server for the controller to call):
go run . -id=w1 -controller=http://localhost:8080 -mode=push -addr=:9001
```

**4. Submit a job:**

```bash
curl -X POST http://localhost:8080/jobs \
  -d '{"image":"cccleo-testjob:local","max_retries":2}'
```

**5. Check status:**

```bash
curl http://localhost:8080/jobs            # list all
curl http://localhost:8080/jobs/<job-id>    # one job
```

## REST API

| Method | Path                  | Purpose                                   |
|--------|-----------------------|--------------------------------------------|
| POST   | /jobs                 | submit a job (Spec JSON)                   |
| GET    | /jobs                 | list all jobs                              |
| GET    | /jobs/{id}            | get one job                                |
| POST   | /jobs/{id}/status     | worker reports running/succeeded/failed    |
| POST   | /workers/register     | worker announces itself (+ callback_url for push) |
| GET    | /workers              | list registered workers                    |
| GET    | /poll?worker_id=X     | pull-mode: worker asks for next job        |

Job spec fields: `image`, `command` (optional override), `env` (optional),
`node_target` (optional — pin to one worker), `max_retries`.

## How duplicate execution is avoided

Every job gets a random unique ID at submission. A job only moves out of
`pending` once a worker claims it (pull) or the controller marks it
assigned (push); a late/duplicate status report for a job that already
succeeded, or that has since been reassigned to a different worker, is
ignored rather than overwriting the result (`Store.ApplyStatus` in
`controller/store.go`).

## Simulating CCCLEO network conditions (next step)

Not wired up yet — planned for the Nov 17-23 week per the proposal
timeline. The idea: run workers in separate network namespaces or VMs and
apply `tc qdisc add ... netem delay 150ms loss 2% rate 500kbit` to a
worker's interface, then compare completion rate / scheduling latency /
reconnect recovery between `-mode=pull` and `-mode=push` runs.

## What's stubbed / not built yet (be upfront about this in the demo)

- **Azure IoT Hub integration** — not wired in; worker registration today
  only talks to the controller directly. Per the proposal's risk section,
  if Azure access is awkward to set up in time, the local REST registration
  path (what's built now) is the fallback and is already in place.
- **Power BI dashboard** — optional per the proposal; `GET /jobs` and
  `GET /workers` return the JSON a dashboard would consume, but nothing
  renders it yet.
- **K3s comparison** — not started.
- **tc/netem experiment harness** — not started (see above).

## Suggested midterm demo script

1. Show the code layout and explain the job schema (30s).
2. Start controller in pull mode + 2 workers.
3. `curl -X POST /jobs` a couple of jobs, show them get distributed and
   `GET /jobs` go to `succeeded`.
4. Kill one worker mid-poll-loop, show the other worker(s) keep picking up
   new jobs — ties back to "pull survives a dead/disconnected node without
   the controller needing to notice."
5. Restart controller in push mode, repeat, and point out the difference:
   in push mode, if a worker's callback URL is unreachable the controller
   just logs and retries next tick (visible in its log output) — a preview
   of what the tc/netem experiments will measure properly.
