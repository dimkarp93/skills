#!/usr/bin/env sh
set -eu

usage() {
    echo "usage: rr-verify.sh [-b BUILD_CMD] [-t TREE_REF] BASE [END]" >&2
    exit 2
}

here=$(dirname "$(readlink -f "$0")")
build="go build -o /dev/null ./..."
tree_ref=""

while [ $# -gt 0 ]; do
    case "$1" in
        -b) [ $# -ge 2 ] || usage; build=$2; shift 2 ;;
        -t) [ $# -ge 2 ] || usage; tree_ref=$2; shift 2 ;;
        -*) usage ;;
        *) break ;;
    esac
done
[ $# -ge 1 ] || usage
base=$1
end=${2:-HEAD}
rc=0

cd "$(git rev-parse --show-toplevel)"

if [ -n "$tree_ref" ]; then
    if [ "$(git rev-parse "$end^{tree}")" = "$(git rev-parse "$tree_ref^{tree}")" ]; then
        echo "tree: identical to $tree_ref"
    else
        echo "tree: DIFFERS from $tree_ref" >&2
        git diff --stat "$tree_ref" "$end" >&2
        rc=1
    fi
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

for sha in $(git rev-list --reverse "$base..$end"); do
    rm -rf "$tmp/src"
    mkdir -p "$tmp/src"
    git archive "$sha" | tar -x -C "$tmp/src"
    if (cd "$tmp/src" && sh -c "$build") >/dev/null 2>&1; then
        echo "build ok   $(git log -1 --format='%h %s' "$sha")"
    else
        echo "build FAIL $(git log -1 --format='%h %s' "$sha")" >&2
        rc=1
    fi
done

ck=$("$here/ensure-commitkind.sh")
"$ck" check "$base..$end" || rc=1
exit "$rc"
