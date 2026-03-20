#!/usr/bin/env bash
set -xeuo pipefail

db_user="pvauto"
db_pass="REDACTED_DB_PASS"
db_pass_encoded="REDACTED_DB_PASS"
db_name="pvauto_db"
db_port="5434"
db_dir=""

cleanup() {
    echo "Cleaning up..."
    podman stop test-postgres || true
    podman unshare rm -rf "$db_dir" || true
}

postgres_start() {
    db_dir="$(mktemp -dp ~/databases)"

    chmod 755 ~/databases
    chmod 755 "$db_dir"

    podman run \
        --rm -d --name test-postgres \
        -e POSTGRES_USER="$db_user" \
        -e POSTGRES_PASSWORD="$db_pass" \
        -e POSTGRES_DB="$db_name" \
        -v "$db_dir:/var/lib/postgresql:Z" \
        -p "$db_port:5432" docker.io/library/postgres

    echo "Waiting for PostgreSQL to be ready..."

    sleep 5
}

migrations_run() {
    podman run --rm -v ./internal/database/migrations:/migrations:Z \
        --network host \
        docker.io/migrate/migrate -path /migrations/ \
            -database \
                "postgresql://${db_user}:${db_pass_encoded}@127.0.0.1:${db_port}/${db_name}?sslmode=disable" \
            -verbose up
}

database_up() {
    postgres_start
    trap cleanup EXIT

    migrations_run
}

test_program() {
    postgres_start
    trap cleanup EXIT

    migrations_run

    export PV_API_SECURE_COOKIE=no
    export PV_API_SECRET=secret
    export PV_API_DB_HOST=127.0.0.1
    export PV_API_DB_USER="$db_user"
    export PV_API_DB_PASS="$db_pass"
    export PV_API_DB_BASE="$db_name"
    export PV_API_DB_PORT="$db_port"

    export PV_API_ADMIN_USER=linhazero
    export PV_API_ADMIN_MAIL=sistemas@linhazero.com
    export PV_API_ADMIN_PASS="REDACTED_ADMIN_PASS"

    export PV_API_LOG_LEVEL=debug

    go run .
}

CMD=${1:-"help"}

case "$CMD" in
    image)
        podman rmi -f linhazero/pvapi:latest || true
        podman build -t linhazero/pvapi:latest .
        ;;

    postgres)
        postgres_start
        ;;

    cleanup)
        cleanup
        ;;

    migrate)
        migrations_run
        ;;

    database)
        database_up
        read
        ;;

    test)
        test_program
        ;;

    *)
        echo "Usage: $0 {image|postgres|cleanup|migrate|database|test}"
        exit 1
        ;;
esac
