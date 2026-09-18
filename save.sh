#!/usr/bin/env sh
set -eu

usage() {
    cat <<EOF
Usage: $(basename "$0") [FLAGS]

Copies the local skills and configuration into this repository, then shows
what git sees. Nothing is committed.

  -n, --dry-run    show what would change and stop
  -y, --yes        do not ask before overwriting or deleting files
  -h, --help       show this help
EOF
}

ARGS=""

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        -n|--dry-run) ARGS="$ARGS -n"; shift ;;
        -y|--yes) ARGS="$ARGS -y"; shift ;;
        *) echo "Unknown argument: $1" >&2; usage >&2; exit 1 ;;
    esac
done

REPO=$(dirname "$(readlink -f "$0")")

for tool in skills-sync config-sync; do
    [ -x "$REPO/$tool" ] || { echo "Error: $REPO/$tool is missing or not executable" >&2; exit 1; }
done

echo "== skills =="
"$REPO/skills-sync" $ARGS push
echo
echo "== config =="
"$REPO/config-sync" $ARGS push

if command -v git >/dev/null 2>&1 && git -C "$REPO" rev-parse --git-dir >/dev/null 2>&1; then
    echo
    echo "== git =="
    status=$(git -C "$REPO" status --short)
    if [ -z "$status" ]; then
        echo "nothing to commit"
    else
        printf '%s\n' "$status"
        echo
        echo "Review and commit: git -C $REPO add -A && git -C $REPO commit"
    fi
fi
