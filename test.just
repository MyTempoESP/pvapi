mod tools

set guards
set lazy

db_user     := "test_user"
db_pass     := "testing_database"
db_pass_url := "testing_database"
db_host     := "127.0.0.1"
db_port     := "5434"
db_name     := "test_db"
db_container:= "test-postgres"

db_base_dir := join(invocation_directory(), "test_databases")
db_dir      := join(db_base_dir, uuid())

db_url := (
    "postgresql://" +
        db_user +
        ":" + db_pass_url + "@" + db_host +
        ":" + db_port +
        "/" + db_name + "?sslmode=disable"
)

# ensures the test script properly executes cleanup
test_header := '''
    set -xeuo pipefail
    cleanup() {
        just test::clean
    }
    trap cleanup EXIT
'''

# default recipe
[group('cmd')]
[script('bash')]
all:
    {{test_header}}
    just test::worker

[private]
worker: db (tools::pg_poll db_container) (tools::migrate db_url) go

[group('tool')]
[group('database')]
@db:
    echo "Creating directory '{{db_dir}}'"
    mkdir -p "{{db_dir}}"
    chmod 755 "{{db_dir}}"

    echo "Starting test database"
    ?(podman run \
        --timeout 600 \
        --rm -d --name {{db_container}} \
        -e       POSTGRES_DB="{{db_name}}" \
        -e     POSTGRES_USER="{{db_user}}" \
        -e POSTGRES_PASSWORD="{{db_pass}}" \
        -v "{{db_dir}}:/var/lib/postgresql:Z" \
        -p "{{db_port}}:5432" docker.io/library/postgres || exit 1)

[group('code')]
[group('tool')]
@go *flags="-v":
    printf "{{BOLD + WHITE}}go run {{flags}} .{{NORMAL}}\n"
    ?PV_API_SECURE_COOKIE="no" \
    PV_API_SECRET="secret" \
    PV_API_DB_HOST="{{db_host}}" \
    PV_API_DB_USER="{{db_user}}" \
    PV_API_DB_PASS="{{db_pass}}" \
    PV_API_DB_BASE="{{db_name}}" \
    PV_API_DB_PORT="{{db_port}}" \
    PV_API_ADMIN_USER="lzero" \
    PV_API_ADMIN_MAIL="sistemas@lzero.org" \
    PV_API_ADMIN_PASS="TestingPassword" \
    PV_API_LOG_LEVEL="debug" \
        go run {{flags}} .

[group('tool')]
[group('cleanup')]
@clean:
    echo "Stopping test database"
    -podman stop test-postgres
    echo "Removing test database directory '{{db_base_dir}}'"
    -podman unshare rm -rf "{{db_base_dir}}"
