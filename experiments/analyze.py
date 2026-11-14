"""Compare measured results (results/*.json from run.sh) with the simple
analytical model. Standard library only.  Usage:  python analyze.py

Model (all times in ms; D = extra round-trip time added by netem,
T = worker poll interval, K = push tick, R = fixed cost measured on the
clean "base" link: docker run + HTTP overheads):

  scheduling delay (submit -> job leaves "pending")
    push:  K/2            controller hands it out on the next tick;
                          the network is NOT in this number (see README)
    pull:  (T + D)/2      wait for the next poll; one poll cycle = T + D

  total latency (submit -> result reported)
    push:  K/2       + R_push + D    (status report crosses the link)
    pull:  (T + D)/2 + R_pull + D

  outage (link down X s in every Y s)
    push completion ~ 1 - max(0, X - 5)/Y
                      a push fails only if the link stays down longer than
                      the 5 s push timeout; a failed push is never re-sent,
                      so that job stays "running" forever
    pull completion ~ 1           the worker just polls again later
    pull extra scheduling delay ~ (X/Y) * (X/2)  (ignores TCP back-off)

  packet loss: no simple formula -> measured only.
"""
import csv
import glob
import json
import os
import re
import statistics
from collections import defaultdict
from datetime import datetime

HERE = os.path.dirname(os.path.abspath(__file__))
RES = os.path.join(HERE, "results")
PUSH_TIMEOUT_S = 5  # http.Client timeout in pushJobToWorker (controller/main.go)


def parse_time(s):
    # Go writes up to 9 fractional digits; Python wants at most 6.
    s = re.sub(r"(\.\d{6})\d+", r"\1", s).replace("Z", "+00:00")
    return datetime.fromisoformat(s)


def ms(a, b):
    return (parse_time(b) - parse_time(a)).total_seconds() * 1000


def extra_rtt_ms(netem):
    m = re.search(r"delay (\d+)ms", netem)
    return int(m.group(1)) if m else 0


# ---- load every run ----
runs = defaultdict(list)  # (profile, mode) -> list of per-run dicts
for meta_path in sorted(glob.glob(os.path.join(RES, "*.meta.json"))):
    meta = json.load(open(meta_path))
    jobs = json.load(open(meta_path.replace(".meta.json", ".jobs.json"))) or []
    sched = [ms(j["created_at"], j["assigned_at"]) for j in jobs if j.get("assigned_at")]
    total = [ms(j["created_at"], j["finished_at"]) for j in jobs
             if j.get("finished_at") and j["status"] == "succeeded"]
    done = sum(1 for j in jobs if j["status"] == "succeeded")
    runs[(meta["profile"], meta["mode"])].append({
        "meta": meta, "n": len(jobs), "done": done, "sched": sched, "total": total})

if not runs:
    raise SystemExit("No results found. Run ./run.sh first.")


def pooled(key, field):
    vals = [v for r in runs.get(key, []) for v in r[field]]
    return statistics.mean(vals) if vals else None


def median(key, field):
    vals = [v for r in runs.get(key, []) for v in r[field]]
    return statistics.median(vals) if vals else None


# R = fixed cost per mode, measured on the clean link
R = {}
for mode in ("push", "pull"):
    s, t = pooled(("base", mode), "sched"), pooled(("base", mode), "total")
    R[mode] = (t - s) if s is not None and t is not None else None


def model(profile, mode, meta):
    T, K = meta["poll_ms"], meta["tick_ms"]
    D = extra_rtt_ms(meta["netem"])
    if profile.startswith("l"):           # packet loss: no model
        return None, None, None
    sched = K / 2 if mode == "push" else (T + D) / 2
    completion = 1.0
    if profile == "out":
        X, Y = meta["out_down_s"], meta["out_cycle_s"]
        if mode == "push":
            completion = 1 - max(0, X - PUSH_TIMEOUT_S) / Y
        else:
            sched += (X / Y) * (X / 2) * 1000
    total = sched + R[mode] + D if R.get(mode) is not None else None
    return sched, total, completion


def fmt(x, pct=False):
    if x is None:
        return "-"
    return f"{x:.0%}" if pct else f"{x:.0f}"


def err(measured, predicted):
    if measured is None or predicted is None or predicted == 0:
        return "-"
    return f"{(measured - predicted) / predicted:+.0%}"


order = ["base", "d50", "d200", "d600", "l1", "l5", "l10", "out"]
keys = sorted(runs, key=lambda k: (order.index(k[0]) if k[0] in order else 99, k[1]))

header = ["profile", "mode", "runs", "completion", "completion_model",
          "sched_mean_ms", "sched_median_ms", "sched_model_ms", "sched_err",
          "total_mean_ms", "total_median_ms", "total_model_ms", "total_err"]
rows = []
for key in keys:
    profile, mode = key
    rs = runs[key]
    completion = sum(r["done"] for r in rs) / max(1, sum(r["n"] for r in rs))
    m_sched, m_total, m_comp = model(profile, mode, rs[0]["meta"])
    s_mean, t_mean = pooled(key, "sched"), pooled(key, "total")
    rows.append([profile, mode, len(rs), fmt(completion, True), fmt(m_comp, True),
                 fmt(s_mean), fmt(median(key, "sched")), fmt(m_sched), err(s_mean, m_sched),
                 fmt(t_mean), fmt(median(key, "total")), fmt(m_total), err(t_mean, m_total)])

with open(os.path.join(RES, "summary.csv"), "w", newline="") as f:
    csv.writer(f).writerows([header] + rows)

# Markdown table: paste straight into the report / logbook
print("R (fixed cost from base runs, ms):",
      {k: fmt(v) for k, v in R.items()})
print()
print("| " + " | ".join(header) + " |")
print("|" + "---|" * len(header))
for r in rows:
    print("| " + " | ".join(str(c) for c in r) + " |")
print(f"\nSaved {os.path.join('results', 'summary.csv')}")
