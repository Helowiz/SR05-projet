#!/bin/bash

APP=$1
CTL=$2
NET=$3

FIFOS=(/tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1 /tmp/in_N1 /tmp/out_N1
       /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2 /tmp/in_N2 /tmp/out_N2
       )

cleanup() {
    kill $(jobs -p) 2>/dev/null
    rm -f "${FIFOS[@]}"
}

trap cleanup EXIT INT TERM

mkfifo "${FIFOS[@]}"

cat /tmp/out_A1 > /tmp/in_C1 &
cat /tmp/out_A2 > /tmp/in_C2 &


cat /tmp/out_C1 | tee /tmp/in_N1 > /tmp/in_A1 &
cat /tmp/out_C2 | tee /tmp/in_N2 > /tmp/in_A2 &      


cat /tmp/out_N1 > /tmp/in_C1 &
cat /tmp/out_N2 > /tmp/in_C2 &





$APP --port 4444 -id 1 < /tmp/in_A1 > /tmp/out_A1 &
$CTL -n C1 -nbsites 2 -id 1 < /tmp/in_C1 > /tmp/out_C1 &
$NET -p 8080 -a < /tmp/in_N1 > /tmp/out_N1 &

$APP --port 4445 -id 2 < /tmp/in_A2 > /tmp/out_A2 &
$CTL -n C2 -nbsites 2 -id 2 < /tmp/in_C2 > /tmp/out_C2 &
$NET -p 8081 -a -sp 8080 < /tmp/in_N2 > /tmp/out_N2 &










sleep 3600