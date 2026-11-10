#!/bin/sh
set -eu
cd "$(dirname "$0")"

mkdir -p coverage
temp=$(mktemp -d)
trap 'rm -rf "$temp"' EXIT
mkdir -p "$temp/startup" "$temp/web"
printf '%s\n' '<!doctype html><title>Calculator</title>' > "$temp/web/index.html"

go test -coverprofile="$temp/unit.out" ./...
go build -cover -o "$temp/calculator" .

# The startup tests exercise the real CLI without binding a port.
expect_startup_failure() {
  if GOCOVERDIR="$temp/startup" PORT="$1" STATIC_DIR="$2" "$temp/calculator" > "$temp/server.log" 2>&1; then
    echo 'Expected server startup to fail' >&2
    exit 1
  fi
  if ! grep -Fq "$3" "$temp/server.log"; then
    cat "$temp/server.log" >&2
    exit 1
  fi
}

expect_startup_failure '' "$temp/missing" 'STATIC_DIR must contain index.html'
expect_startup_failure 'invalid:port' "$temp/web" 'too many colons'
expect_startup_failure 'invalid:port' '' 'too many colons'

go tool covdata textfmt -i="$temp/startup" -o="$temp/startup.out"
# Go's cover tool merges matching blocks from the two set-mode profiles.
{ cat "$temp/unit.out"; tail -n +2 "$temp/startup.out"; } > coverage/coverage.out
go tool cover -func=coverage/coverage.out
go tool cover -html=coverage/coverage.out -o coverage/index.html
