set quiet

list-header-style := BG_CYAN + BOLD + WHITE

[private]
default:
    just --list

[group('util')]
count-lines:
    find . \
        \( \
        -name "*.go" -or \
        -name "*.sh" -or \
        -name "*.sql" -or \
        -name "*.txt" -or \
        -name "*.py" -or \
        -name "*.just" -or \
        -name "*.ejs.t" \
        \) \
        -not -path "./vendor/*" \
        -not -path "./_templates/generator/*" \
        -not -path "./_templates/init/*" \
        -exec wc -l {} \; | awk '{s+=$1} END {print s}'

[private]
list *modules:
    for i in {{modules}}; do \
        printf "{{list-header-style}} $i {{NORMAL}}\n"; \
        just  --color=always --list $i | sed '1d'; \
    done
