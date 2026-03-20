#!/usr/bin/env bash

set -xeuo pipefail

podman run -v internal/database/migrations:/migrations \
    --network host \
    docker.io/migrate/migrate -path /migrations/ \
        -database \
            "postgresql://pvauto:ajF0a%26_jmKI19E@127.0.0.1:5434/pvauto_db?sslmode=disable" \
        -verbose up

read
