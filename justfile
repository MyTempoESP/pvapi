mod prod
mod gen

set shell := ['bash', '-uc']
set guards
set lazy

db_user     := "test_user"
db_pass     := "testing_database"
db_pass_url := "testing_database"
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

# testing environment vars
export PV_API_SECURE_COOKIE := "no"
export PV_API_SECRET        := "secret"
export PV_API_DB_HOST       := "127.0.0.1"
export PV_API_DB_USER       := "{{db_user}}"
export PV_API_DB_PASS       := "{{db_pass}}"
export PV_API_DB_BASE       := "{{db_name}}"
export PV_API_DB_PORT       := "{{db_port}}"

export PV_API_ADMIN_USER := "lzero"
export PV_API_ADMIN_MAIL := "sistemas@lzero.org"
export PV_API_ADMIN_PASS := "TestingPassword"
export PV_API_LOG_LEVEL  := "debug"

default:
    just --list

[group('test')]
[group('tool')]
[group('database')]
test_database:
    mkdir -p "{{db_dir}}"
    chmod 755 "{{db_dir}}"

    # wrap subshell to catch any non-1 exit codes
    # doing this mostly to keep cleanup working
    # if i mess some config up
    ?(podman run \
        --rm --name test-postgres \
        -e       POSTGRES_DB="{{db_name}}" \
        -e     POSTGRES_USER="{{db_user}}" \
        -e POSTGRES_PASSWORD="{{db_pass}}" \
        -v "{{db_dir}}:/var/lib/postgresql:Z" \
        -p "{{db_port}}:5432" docker.io/library/postgres \
        --timeout 600 || exit 1)

[group('test')]
[group('tool')]
test_run:
    ?go run .

[group('test')]
[group('cleanup')]
test_cleanup:
    # cleanup commands can fail
    -podman stop test-postgres
    -podman unshare rm -rf "{{db_base_dir}}"

[private]
_test_migrate: (_migrate db_url)

[group('test')]
[group('cmd')]
test: test_database (_wait_db_ready "127.0.0.1" db_port) _test_migrate test_run test_cleanup

[private]
_migrate url:
    podman run --rm -v ./internal/database/migrations:/migrations:Z \
        --network host \
        docker.io/migrate/migrate -path /migrations/ \
        -database \
        "{{url}}" \
        -verbose up

[private]
_wait_db_ready_worker host port:
    @until \
        pg_isready -h "{{host}}" -p "{{port}}"; do sleep 1; done
[private]
_wait_db_ready host port:
    @echo "Waiting for PostgreSQL @ {{host}}:{{port}}"
    @timeout 10s just _wait_db_ready_worker "{{host}}" "{{port}}"
    @echo "PostgreSQL is ready!"

[group('cleanup')]
cleanup: test_cleanup
