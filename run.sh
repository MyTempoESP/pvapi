#!/usr/bin/env bash

set -xeuo pipefail

cleanup() {
    podman stop test-postgres
}

podman run \
    --rm -d --name test-postgres \
    -e POSTGRES_PASSWORD=1234 \
    -e POSTGRES_USER=pvauto \
    -e POSTGRES_DB=pvapi \
    -v "/home/rodrigo/postgres_data:/var/lib/postgresql:Z" \
    -p 5433:5432 postgres
trap cleanup EXIT

sleep 1

migrate -path=migrations \
    -database \
        "postgresql://pvauto:1234@localhost:5433/pvapi?sslmode=disable" \
    -verbose up

export PV_API_SECRET=secret
export PV_API_DB_HOST=localhost
export PV_API_DB_USER=pvauto
export PV_API_DB_PASS=1234
export PV_API_DB_BASE=pvapi
export PV_API_DB_PORT=5433

./pvapi

