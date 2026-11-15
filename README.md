# Lightweight Container-Based Task Orchestration for CCCLEO

**A course project for CSC 579 (Fall 2026) @ UVic**
by Yi-Ze Chen

**Home** · [Logbook](docs/logbook.md) · [Project Proposal](docs/proposal.md) · [Project Updates](docs/updates.md) · [Demo & Presentation](docs/demo.md) · [Project Report](docs/report.md) · [Code](docs/code.md)

*Last updated: November 14, 2026*

---

## Project topic

The [CCCLEO testbed](https://onlineacademiccommunity.uvic.ca/cccleo/)
links computers across Canada over a low-Earth-orbit satellite network.
Its nodes range from VMs to small ARM devices, and its links can be slow,
lossy, or briefly cut off. Full orchestrators such as Kubernetes or K3s
are heavy for a testbed like this.

This project builds a small Go controller and worker agent that run
tasks as Docker containers, and compares two ways of handing out work:

- **Push**: the controller sends each job to a worker.
- **Pull**: each worker asks the controller for work every few seconds.

## Research question

How do push and pull scheduling compare in **latency** and **task
completion** when the network gets worse (added delay, packet loss,
temporary outages)?

## Approach

Two methods, checked against each other:

1. **Emulation + measurement**: controller and worker run in separate
   containers; Linux `tc/netem` degrades the worker's link; the same
   workload runs under 10 network profiles in both modes.
2. **Analytical model**: simple formulas predict the latency of each
   mode; the predictions are compared with the measurements.

## Project info

| | |
|---|---|
| Student | Yi-Ze Chen |
| Course | CSC 579, Fall 2026, University of Victoria |
| Current phase | Experiments and analysis (Nov 10 – 23) |
| Final presentation | December 4, 2026 |
| Final report | December 14, 2026 |
| Source code | this repository ([how to run](docs/code.md)) |

## Latest news

- **Nov 14** — Network emulation experiments done; the model predicted
  where pull catches up with push (~1 s of extra round-trip time), and
  the experiment confirmed it. See [Update 3](docs/updates.md#update-3--nov-16-2026).
- **Nov 2** — Controller/worker prototype working end to end.
- **Oct 9** — [Project proposal](docs/proposal.md) submitted.
