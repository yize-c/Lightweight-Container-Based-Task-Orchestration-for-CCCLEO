# Lightweight Container-Based Task Orchestration for CCCLEO

Go controller + worker agent that run jobs as Docker containers. CSC 579
course project.

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
- `avg_scheduling_delay_ms` — submit → worker picks it up. **This is the pull vs push number.**
- `avg_total_latency_ms` — submit → finished
- `stale_reports_ignored` — duplicate-execution guard caught something

## How it avoids running a job twice

Each job has a unique ID and one status. Once a worker has it
(`running`), a report from any other worker for that job is ignored.
See `ApplyStatus` in `controller/store.go`.

## Not built yet

Azure IoT Hub, Power BI dashboard, tc/netem network simulation, K3s
comparison, project website.
