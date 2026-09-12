#!/usr/bin/env sh
set -eu

usage() {
    cat <<EOF
Usage: $(basename "$0") [FLAGS]

Links skills-sync from this repository into a directory on your PATH.

  --bin DIR    install into DIR (default: ~/.local/bin)
  -f, --force  replace an existing file without asking
  -h, --help   show this help
EOF
}

BIN="$HOME/.local/bin"
FORCE=0

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        -f|--force) FORCE=1; shift ;;
        --bin) [ $# -ge 2 ] || { echo "Error: --bin needs a directory" >&2; exit 1; }; BIN=$2; shift 2 ;;
        *) echo "Unknown argument: $1" >&2; usage >&2; exit 1 ;;
    esac
done

REPO=$(dirname "$(readlink -f "$0")")
SRC="$REPO/skills-sync"
TARGET="$BIN/skills-sync"

[ -x "$SRC" ] || { echo "Error: $SRC is missing or not executable" >&2; exit 1; }

mkdir -p "$BIN"

if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then
    if [ "$(readlink -f "$TARGET" 2>/dev/null || true)" = "$SRC" ]; then
        echo "Already installed: $TARGET"
        exit 0
    fi
    if [ "$FORCE" -eq 0 ]; then
        [ -t 0 ] || { echo "Error: $TARGET exists, rerun with --force" >&2; exit 1; }
        printf 'Replace %s? [y/N] ' "$TARGET"
        read -r answer
        case "$answer" in
            y|Y|yes|YES) ;;
            *) echo "Aborted"; exit 1 ;;
        esac
    fi
fi

ln -sfn "$SRC" "$TARGET"
echo "Installed: $TARGET -> $SRC"

case ":$PATH:" in
    *":$BIN:"*) ;;
    *) echo "Note: $BIN is not in your PATH" ;;
esac
