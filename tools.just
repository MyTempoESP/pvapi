set quiet

list-header-style := BG_CYAN + BOLD + WHITE

[private]
default: (list 'tools')

[private]
list *modules:
    for i in {{modules}}; do \
        printf "{{list-header-style}} $i {{NORMAL}}\n"; \
        just --color=always --list $i | sed 1d; \
    done

[group('database')]
migrate url:
    podman run --rm -v ./internal/database/migrations:/migrations:Z \
        --network host \
        docker.io/migrate/migrate -path /migrations/ \
        -database \
        "{{url}}" \
        -verbose up

[private]
[no-exit-message]
pg_poller container:
    podman exec {{container}} pg_isready 2>&1 > /dev/null

[private]
[no-exit-message]
pg_poll_worker container:
    until (just tools::pg_poller {{container}}) do \
        printf "."; sleep 1s; done

[group('database')]
pg_poll container:
    timeout 10s just tools::pg_poll_worker {{container}}
    sleep 1s
    echo "PostgreSQL database is ready!"
