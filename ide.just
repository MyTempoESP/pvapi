set quiet

[private]
default:
	just --list IDE


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

