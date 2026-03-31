set quiet

mod misc

[private]
default: (misc::list 'new')

[group('database')]
migration name:
    podman run --rm -v ./internal/database/migrations:/migrations:Z \
        --network host \
        docker.io/migrate/migrate \
        -verbose create -dir=/migrations/ -seq -ext sql "{{name}}"
