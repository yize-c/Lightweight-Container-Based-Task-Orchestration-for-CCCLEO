# Push vs pull experiments (tc/netem emulation)

Controller and worker each run in their own container, so the worker has
its own network link. `tc netem` degrades that link; `run.sh` repeats the
same workload under each network profile in both modes and saves the raw
job data; `analyze.py` compares the measurements with a simple
analytical model (cross-validation).

## Run

Needs Docker Desktop (running) and Git Bash. From the repo root:

```bash
# 1. test image (once)
cd testimg && CGO_ENABLED=0 GOOS=linux go build -o testjob main.go \
  && docker build -t cccleo-testjob:local . && cd ..

# 2. all profiles x both modes (~40 min)
cd experiments && ./run.sh

# 3. results table (also saved as results/summary.csv)
python analyze.py
```

Quick check first: `JOBS=5 PROFILES="base d200" ./run.sh`.
Final numbers: `REPS=3 ./run.sh` (3 repetitions, ~2 h).

## Network profiles (applied to the worker's link)

| profile | netem | meaning |
|---|---|---|
| base | delay 0ms | clean link (also measures fixed cost R) |
| d50 / d200 / d600 | delay 50/200/600ms | extra round-trip time (LEO-like to GEO-like) |
| d1000 / d1500 | delay 1000/1500ms | very high delay; run separately: `REPS=3 PROFILES="d1000" ./run.sh`, `GAP=8 REPS=3 PROFILES="d1500" ./run.sh` |
| l1 / l5 / l10 | loss 1/5/10% | packet loss on the worker's outgoing traffic |
| out | loss 100% + iptables DROP, 10 s every 40 s | intermittent link, both directions |

Workload: 30 jobs per run, one every 4-5 s (4 s + random 0-1 s). The gap
is longer than a job takes, so jobs don't queue behind each other, and
the random part keeps jobs from always landing at the same point in the
poll cycle / push tick.

## What the metrics mean here

- **Scheduling delay** = submit -> job leaves `pending`. In push mode the
  controller sets this *before* sending the job over the network, so it
  does not include the link. Compare the two modes on **total latency**
  and **completion rate**; use scheduling delay to check the model.
- **Completion rate** < 100% in push mode under outages is expected: a
  push that times out is not re-sent, so that job stays `running`.

The model's formulas are at the top of `analyze.py`.
