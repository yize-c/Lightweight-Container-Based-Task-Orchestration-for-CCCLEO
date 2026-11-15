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

Go, Docker, REST/HTTP, Linux `tc/netem`, K3s (baseline), Azure IoT Hub
(device management), Jira (Scrum, two-week sprints), Python/Power BI
(charts, optional).

## Timeline

| Period | Milestone |
|---|---|
| Oct 13–19 | Job schema, REST API, metrics, test VMs |
| Oct 20–26 | Controller, job store, worker registration |
| Oct 27–Nov 9 | Worker agent, push/pull scheduling, midterm demo |
| Nov 10–16 | Azure IoT Hub, reconnection recovery |
| Nov 17–23 | Network emulation and image caching experiments |
| Nov 24–end | Analysis, presentation, report, website |

**Priority if time is short:** the working prototype and push/pull
experiments come first; hybrid scheduling, a web frontend and Power BI
are optional.
