#!/usr/bin/env bash

set -euo pipefail

go clean -cache
go build -o ./bin/app ../main

ADMIS=false
DEV=false

for arg in "$@"; do
    case "$arg" in
        --admis) ADMIS=true ;;
        --local) DEV=true ;;
    esac
done

ARGS=(-p 8080)

[ "$ADMIS" = true ] && ARGS+=(-a)
[ "$DEV" = true ] && ARGS+=(-dev)

exec ./bin/app "${ARGS[@]}"