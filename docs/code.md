# Code

[Home](../README.md) · [Logbook](logbook.md) · [Project Proposal](proposal.md) · [Project Updates](updates.md) · [Demo & Presentation](demo.md) · [Project Report](report.md) · **Code**

---

Go controller + worker agent that run jobs as Docker containers.

## Files

```
pkg/job/        job schema (shared types)
controller/     REST API + job store + pull/push scheduling
worker/         runs jobs via docker, pull mode or push mode
testimg/        tiny test image, no registry needed
```

## Run it

```bash
# 1. build the test image (once)
cd testimg
CGO_ENABLED=0 GOOS=linux go build -o testjob main.go
docker build -t cccleo-testjob:local .

# 2. start the controller
cd ../controller
go run . -mode=pull     # or -mode=push

# 3. start a worker
cd ../worker
go run . -id=w1 -controller=http://localhost:8080 -mode=pull

# 4. submit a job
curl -X POST http://localhost:8080/jobs -d '{"image":"cccleo-testjob:local","max_retries":2}'

# 5. check it
curl http://localhost:8080/jobs
```

## API

| Method | Path | Does |
|---|---|---|
| POST | /jobs | submit a job |
| GET | /jobs | list jobs |
| GET | /jobs/{id} | one job |
| POST | /jobs/{id}/status | worker reports result |
| POST | /workers/register | worker announces itself |
| GET | /poll?worker_id=X | pull mode: worker asks for work |
| GET | /metrics | numbers below |

Job fields: `image`, `command`, `env`, `node_target`, `max_retries`.

## Metrics (`GET /metrics`)

- `completion_rate` — succeeded / total
- `avg_scheduling_delay_ms` — submit → job leaves `pending`. (In push mode this is set before the job is sent, so it does not include the network.)
- `avg_total_latency_ms` — submit → result reported. **Use this one to compare push and pull.**
- `stale_reports_ignored` — duplicate-execution guard caught something

## How it avoids running a job twice

Each job has a unique ID and one status. Once a worker has it
(`running`), a report from any other worker for that job is ignored.
See `ApplyStatus` in [`controller/store.go`](../controller/store.go).

## Not built yet

Azure IoT Hub, image caching experiment, K3s comparison, Power BI dashboard.

## Experiments

Push vs pull under emulated network conditions (tc/netem): see
[experiments/README.md](../experiments/README.md).
