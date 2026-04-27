#!/bin/bash
# Generates traffic to all example endpoints.
# Requires: hey (go install github.com/rakyll/hey@latest)

BASE="http://localhost:3000"
DURATION=10s

while true; do
    echo "[$(date +%H:%M:%S)] Generating traffic..."

    # Normal requests
    hey -z "$DURATION" -q 10 "${BASE}/" > /dev/null &

    # Slow requests (demonstrates latency histogram)
    hey -z "$DURATION" -q 3 "${BASE}/slow" > /dev/null &

    # Failing requests (demonstrates error spans)
    hey -z "$DURATION" -q 2 "${BASE}/fail" > /dev/null &

    wait
    sleep 1
done
