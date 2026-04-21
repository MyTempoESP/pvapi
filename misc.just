set quiet

list-header-style := BG_CYAN + BOLD + WHITE


[private]
default:
    just --list


[private]
list *modules:
    for i in {{modules}}; do \
        printf "{{list-header-style}} $i {{NORMAL}}\n"; \
        just  --color=always --list $i | sed '1d'; \
    done
