#!/bin/sh
# Runs the unit tests covering the users storage, the bolt users
# backend and the user-related HTTP handlers.
set -e

echo "==> users package (Validate/Init, LastUpdate concurrency)"
go test -race -v ./users/

echo "==> bolt storage (transactional Update)"
go test -race -v ./storage/bolt/

echo "==> http handlers (user handler error format)"
go test -race -v ./http/

echo "==> all tests passed"
