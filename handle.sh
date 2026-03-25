#!/usr/bin/env bash
set -xeuo pipefail

db_user="pvauto"
db_pass="REDACTED_DB_PASS"
db_pass_encoded="REDACTED_DB_PASS"
db_name="pvauto_db"
db_port="5434"
db_dir=""

pvapi_db_data_directory="/home/mytempogps/pvauto/pgdata/18"
pvapi_psql_service_name="pva-postgresql"
pvapi_main_service_name="pvapi"

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

migrations_create() {
    podman run --rm -v ./internal/database/migrations:/migrations:Z \
        --network host \
        docker.io/migrate/migrate \
        -verbose create -dir=/migrations/ -seq -ext sql "$1"
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

build_ci() {
    podman build -t linhazero/pvapi:latest .
    systemctl --user restart "$pvapi_main_service_name"

    # user responsibility:
    # podman image prune -f
}

database_reset() {
    systemctl --user stop "$pvapi_psql_service_name"
    systemctl --user stop "$pvapi_main_service_name"

    sleep 2
    podman unshare rm -rf "$pvapi_db_data_directory"

    systemctl --user start "$pvapi_psql_service_name"
    systemctl --user start "$pvapi_main_service_name"

    exec "$0" migrate
}

CMD=${1:-"help"}
shift

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

ci)
    build_ci
    ;;

prod)
    git fetch && git merge
    build_ci
    database_reset
    ;;

migrate)
    migrations_run
    ;;

new_migration)
    [[ -z "$1" ]] && exit 1
    migrations_create "$1"
    ;;

database)
    database_up
    read -r
    ;;

test)
    test_program
    ;;

*)
    echo "Usage: $0 {image|postgres|cleanup|migrate|database|test}"
    exit 1
    ;;
esac
