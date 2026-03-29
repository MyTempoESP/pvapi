set shell := ["bash", "-uc"]

prod_db_user     := "pvauto"
prod_db_pass     := "ajF0a&_jmKI19E#"
prod_db_pass_url := "ajF0a%26_jmKI19E%23"
prod_db_port     := "5434"
prod_db_name     := "pvauto_db"
prod_db_host     := "127.0.0.1"

prod_dir    := "/home/mytempogps/pvauto"
prod_db_dir := prod_dir + "/pgdata/18"

prod_api_service := "pvapi"
prod_db_service  := "pva-postgresql"

prod_db_url := (
    "postgresql://" +
        prod_db_user +
        ":" + prod_db_pass_url + "@127.0.0.1:" + prod_db_port +
        "/" + prod_db_name + "?sslmode=disable"
)

init := "systemctl --user"

fetch:
    git fetch && git merge

image:
    podman build -t linhazero/pvapi:latest

restart:
    {{init}} restart "{{prod_api_service}}"

kill:
    {{init}} stop "{{prod_api_service}}"
    {{init}} stop "{{prod_db_service}}"

_clean:
    podman unshare rm -rf "{{prod_db_dir}}"

fresh: kill _clean
    {{init}} start "{{prod_db_service}}"
    @until \
        pg_isready -h "{{prod_db_host}}" -p "{{prod_db_port}}"; do sleep 1; done
    just _migrate "{{prod_db_url}}"

ci: fetch image restart
prod: ci fresh restart

migrate:
    @just _migrate "{{prod_db_url}}"
    @printf "{{BOLD + BG_GREEN}}OK{{NORMAL}}\n"
