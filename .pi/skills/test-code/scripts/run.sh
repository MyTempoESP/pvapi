#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$(readlink -f "$0")")/../../../.." && pwd)"

usage() {
  cat <<EOF
Usage: $0 [options] [test-pattern]

Options:
  -r, --run PATTERN    Go test -run pattern (default: none = full suite via just t)
  -v, --verbose        Enable -v flag on go test
  -c, --count N        Repeat count (default: 1)
  -t, --timeout DUR    Test timeout (default: 120s)
  --race               Enable race detector
  --short              Skip slow tests
  --no-lint            Skip golangci-lint
  --keep-db            Don't clean the test DB after run
  -h, --help           Show this help

Examples:
  $0 -r TestCartAndCheckout          # single test
  $0 -r 'TestCartAndCheckout/add'     # sub-test
  $0 -r TestAuth -v --race            # verbose + race
  $0                                  # full suite (just t)
EOF
  exit 0
}

TEST_RUN=""
VERBOSE=""
COUNT=1
TIMEOUT="120s"
RACE=""
SHORT=""
NO_LINT=""
KEEP_DB=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -r|--run)       TEST_RUN="$2"; shift 2 ;;
    -v|--verbose)   VERBOSE="-v"; shift ;;
    -c|--count)     COUNT="$2"; shift 2 ;;
    -t|--timeout)   TIMEOUT="$2"; shift 2 ;;
    --race)         RACE="-race"; shift ;;
    --short)        SHORT="-short"; shift ;;
    --no-lint)               NO_LINT=1; shift ;;
    --keep-db)               KEEP_DB=1; shift ;;
    -h|--help)    usage ;;
    *) echo "Unknown: $1"; usage ;;
  esac
done

cd "$PROJECT_ROOT"

# ── Lint ──────────────────────────────────────────────────────────────────────
if [[ -z "$NO_LINT" ]]; then
  echo "==> Running golangci-lint"
  golangci-lint run --fix || {
    echo "ERROR: lint failed. Fix issues and rerun."
    exit 1
  }
fi

# ── Full suite path ──────────────────────────────────────────────────────────
if [[ -z "$TEST_RUN" ]]; then
  echo "==> Running full test suite"
  just t
  exit $?
fi

# ── Targeted test path ───────────────────────────────────────────────────────
echo "==> Starting test database"
just t::test-db

cleanup() {
  if [[ -z "$KEEP_DB" ]]; then
    echo ""
    echo "==> Cleaning up test database"
    just t::clean
  else
    echo ""
    echo "==> Test DB left running (--keep-db)"
  fi
}
trap cleanup EXIT

FLAGS=()
[[ -n "$VERBOSE" ]] && FLAGS+=("-v")
[[ -n "$RACE"    ]] && FLAGS+=("-race")
[[ -n "$SHORT"   ]] && FLAGS+=("-short")
FLAGS+=("-count=$COUNT" "-timeout" "$TIMEOUT" "-run" "$TEST_RUN" ".")

echo "==> Running: go test ${FLAGS[*]}"
go test "${FLAGS[@]}"
