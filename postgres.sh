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
    -p 5433:5432 postgres
trap cleanup EXIT

sleep 5

read

