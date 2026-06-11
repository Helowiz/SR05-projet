#!/bin/bash

APP=$1
CTL=$2
NET=$3

FIFOS=(/tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1 /tmp/in_N1 /tmp/out_N1
       /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2 /tmp/in_N2 /tmp/out_N2
       /tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3 /tmp/in_N3 /tmp/out_N3
       /tmp/in_A4 /tmp/out_A4 /tmp/in_C4 /tmp/out_C4 /tmp/in_N4 /tmp/out_N4
       /tmp/in_A5 /tmp/out_A5 /tmp/in_C5 /tmp/out_C5 /tmp/in_N5 /tmp/out_N5
       /tmp/in_A6 /tmp/out_A6 /tmp/in_C6 /tmp/out_C6 /tmp/in_N6 /tmp/out_N6
       /tmp/in_A7 /tmp/out_A7 /tmp/in_C7 /tmp/out_C7 /tmp/in_N7 /tmp/out_N7)

cleanup() {
    kill $(jobs -p) 2>/dev/null
    rm -f "${FIFOS[@]}"
}

trap cleanup EXIT INT TERM

mkfifo "${FIFOS[@]}"

cat /tmp/out_A1 > /tmp/in_C1 &
cat /tmp/out_A2 > /tmp/in_C2 &
cat /tmp/out_A3 > /tmp/in_C3 &
cat /tmp/out_A4 > /tmp/in_C4 &
cat /tmp/out_A5 > /tmp/in_C5 &

cat /tmp/out_C1 | tee /tmp/in_N1 > /tmp/in_A1 &
cat /tmp/out_C2 | tee /tmp/in_N2 > /tmp/in_A2 &      
cat /tmp/out_C3 | tee /tmp/in_N3 > /tmp/in_A3 &
cat /tmp/out_C4 | tee /tmp/in_N4 > /tmp/in_A4 &
cat /tmp/out_C5 | tee /tmp/in_N5 > /tmp/in_A5 &

cat /tmp/out_N1 > /tmp/in_C1 &
cat /tmp/out_N2 > /tmp/in_C2 &
cat /tmp/out_N3 > /tmp/in_C3 &
cat /tmp/out_N4 > /tmp/in_C4 &



cat /tmp/out_A6 > /tmp/in_C6 &
cat /tmp/out_A7 > /tmp/in_C7 &


cat /tmp/out_C6 | tee /tmp/in_N6 > /tmp/in_A6 &
cat /tmp/out_C7 | tee /tmp/in_N7 > /tmp/in_A7 &


cat /tmp/out_N5 > /tmp/in_C5 &
cat /tmp/out_N6 > /tmp/in_C6 &
cat /tmp/out_N7 > /tmp/in_C7 &


$APP --port 4444 -id 1 < /tmp/in_A1 > /tmp/out_A1 &
$CTL -n C1 -nbsites 7 -id 1 < /tmp/in_C1 > /tmp/out_C1 &
$NET -p 8080 -a < /tmp/in_N1 > /tmp/out_N1 &

$APP --port 4445 -id 2 < /tmp/in_A2 > /tmp/out_A2 &
$CTL -n C2 -nbsites 7  -id 2 < /tmp/in_C2 > /tmp/out_C2 &
$NET -p 8081 -a -sp 8080 < /tmp/in_N2 > /tmp/out_N2 &


$APP --port 4450 -id 3 < /tmp/in_A3 > /tmp/out_A3 &
$CTL -n C3 -nbsites 7  -id 3 < /tmp/in_C3 > /tmp/out_C3 &
$NET -p 8087 -a -sp 8080 < /tmp/in_N3 > /tmp/out_N3 &


$APP --port 4446 -id 4 < /tmp/in_A4 > /tmp/out_A4 &
$CTL -n C4 -nbsites 7 -id 4 < /tmp/in_C4 > /tmp/out_C4 &
$NET -p 8082 -sp 8081 < /tmp/in_N4 > /tmp/out_N4 &
$APP --port 4447    -id 5 < /tmp/in_A5 > /tmp/out_A5 &
$CTL -n C5 -nbsites 7 -id 5 < /tmp/in_C5 > /tmp/out_C5 &
$NET -p 8083 -sp 8080 < /tmp/in_N5 > /tmp/out_N5 &


sleep 5 # wait for the apps to start and connect to each other, to ensure the election works properly
$APP --port 4448   -id 6 < /tmp/in_A6 > /tmp/out_A6 &
$CTL -n C6 -nbsites 7 -id 6 < /tmp/in_C6 > /tmp/out_C6 &
$NET -p 8085 -sp 8083 < /tmp/in_N6 > /tmp/out_N6 &

$APP --port 4449   -id 7 < /tmp/in_A7 > /tmp/out_A7 &
$CTL -n C7 -nbsites 7 -id 7 < /tmp/in_C7 > /tmp/out_C7 &
$NET -p 8086 -sp 8082 < /tmp/in_N7 > /tmp/out_N7 &








sleep 3600