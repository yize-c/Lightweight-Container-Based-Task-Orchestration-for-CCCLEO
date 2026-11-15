# Project Report

[Home](../README.md) · [Logbook](logbook.md) · [Project Proposal](proposal.md) · [Project Updates](updates.md) · [Demo & Presentation](demo.md) · **Project Report** · [Code](code.md)

---

*Due December 14, 2026. The final report (PDF) will be posted here.*

## Planned outline

1. Introduction: CCCLEO and why a lightweight orchestrator
2. Related work: Kubernetes/K3s, lightweight edge orchestration, push vs
   pull task distribution
3. System design: controller, worker, job states, duplicate guard
4. Methodology: network emulation with `tc/netem`, workload, metrics,
   analytical model
5. Results: delay, packet loss, outages; model vs measurement
6. Discussion: when to use push or pull, limits of this study
7. Conclusion and future work
8. Appendix: contributions statement

## Results so far

Raw data and the summary table are in
[experiments/results/](../experiments/results/). Main findings are in
[Midterm Update](updates.md#update-3--nov-16-2026).
