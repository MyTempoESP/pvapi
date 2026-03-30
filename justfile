mod tools
mod prod
mod gen

set shell := ['bash', '-uc']
set script-interpreter := ['bash']
set guards

db_user     := "test_user"
db_pass     := "testing_database"
db_pass_url := "testing_database"
db_host     := "127.0.0.1"
db_port     := "5434"
db_name     := "test_db"

db_base_dir := join(invocation_directory(), "test_databases")
db_dir      := join(db_base_dir, uuid())

db_url := (
    "postgresql://" +
        db_user +
        ":" + db_pass_url + "@127.0.0.1:" + db_port +
        "/" + db_name + "?sslmode=disable"
)

# golang build flags
goflags := "-v"

# testing environment vars
export PV_API_SECURE_COOKIE := "no"
export PV_API_SECRET        := "secret"
export PV_API_DB_HOST       := db_host
export PV_API_DB_USER       := db_user
export PV_API_DB_PASS       := db_pass
export PV_API_DB_BASE       := db_name
export PV_API_DB_PORT       := db_port

export PV_API_ADMIN_USER := "lzero"
export PV_API_ADMIN_MAIL := "sistemas@lzero.org"
export PV_API_ADMIN_PASS := "TestingPassword"
export PV_API_LOG_LEVEL  := "debug"

# ensures the test script properly executes cleanup
test_header := '''
    set -euo pipefail

    cleanup() {
        just test_cleanup
    }

    trap cleanup EXIT
'''

default:
    just --list

[group('test')]
[group('tool')]
[group('database')]
@test_database:
    echo "Creating directory '{{db_dir}}'"
    mkdir -p "{{db_dir}}"
    chmod 755 "{{db_dir}}"

    echo "Starting test database"
    ?(podman run \
        --timeout 600 \
        --rm -d --name test-postgres \
        -e       POSTGRES_DB="{{db_name}}" \
        -e     POSTGRES_USER="{{db_user}}" \
        -e POSTGRES_PASSWORD="{{db_pass}}" \
        -v "{{db_dir}}:/var/lib/postgresql:Z" \
        -p "{{db_port}}:5432" docker.io/library/postgres || exit 1)

[group('test')]
[group('tool')]
test_run *args=goflags:
    ?go run {{args}} .

[group('test')]
[group('cleanup')]
[no-exit-message]
@test_cleanup:
    echo "cleanup commands can fail"
    echo "Stopping test database"
    -podman stop test-postgres
    echo "Removing test database directory '{{db_base_dir}}'"
    -podman unshare rm -rf "{{db_base_dir}}"

[private]
_test_migrate: (tools::_migrate db_url)

[private]
_test: test_database (tools::_wait_db "test-postgres") _test_migrate test_run

[group('test')]
[group('cmd')]
[script]
[no-exit-message]
test:
    {{test_header}}
    trap cleanup EXIT
    just _test

alias cleanup := test_cleanup
