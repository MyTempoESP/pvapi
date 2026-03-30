set quiet

[private]
default:
    just --list

[group('cmd')]
count-lines:
    find . \
        \( \
        -name "*.go" -or \
        -name "*.sh" -or \
        -name "*.sql" -or \
        -name "*.txt" -or \
        -name "*.ejs.t" \
        \) \
        -not -path "./vendor/*" \
        -not -path "./_templates/generator/*" \
        -not -path "./_templates/init/*" \
        -exec wc -l {} \; | awk '{s+=$1} END {print s}'
