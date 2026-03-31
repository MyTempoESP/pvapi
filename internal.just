set script-interpreter := ['bash']

set quiet

default:
    echo "'internal' has no public targets."

[private]
[script]
[no-exit-message]
ensure t1 t2:
    cleanup() { just {{t2}}; }
    trap cleanup EXIT
    just {{t1}}
