#!/usr/bin/env sh
set -eu

usage() {
    echo "usage: rr-commit.sh -m MESSAGE [-b BUILD_CMD] [--body TEXT] -- PATH..." >&2
    exit 2
}

here=$(dirname "$(readlink -f "$0")")
msg=""
body=""
build="go build -o /dev/null ./..."

while [ $# -gt 0 ]; do
    case "$1" in
        -m) [ $# -ge 2 ] || usage; msg=$2; shift 2 ;;
        -b) [ $# -ge 2 ] || usage; build=$2; shift 2 ;;
        --body) [ $# -ge 2 ] || usage; body=$2; shift 2 ;;
        --) shift; break ;;
        *) usage ;;
    esac
done
[ -n "$msg" ] && [ $# -gt 0 ] || usage

cd "$(git rev-parse --show-toplevel)"

ck=$("$here/ensure-commitkind.sh")
tags=$("$ck" tags | tr ' ' '|')
re="\\[[0-9a-z-]+\\] \\[($tags)\\] [^ ]"

printf '%s\n' "$msg" | grep -Eq "$re" || { echo "message must contain [feature] [kind] text, kinds: $("$ck" tags)" >&2; exit 1; }

git add -A -- "$@"
git diff --cached --quiet && { echo "nothing staged" >&2; exit 1; }

if [ -n "$build" ] && [ "$build" != "none" ]; then
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    git checkout-index -a -f --prefix="$tmp/"
    (cd "$tmp" && sh -c "$build") || { git restore --staged -- "$@"; echo "build failed on the staged state, nothing committed, files unstaged" >&2; exit 1; }
fi

if [ -n "$body" ]; then
    git commit -q -m "$msg" -m "$body"
else
    git commit -q -m "$msg"
fi

subject=$(git log -1 --format=%s)
printf '%s\n' "$subject" | grep -Eq "$re" || { echo "hooks removed the tag from: $subject" >&2; exit 1; }
git log -1 --format='%h %s'
