# Project Updates

[Home](../README.md) · [Logbook](logbook.md) · [Project Proposal](proposal.md) · **Project Updates** · [Demo & Presentation](demo.md) · [Project Report](report.md) · [Code](code.md)

---

## Update 3 — Nov 16, 2026

**Done since Update 2**
- Experiment setup: controller and worker each run in their own Docker
  container; `tc/netem` changes the worker's link; one script runs every
  combination and saves the raw data ([experiments/](../experiments/))
- 10 network profiles × 2 modes, 30 jobs each
- Analytical model of latency for each mode, compared with measurements

**Main results** (total latency = submit → result reported, ms)

| Extra RTT | Pull measured | Pull model | Push measured | Push model |
|---|---|---|---|---|
| 0 ms | 2006 | 1996 | 1436 | 1522 |
| 200 ms | 2406 | 2296 | 1932 | 1922 |
| 600 ms | 2853 | 2896 | 2759 | 2722 |
| 1000 ms | 3495 | 3496 | 3527 | 3522 |
| 1500 ms* | 4355 | 4246 | 4581 | 4522 |

\*submitted every 8 s instead of 4 s so the worker isn't overloaded

- Push is faster on a good link, but its latency grows ~2× faster with
  delay; the model predicted they meet at ~1 s extra RTT, and they did
- First model was off by 30% for push: each push crosses the slow link
  one extra time; after correcting it, all delay profiles are within 6%
- With a job every ~4.5 s and 1500 ms delay, pull fell behind (13 s
  average wait): each worker runs one job at a time, so jobs queue up
- **Outages** (10 s every 40 s): pull completed 100%, push 97% (a failed
  push is never re-sent); **packet loss** up to 10% changed little

**Next:** repeat runs 3×, decide on remaining proposal items (Azure IoT
Hub, image caching, K3s), start the report.

## Update 2 — Nov 5, 2026

- Prototype tested end to end with screenshots: jobs submitted through
  the REST API, run as containers, results reported back
- Simplified the code so every part is easy to explain
- [Update 2 (PDF)](files/update2.pdf)

## Update 1 — Oct 22, 2026

- Project plan in my own words and first progress
- [Update 1 (PDF)](files/update1.pdf)
