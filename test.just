mod internal
mod database

set guards
set lazy

db_user	 := "test_user"
db_pass	 := "testing_database"
db_pass_url := "testing_database"
db_host	 := "127.0.0.1"
db_port	 := "5434"
db_name	 := "test_db"
db_container:= "test-postgres"

db_base_dir := join(invocation_directory(), "test_databases")
db_dir	  := join(db_base_dir, uuid())

db_url := (
	"postgresql://" +
		db_user +
		":" + db_pass_url + "@" + db_host +
		":" + db_port +
		"/" + db_name + "?sslmode=disable"
)

# time for the db and program to run before killed
max-life-db := "1200"
max-life-go := "300s"


# default recipe
[group('target')]
all: (internal::ensure 't::worker' 't::clean')


[group('tool')]
attach-db:
	PGPASSWORD='{{db_pass}}' psql \
		-h '{{db_host}}' -p '{{db_port}}' '{{db_name}}' -U '{{db_user}}'


[private]
@db-ui: test-db && attach-db
	printf "\n{{BOLD + BG_CYAN + GREEN}}PostgreSQL database still running.\n"
	printf "\n{{NORMAL + YELLOW}}Exit the console to stop...{{NORMAL}}\n"


[doc('Roll a simple test database')]
[group('target')]
[script]
db: (internal::ensure 't::db-ui' 't::clean')

alias b:=db


# internal

[private]
worker: db go


[private]
[group('target')]
@test-db: && (database::pg_poll db_container) (database::migrate db_url)
	echo "Creating directory '{{db_dir}}'"
	mkdir -p "{{db_dir}}"
	chmod 755 "{{db_dir}}"

	echo "Starting test database"
	?(podman run \
		--timeout {{max-life-db}} \
		--rm -d --name {{db_container}} \
		-e	   POSTGRES_DB="{{db_name}}" \
		-e	 POSTGRES_USER="{{db_user}}" \
		-e POSTGRES_PASSWORD="{{db_pass}}" \
		-v "{{db_dir}}:/var/lib/postgresql:Z" \
		-p "{{db_port}}:5432" docker.io/library/postgres || exit 1)


[group('target')]
@go *flags="-v":
	printf "{{BOLD + WHITE}}go test {{flags}} .{{NORMAL}}\n"
	go test {{flags}} .


[group('util')]
clean:
	-podman stop test-postgres
	-podman unshare rm -rf "{{db_base_dir}}"

