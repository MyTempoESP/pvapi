mod database

set shell := ["bash", "-uc"]

prod_db_user	 := "pvauto"
prod_db_pass	 := "REDACTED_DB_PASS"
prod_db_pass_url := "REDACTED_DB_PASS"
prod_db_port	 := "5434"
prod_db_name	 := "pvauto_db"
prod_db_host	 := "127.0.0.1"

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

prod_psql := (
	'psql ' +
		'-h ' + '{{prod_db_host}} ' +
		'-p ' + '{{prod_db_port}} ' +
		'-U ' + '{{prod_db_user}} ' +
		'-d ' + '{{prod_db_name}}'
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
@migrate: (database::migrate prod_db_url)
	printf "{{BOLD + BG_GREEN}}OK{{NORMAL}}\n"


[group('util')]
@migrate-manual m:
	{{prod_psql}} < '{{m}}'


[group('util')]
@fetch:
	echo "Fetching updates from remote"
	git fetch
	git merge

[group('util')]
@clean:
	podman image prune

[private]
restart: kill && db (database::pg_poll prod_db_container) api

[private]
db:
	{{init}} start "{{prod_db_service}}"

[private]
api:
	{{init}} start "{{prod_api_service}}"

[private]
kill:
	{{init}} stop "{{prod_api_service}}"
	{{init}} stop "{{prod_db_service}}"

[private]
erase:
	podman unshare rm -rf "{{prod_db_dir}}"

[private]
fresh: kill erase db (database::pg_poll prod_db_container) && migrate api

