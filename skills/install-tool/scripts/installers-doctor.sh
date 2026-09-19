#!/usr/bin/env sh
set -eu

REPO="${INSTALL_REPO:-$HOME/tools/install}"
SCRIPTS="github_install.sh gitea_install.sh local_install.sh go_install.sh check_install.sh init_install.sh"

if [ ! -d "$REPO" ]; then
    echo "Working copy not found: $REPO (set INSTALL_REPO to override)" >&2
    exit 2
fi

STATUS=0

for name in $SCRIPTS; do
    src="$REPO/$name"
    if [ ! -f "$src" ]; then
        printf '%-20s [FAIL] missing in the working copy\n' "$name"
        STATUS=1
        continue
    fi

    found=""
    IFS=:
    for dir in $PATH; do
        [ -n "$dir" ] || dir=.
        [ -f "$dir/$name" ] || continue
        found="$found $dir/$name"
    done
    unset IFS

    set -- $found
    if [ $# -eq 0 ]; then
        printf '%-20s [WARN] not in PATH\n' "$name"
        STATUS=1
        continue
    fi

    winner="$1"
    if cmp -s "$src" "$winner"; then
        printf '%-20s [OK]   %s\n' "$name" "$winner"
    else
        printf '%-20s [FAIL] %s is stale\n' "$name" "$winner"
        STATUS=1
    fi

    if [ $# -gt 1 ]; then
        shift
        for extra in "$@"; do
            printf '%-20s [WARN] shadowed copy: %s\n' "$name" "$extra"
        done
        STATUS=1
    fi
done

if [ "$STATUS" != 0 ]; then
    echo
    echo "Fix: $REPO/dev-bootstrap.sh   (or --user-only for ~/.local/bin)"
    echo "Remove the shadowed copies listed above so one directory stays authoritative."
fi

exit $STATUS
