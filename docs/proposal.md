# Project Proposal

[Home](../README.md) · [Logbook](logbook.md) · **Project Proposal** · [Project Updates](updates.md) · [Demo & Presentation](demo.md) · [Project Report](report.md) · [Code](code.md)

*Submitted October 9, 2026* · [Proposal (PDF)](files/proposal.pdf)

---

## Motivation

The CCCLEO testbed mixes VMs, mini PCs and small ARM devices connected
over satellite links with high latency, limited bandwidth and
intermittent connectivity. Kubernetes and K3s handle workloads well but
may be heavier than such a testbed needs. This project asks whether a
small, task-oriented orchestrator can run containers reliably under
these conditions.

## Research questions (as proposed)

1. **Push vs pull scheduling**: effect on task completion, scheduling
   latency, and recovery after disconnections
2. **Resource efficiency**: CPU and memory use compared with K3s
3. **Configuration simplicity**: can a small JSON job spec replace
   Kubernetes-style configuration?
4. **Image caching**: how much it helps over bandwidth-limited links

## Design

- **Controller**: REST API for jobs, job store, worker registry,
  push and pull scheduling, retries, metrics
- **Worker agent**: registers, gets jobs (polls or receives them), runs
  them with Docker, reports results
- Jobs have unique IDs so a job is never counted as run twice

## Planned tools

Go, Docker, REST/HTTP, Linux `tc/netem`, K3s (baseline), Power BI
(charts).

## Timeline (revised Nov 14, 2026)

| Period | Milestone | Status |
|---|---|---|
| Oct 13–26 | Job schema, REST API, controller, job store, worker registration | Done |
| Oct 27–Nov 9 | Worker agent, push/pull scheduling, end-to-end testing | Done |
| Nov 10–16 | Network emulation with `tc/netem`, push vs pull experiments (3 runs each), analytical model | Done |
| Nov 17–23 | Power BI charts of the results; image caching / K3s comparison if time allows | Planned |
| Nov 24–Dec 4 | Final presentation video (due Dec 4) | Planned |
| Dec 5–14 | Final report (due Dec 14) | Planned |
| Dec 15–21 | Website finalized (due Dec 21) | Planned |

**Changes from the original plan:** Azure IoT Hub and Jira were dropped to focus on the experiments; Docker containers replaced VMs for the emulated testbed; Power BI is now used for the result charts. Image caching and the K3s comparison are optional.

**Priority if time is short:** the working prototype and push/pull
experiments come first; hybrid scheduling, a web frontend and Power BI
are optional.
