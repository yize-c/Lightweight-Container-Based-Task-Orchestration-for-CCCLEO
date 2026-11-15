# Project Updates

[Home](../README.md) · [Logbook](logbook.md) · [Project Proposal](proposal.md) · **Project Updates** · [Demo & Presentation](demo.md) · [Project Report](report.md) · [Code](code.md)

---

## Midterm Update — Nov 16, 2026

**Done since Update 2**
- Experiment setup: controller and worker each run in their own Docker
  container; `tc/netem` changes the worker's link; one script runs every
  combination and saves the raw data ([experiments/](../experiments/))
- 10 network profiles × 2 modes, 30 jobs per run, every run repeated 3 times
- Analytical model of latency for each mode, compared with measurements

**Main results** (total latency = submit → result reported, ms, mean of 3 runs)

| Extra RTT | Pull measured | Pull model | Push measured | Push model |
|---|---|---|---|---|
| 0 ms | 2172 | 2044 | 1552 | 1543 |
| 50 ms | 2177 | 2119 | 1668 | 1643 |
| 200 ms | 2312 | 2344 | 1994 | 1943 |
| 600 ms | 2915 | 2944 | 2796 | 2743 |
| 1000 ms | 3764 | 3544 | 3769 | 3543 |
| 1500 ms* | 4508 | 4294 | 4793 | 4543 |

\*submitted every 8 s instead of 4 s so the worker isn't overloaded

- Push is faster on a good link, but its latency grows faster with delay;
  the model predicted they meet at ~1 s extra RTT. Measured: tie at
  1000 ms (3764 vs 3769), pull faster at 1500 ms in all 3 runs
- First model was off by up to 30% for push: each push crosses the slow
  link one extra time; after correcting it, every delay profile is
  within 6% of the model
- With a job every ~4.5 s and 1500 ms delay, pull fell behind (13 s
  average wait): each worker runs one job at a time, so jobs queue up
- **Outages** (10 s every 40 s): pull completed 100% in all runs, push
  96% (4 of 90 jobs lost, because a failed push is never re-sent)
- **Packet loss** up to 10% changed little (at most ~230 ms)

**Next:** decide on remaining proposal items (Azure IoT Hub, image
caching, K3s), make charts, start the report.

## Update 2 — Nov 5, 2026

**Done since Update 1:** the worker agent, with pull mode (polls the controller) and push mode (receives jobs). One controller and one worker run on a Windows laptop with Docker Desktop; jobs use a small local test image.

- **Pull mode runs jobs:** a worker started after two jobs were pending picked up both and finished them; the controller shows them as succeeded.
- **Failed jobs are retried:** a job with a missing image and `max_retries: 2` was tried twice, then marked failed.
- **First pull vs push numbers** (5 jobs, 1 worker, no network problems):

| Metric | Pull | Push |
|---|---|---|
| Jobs completed | 5/5 | 5/5 |
| Avg scheduling delay | 2850 ms | 474 ms |
| Avg total latency | 4134 ms | 2344 ms |

Push was faster, but the test was not yet fair: the pull worker runs one job at a time, while push ran all five at once, and the network was perfect.

**Next:** make the modes comparable, add delay, loss and outages with `tc/netem`.

[Update 2 (PDF)](files/update2.pdf)

## Update 1 — Oct 22, 2026

**Goal:** a small controller/worker system that runs jobs as Docker containers, to test whether pulling tasks copes with bad links better than pushing them. This update covers the controller.

- **Job format:** five fields — `image`, `command`, `env`, `node_target`, `max_retries` — much smaller than a Kubernetes spec.
- **Controller in Go:** standard library only, jobs kept in memory, a REST API to submit jobs, register workers, poll for work, report results and read metrics.
- **Job life cycle:** each job has a unique ID and moves through pending, running, succeeded or failed; a failed job is retried until it runs out of retries.
- **No duplicate runs:** once a job is given to a worker, a result from any other worker is ignored and counted.
- **Two scheduling modes:** in pull mode the controller waits for workers to ask; in push mode it sends pending jobs every second.
- **Demo:** with no worker running, two submitted jobs stay pending, each with its own ID.

**Next:** the worker agent, first end-to-end tests, a metrics endpoint.

[Update 1 (PDF)](files/update1.pdf)
