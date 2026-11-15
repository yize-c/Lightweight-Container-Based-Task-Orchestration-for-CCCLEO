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

- Prototype tested end to end with screenshots: jobs submitted through
  the REST API, run as containers, results reported back
- Simplified the code so every part is easy to explain
- [Update 2 (PDF)](files/update2.pdf)

## Update 1 — Oct 22, 2026

- Project plan in my own words and first progress
- [Update 1 (PDF)](files/update1.pdf)
