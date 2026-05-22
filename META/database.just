set quiet

quiet        := '>/dev/null 2>&1'
quiet-errors := '2>/dev/null'

[private]
default:
    :

[group('data')]
migrate url:
    podman run --rm -v ../internal/database/migrations:/migrations:Z \
        --network host \
        docker.io/migrate/migrate -path /migrations/ \
        -database \
        "{{url}}" \
        -verbose up

[private]
[no-exit-message]
pg_poller container:
    until (podman exec {{container}} pg_isready {{quiet}}) do \
        printf "."; sleep 1s; done

[group('connection')]
pg_poll container:
    timeout 10s \
        just database::pg_poller {{container}} {{quiet-errors}} \
    || { \
        EXIT=$?; printf "\n"; \
        [ $EXIT -eq 124 ] && echo "Database unreachable (timeout)."; \
        [ $EXIT -ne 124 ] && echo "Error ($EXIT) in database polling."; \
        exit 1; \
    }
    sleep 1s
    echo "PostgreSQL database is ready!"
