#!/usr/bin/env bash
set -euo pipefail

compose_file="${COMPOSE_FILE:-docker-compose.test.yml}"
export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-jaiu-test}"

cleanup() {
  docker compose -f "$compose_file" down -v --remove-orphans
}
trap cleanup EXIT

docker compose -f "$compose_file" down -v --remove-orphans || true
docker compose -f "$compose_file" up -d --force-recreate db redis
docker compose -f "$compose_file" run --rm migrate
docker compose -f "$compose_file" run --rm fixtures

export JAIU_INTEGRATION=true
export TEST_DATABASE_URL="postgres://jaiu:jaiu@localhost:5433/jaiu_test?sslmode=disable"
export TEST_REDIS_ADDR="localhost:6380"
export GOTOOLCHAIN=auto

go test -tags=integration ./internal/adapter/postgres ./internal/adapter/redis
