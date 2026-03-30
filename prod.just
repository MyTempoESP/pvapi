mod tools

set shell := ["bash", "-uc"]

prod_db_user     := "pvauto"
prod_db_pass     := "REDACTED_DB_PASS"
prod_db_pass_url := "REDACTED_DB_PASS"
prod_db_port     := "5434"
prod_db_name     := "pvauto_db"
prod_db_host     := "127.0.0.1"

prod_dir    := "/home/mytempogps/pvauto"
prod_db_dir := prod_dir + "/pgdata/18"

prod_api_service := "pvapi"
prod_db_service  := "pva-postgresql"
prod_db_container:= "systemd-pva-postgresql"

prod_db_url := (
    "postgresql://" +
        prod_db_user +
        ":" + prod_db_pass_url + "@127.0.0.1:" + prod_db_port +
        "/" + prod_db_name + "?sslmode=disable"
)

init := "systemctl --user"

[private]
@default:
    just --list prod

[group('target')]
ci: fetch image restart

[group('target')]
new: fetch image fresh

[group('util')]
image:
    podman build -t linhazero/pvapi:latest .

[group('util')]
@migrate: (tools::migrate prod_db_url)
    printf "{{BOLD + BG_GREEN}}OK{{NORMAL}}\n"

[group('util')]
@fetch:
    echo "Fetching updates from remote"
    git fetch
    git merge

[private]
restart:
    {{init}} restart "{{prod_api_service}}"

[private]
kill:
    {{init}} stop "{{prod_api_service}}"
    {{init}} stop "{{prod_db_service}}"

[private]
erase:
    podman unshare rm -rf "{{prod_db_dir}}"

[private]
fresh: kill erase restart (tools::pg_poll prod_db_container) migrate
