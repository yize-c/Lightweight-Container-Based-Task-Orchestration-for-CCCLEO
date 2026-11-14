#!/usr/bin/env bash
# Runs every (mode x network profile) combination on the emulated testbed
# and saves the raw job data to results/. Then run: python analyze.py
#
#   cd experiments
#   ./run.sh                                  # everything
#   PROFILES="base d200" MODES=pull ./run.sh  # just a subset
#   REPS=3 ./run.sh                           # repeat each run 3 times
set -euo pipefail
export MSYS_NO_PATHCONV=1     # Git Bash: don't rewrite "/..." arguments
cd "$(dirname "$0")"

# ---- settings (override any of these as env vars) ----
JOBS=${JOBS:-30}          # jobs submitted per run
GAP=${GAP:-4}             # seconds between submissions (+ random 0-1s)
DRAIN=${DRAIN:-60}        # max seconds to wait for unfinished jobs at the end
REPS=${REPS:-1}           # repetitions of each run
MODES=${MODES:-"pull push"}
PROFILES=${PROFILES:-"base d50 d200 d600 l1 l5 l10 out"}
OUT_DOWN=${OUT_DOWN:-10}  # "out" profile: link down OUT_DOWN seconds...
OUT_CYCLE=${OUT_CYCLE:-40} # ...in every OUT_CYCLE seconds
POLL=2000                 # worker's default poll interval (ms), for analyze.py
TICK=1000                 # controller's push tick (ms), for analyze.py

API=http://localhost:8080
mkdir -p results

# netem settings for each profile, applied to worker1's link
netem_args() {
  case $1 in
    base) echo "delay 0ms" ;;
    d50)  echo "delay 50ms" ;;     # extra round-trip time, LEO-like
    d200) echo "delay 200ms" ;;
    d600) echo "delay 600ms" ;;    # GEO-satellite-like
    l1)   echo "loss 1%" ;;
    l5)   echo "loss 5%" ;;
    l10)  echo "loss 10%" ;;
    out)  echo "delay 0ms" ;;      # outages are switched on/off below
    *) echo "unknown profile: $1" >&2; exit 1 ;;
  esac
}

wexec()      { docker compose exec -T worker1 "$@"; }
set_link()   { wexec tc qdisc replace dev eth0 root netem $1; }
# Outage = drop everything in both directions (like losing the satellite).
outage_on()  { wexec tc qdisc replace dev eth0 root netem loss 100%
               wexec iptables -I INPUT -i eth0 -j DROP; }
outage_off() { wexec iptables -D INPUT -i eth0 -j DROP 2>/dev/null || true
               set_link "delay 0ms"; }
outage_loop() {
  while true; do
    sleep $((OUT_CYCLE - OUT_DOWN)); outage_on
    sleep "$OUT_DOWN";                outage_off
  done
}

start_testbed() {   # fresh controller (empty job store) + worker
  MODE=$1 docker compose up -d --force-recreate >/dev/null 2>&1
  for _ in $(seq 1 30); do
    curl -sf "$API/workers" 2>/dev/null | grep -q '"ID":"w1"' && return 0
    sleep 1
  done
  echo "worker never registered; check: docker compose logs" >&2; exit 1
}

# ---- preflight ----
docker image inspect cccleo-testjob:local >/dev/null 2>&1 || {
  echo "Missing test image. Build it first (see README step 1)." >&2; exit 1; }
echo "Building image..."
docker compose build -q
start_testbed pull
if ! set_link "delay 0ms" 2>/dev/null; then
  echo "tc/netem failed inside the container: this Docker kernel may lack netem." >&2
  docker compose down >/dev/null 2>&1; exit 1
fi

# ---- main loop ----
for rep in $(seq 1 "$REPS"); do
for profile in $PROFILES; do
for mode in $MODES; do
  name="${mode}_${profile}_r${rep}"
  args=$(netem_args "$profile")
  echo "== $name  (netem: $args)"

  start_testbed "$mode"
  set_link "$args"
  loop_pid=""
  if [ "$profile" = out ]; then outage_loop & loop_pid=$!; fi

  for i in $(seq 1 "$JOBS"); do
    curl -sf -X POST "$API/jobs" \
      -d '{"image":"cccleo-testjob:local","max_retries":1}' >/dev/null
    # random extra 0-999 ms so jobs don't always land at the same point
    # in the worker's poll cycle / the controller's push tick
    sleep "$(printf '%d.%03d' "$GAP" $((RANDOM % 1000)))"
  done

  if [ -n "$loop_pid" ]; then
    kill "$loop_pid" 2>/dev/null || true
    sleep 2        # let any tc/iptables command already in progress finish
    outage_off
  fi
  set_link "delay 0ms"   # let stragglers finish on a clean link

  for _ in $(seq 1 "$DRAIN"); do
    m=$(curl -sf "$API/metrics")
    echo "$m" | grep -q '"pending":0,' && echo "$m" | grep -q '"running":0,' && break
    sleep 1
  done

  curl -sf "$API/jobs"    > "results/$name.jobs.json"
  curl -sf "$API/metrics" > "results/$name.metrics.json"
  printf '{"mode":"%s","profile":"%s","rep":%d,"netem":"%s","poll_ms":%d,"tick_ms":%d,"out_down_s":%d,"out_cycle_s":%d}\n' \
    "$mode" "$profile" "$rep" "$args" "$POLL" "$TICK" "$OUT_DOWN" "$OUT_CYCLE" \
    > "results/$name.meta.json"
  echo "   $(cat "results/$name.metrics.json")"
done; done; done

docker compose down >/dev/null 2>&1
echo "Done. Now run: python analyze.py"
