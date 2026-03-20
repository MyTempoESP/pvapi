#!/usr/bin/env bash

set -xeuo pipefail

podman run -v internal/database/migrations:/migrations \
    --network host \
    docker.io/migrate/migrate -path /migrations/ \
        -database \
            "postgresql://pvauto:REDACTED_DB_PASS@127.0.0.1:5434/pvauto_db?sslmode=disable" \
        -verbose up

read
