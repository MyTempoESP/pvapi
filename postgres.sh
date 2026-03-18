#!/usr/bin/env bash

set -xeuo pipefail

temp=$(mktemp -dp ~/databases)

chmod 755 ~/databases
chmod 755 "$temp"

cleanup() {
    podman stop test-postgres
    podman unshare rm -rf "$temp"
}

podman run \
    --rm -d --name test-postgres \
    -e POSTGRES_USER=pvauto \
    -e POSTGRES_PASSWORD=1234 \
    -e POSTGRES_DB=pvapi \
    -v "$temp:/var/lib/postgresql:Z" \
    -p 5433:5432 docker.io/library/postgres
trap cleanup EXIT

sleep 5

migrate -path internal/database/migrations \
    -database \
        "postgresql://pvauto:1234@localhost:5433/pvapi?sslmode=disable" \
    -verbose up

read

