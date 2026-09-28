#!/usr/bin/env sh
set -eu

here=$(dirname "$(readlink -f "$0")")
src="$here/../commitkind"
cache="${XDG_CACHE_HOME:-$HOME/.cache}/commit-review"
bin="$cache/commitkind"

command -v go >/dev/null 2>&1 || { echo "go is required to build commitkind" >&2; exit 1; }

if [ ! -x "$bin" ] || [ -n "$(find "$src" -name '*.go' -newer "$bin" 2>/dev/null | head -n 1)" ]; then
    mkdir -p "$cache"
    (cd "$src" && go build -o "$bin" .) >&2
fi

printf '%s\n' "$bin"
