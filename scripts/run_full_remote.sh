
FIFOS=(/tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1 /tmp/in_N1 /tmp/out_N1
       /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2 /tmp/in_N2 /tmp/out_N2
       )

cleanup() {
    kill $(jobs -p) 2>/dev/null
    rm -f "${FIFOS[@]}"
}

trap 'cleanup; exit 0' INT QUIT TERM


cleanup()



go clean -cache
echo "==> Compilation des programmes Go..."

if ! go build -o bin/app ../web/server.go; then
    echo "Erreur de compilation pour app"
    exit 1
fi

if ! go build -o bin/ctl ../cmd/controler/main.go; then
    echo "Erreur de compilation pour ctl"
    exit 1
fi


mkfifo "${FIFOS[@]}"

cat /tmp/out_A1 > /tmp/in_C1 &


cat /tmp/out_C1 | tee /tmp/in_N1 > /tmp/in_A1 & 


cat /tmp/out_N1 > /tmp/in_C1 &


./bin/app --port 4444 -id 1 < /tmp/in_A1 > /tmp/out_A1 &
./bin/ctl -n C1 < /tmp/in_C1 > /tmp/out_C1 &
./bin/net -p 8088 < /tmp/in_N1 > /tmp/out_N1 &

