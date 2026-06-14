cleanup() {
    echo "Stopping..."

    # kill all child processes
    pkill -P $$ 2>/dev/null

    # unblock FIFOs (important)
    for f in "${FIFOS[@]}"; do
        [ -p "$f" ] && echo "" > "$f" 2>/dev/null
    done

    rm -f "${FIFOS[@]}"
}