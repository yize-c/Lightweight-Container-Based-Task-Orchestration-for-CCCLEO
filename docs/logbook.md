# Logbook

[Home](../README.md) · **Logbook** · [Project Proposal](proposal.md) · [Project Updates](updates.md) · [Demo & Presentation](demo.md) · [Project Report](report.md) · [Code](code.md)

---

**Week of Sep 21, 2026**
- Chose instructor-offered project idea #2: lightweight container-based task orchestration for the CCCLEO testbed
- Submitted the project topic (Sep 21)

**Week of Oct 5, 2026**
- Wrote and submitted the project proposal (Oct 9): research questions, architecture, tools, timeline

**Week of Oct 19, 2026**
- Created the GitHub repository and posted the proposal (Oct 21)
- Submitted Project Update 1 (Oct 22)

**Week of Oct 26, 2026**
- Built the first controller/worker prototype in Go: REST API, job store, pull and push scheduling, retries, `/metrics` (Nov 1)
- Tested the full flow locally: submit job → worker runs container → result reported

**Week of Nov 2, 2026**
- Simplified the code: removed CPU/memory sampling and worker labels, merged the "assigned" state into "running"
- First informal push vs pull test: push 406 ms vs pull 4572 ms scheduling delay (later found to be mostly queueing, not network)
- Submitted Project Update 2 (Nov 5)

**Week of Nov 9, 2026**
- Built the experiment setup: controller and worker in separate Docker containers, `tc/netem` on the worker's link, scripts to run and analyze
- Ran 10 network profiles (delay, loss, outages) in both modes
- Wrote an analytical model, compared it with the measurements, and corrected it (push crosses the slow link one extra time)
- Model predicted pull catches up with push at ~1 s extra RTT; confirmed at 1000 and 1500 ms
- Found that pull overloads first at very high delay because each worker runs one job at a time

**Next**
- Repeat every run 3 times for final numbers
- Project Update 3 (Nov 16)
- Final presentation video (Dec 4), report (Dec 14), website final (Dec 21)
