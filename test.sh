#!/bin/sh
# test.sh — run all unit tests manually.
#
# Usage: ./test.sh
#
# Requires Go and all module dependencies to be available
# (run "go mod download" first if needed).

set -e

cd "$(dirname "$0")"

echo "==> Running all unit tests..."
go test ./...

echo ""
echo "==> Running race detector on concurrency-sensitive packages..."
go test -race ./users/... ./storage/...

echo ""
echo "==> All tests passed."
