#!/usr/bin/env sh
set -eu

mkdir -p ./tmp
templ generate ./internal/view
go build -gcflags="all=-N -l" -o ./tmp/server ./cmd/server
