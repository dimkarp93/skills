#!/usr/bin/env sh
set -eu

here=$(dirname "$(readlink -f "$0")")

usage() {
    echo "usage: rr-plan.sh base | plan [BASE]" >&2
    exit 2
}

find_base() {
    for ref in origin/HEAD origin/master origin/main master main; do
        if git rev-parse -q --verify "$ref^{commit}" >/dev/null 2>&1; then
            mb=$(git merge-base HEAD "$ref")
            if [ "$mb" != "$(git rev-parse HEAD)" ]; then
                printf '%s\n' "$mb"
                return 0
            fi
        fi
    done
    echo "cannot find a base: pass it explicitly" >&2
    return 1
}

[ $# -ge 1 ] || usage
cmd=$1
base=${2:-}

case "$cmd" in
    base) find_base ;;
    plan)
        [ -n "$base" ] || base=$(find_base)
        ck=$("$here/ensure-commitkind.sh")
        echo "base: $(git log -1 --format='%h %s' "$base")"
        echo
        git diff --stat "$base" HEAD
        echo
        "$ck" hint "$base..HEAD"
        ;;
    *) usage ;;
esac
