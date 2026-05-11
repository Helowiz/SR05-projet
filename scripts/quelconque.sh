#!/bin/bash
APP=$1
CTL=$2

FIFOS=(/tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1
      /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2
      /tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3
      /tmp/in_A4 /tmp/out_A4 /tmp/in_C4 /tmp/out_C4)

cleanup() {
  kill $(jobs -p) 2>/dev/null
  rm -f "${FIFOS[@]}"
}
trap cleanup EXIT INT TERM

mkfifo "${FIFOS[@]}"

$APP --port 4444 -id 1 < /tmp/in_A1 > /tmp/out_A1 &
$CTL -n C1 -nbsites 4 < /tmp/in_C1 > /tmp/out_C1 &
$APP --port 4445 -id 2 < /tmp/in_A2 > /tmp/out_A2 &
$CTL -n C2 -nbsites 4 < /tmp/in_C2 > /tmp/out_C2 &
$APP --port 4446 -id 3 < /tmp/in_A3 > /tmp/out_A3 &
$CTL -n C3 -nbsites 4 < /tmp/in_C3 > /tmp/out_C3 &
$APP --port 4447 -id 4 < /tmp/in_A4 > /tmp/out_A4 &
$CTL -n C4 -nbsites 4 < /tmp/in_C4 > /tmp/out_C4 &

# APP -> CTL
cat /tmp/out_A1 > /tmp/in_C1 &
cat /tmp/out_A2 > /tmp/in_C2 &
cat /tmp/out_A3 > /tmp/in_C3 &
cat /tmp/out_A4 > /tmp/in_C4 &

# CTL outputs (directed)
cat /tmp/out_C1 | tee /tmp/in_A1 | tee /tmp/in_C2 > /tmp/in_C3 &
cat /tmp/out_C2 | tee /tmp/in_A2 > /tmp/in_C4 &
cat /tmp/out_C3 | tee /tmp/in_A3 > /tmp/in_C2 &
cat /tmp/out_C4 | tee /tmp/in_A4 > /tmp/in_C1 &

wait
