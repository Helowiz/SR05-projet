#!/bin/bash
# Directory relative to the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIFO_DIR="$SCRIPT_DIR/fifos"
FIFOS=(
"$FIFO_DIR/in_A1"  "$FIFO_DIR/out_A1"
"$FIFO_DIR/in_C1"  "$FIFO_DIR/out_C1"
"$FIFO_DIR/in_N1"  "$FIFO_DIR/out_N1"
"$FIFO_DIR/in_A2"  "$FIFO_DIR/out_A2"
"$FIFO_DIR/in_C2"  "$FIFO_DIR/out_C2"
"$FIFO_DIR/in_N2"  "$FIFO_DIR/out_N2"
)

# Parse arguments
NET_EXTRA_ARGS=()
for arg in "$@"; do
    case "$arg" in
        --admis) NET_EXTRA_ARGS+=("-a") ;;
        --local) NET_EXTRA_ARGS+=("-dev") ;;
    esac
done

cleanup() {
echo "Stopping..."
pkill -P $$
rm -f "${FIFOS[@]}"
rmdir "$FIFO_DIR" 2>/dev/null
}
trap 'cleanup; exit 0' INT QUIT TERM
mkdir -p "$FIFO_DIR"
mkfifo "${FIFOS[@]}"
cat "$FIFO_DIR/out_A1" > "$FIFO_DIR/in_C1" &
cat "$FIFO_DIR/out_C1" | tee "$FIFO_DIR/in_N1" > "$FIFO_DIR/in_A1" &
cat "$FIFO_DIR/out_N1" > "$FIFO_DIR/in_C1" &
./bin/app --port 4444 -id 1 < "$FIFO_DIR/in_A1" > "$FIFO_DIR/out_A1" &
./bin/ctl -n C1          < "$FIFO_DIR/in_C1"  > "$FIFO_DIR/out_C1" &
./bin/net -p 8080 "${NET_EXTRA_ARGS[@]}"       < "$FIFO_DIR/in_N1"  > "$FIFO_DIR/out_N1" &
sleep 3600