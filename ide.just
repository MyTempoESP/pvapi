set quiet
set no-exit-message
set unstable

# TODO: Modularize

file-regex := '^EDIT '
hist       := '/home/rodrigo/.jedai_history~'
editor     := 'vis'

history-tail(n)     := f"grep -e '{{file-regex}}' '{{hist}}' | awk '!seen[$2]++' | tail -n {{n}}"
history-grab(n)     := f"$({{history-tail(n)}} | head -n 1 | cut -d' ' -f2-)"
history-numbered(n) := f"{{history-tail(n)}} | tee '{{hist}}' | cut -d' ' -f2- | tac | nl"

err(msg) := f'{{RED + BOLD}}error{{NORMAL}}: {{msg}}'

[private]
default:
	just --list ide


[group('files')]
edit file:
	echo 'EDIT {{file}}' >> '{{hist}}'
	{{editor}} {{file}}

alias e:=edit


[group('files')]
[arg('n', pattern='\d+')]
resume n='1':
	touch '{{hist}}'
	( FILE={{history-grab(n)}}; \
		[ -z "$FILE" ] && { \
			printf '{{err('History is empty!')}}\n'; \
			exit 1; \
		}; \
		echo "EDIT $FILE" >> '{{hist}}'; \
		{{editor}} $FILE; \
	)

alias r:=resume


[group('history')]
[arg('n', pattern='\d+')]
history n='10':
	touch '{{hist}}'
	{{history-numbered(n)}}

alias h:=history


[group('history')]
history-del n='1':
	touch '{{hist}}'
	[ "$(wc -l '{{hist}}' | cut -d' ' -f1)" -ge '{{n}}' ] || ( \
		printf '{{err('Number out of range!')}}\n'; \
		exit 1; \
	)
	FILE={{history-grab(n)}}; awk "\$2!=\"$FILE\"" '{{hist}}' | tee '{{hist}}'

alias hd:=history-del


[group('docs')]
help cmd:
	'{{cmd}}' --help | less


[group('search')]
definition sym window='10':
	rg -g!vendor -g'*.go' --heading -A '{{window}}' \
		-- 'func\s?(\(.+\))?\s?{{sym}}'

[group('search')]
search pat:
	rg -l --heading -g!vendor -g!'*.pgerd' -- '{{pat}}' | xargs

alias s:=search


[group('test')]
check:
	go vet


[group('setup')]
sam:
	find . -path './vendor/*' -prune -o -name '*.go' -exec sam {} + &


[group('setup')]
setup-dev:
	doas apk add \
		postgresql \
		podman \
		go \
		bash
	echo 'rodrigo:100000:65536' | doas tee /etc/subuid /etc/subgid
	doas modprobe tun
	doas rc-service cgroups start


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

