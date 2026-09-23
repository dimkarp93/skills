#!/usr/bin/env sh
set -eu

usage() {
    echo "usage: bump.sh patch|minor|major [repo-dir]" >&2
    exit 2
}

[ $# -ge 1 ] || usage
level=$1
case "$level" in
    patch|minor|major) ;;
    *) usage ;;
esac

cd "${2:-.}"
root=$(git rev-parse --show-toplevel)
cd "$root"

[ -f versions.txt ] || { echo "versions.txt not found in $root" >&2; exit 1; }
if [ -n "$(git status --porcelain -- versions.txt)" ]; then
    echo "versions.txt has uncommitted changes - commit or restore it first" >&2
    exit 1
fi
git symbolic-ref -q HEAD >/dev/null || { echo "detached HEAD - switch to a branch first" >&2; exit 1; }

runner=""
if [ -f justfile ] || [ -f Justfile ] || [ -f .justfile ]; then
    if command -v just >/dev/null 2>&1; then
        runner=just
    else
        echo "warning: justfile found but just is not installed" >&2
    fi
fi
if [ -z "$runner" ] && { [ -f Makefile ] || [ -f makefile ] || [ -f GNUmakefile ]; }; then
    runner=make
fi

has_target() {
    case "$runner" in
        just) just --summary 2>/dev/null | tr ' ' '\n' | grep -qx "$1" ;;
        make) make -pRrq : 2>/dev/null | grep -Eq "^$1:" ;;
        *) return 1 ;;
    esac
}

target=""
for t in "bump-$level" "$level"; do
    if has_target "$t"; then
        target=$t
        break
    fi
done

for r in $(git remote); do
    git fetch -q --tags "$r" 2>/dev/null || echo "warning: git fetch --tags $r failed" >&2
done

old=$(tr -d '[:space:]' < versions.txt)
before=$(git rev-parse HEAD)
rc=0

if [ -n "$target" ]; then
    echo "resolved: $runner $target"
    "$runner" "$target" || rc=$?
else
    echo "resolved: no bump-$level/$level target in ${runner:-a justfile or Makefile} - bumping versions.txt directly"
    IFS=. read -r MAJ MIN PAT <<V
$old
V
    case "$level" in
        patch) new="$MAJ.$MIN.$((PAT + 1))" ;;
        minor) new="$MAJ.$((MIN + 1)).0" ;;
        major) new="$((MAJ + 1)).0.0" ;;
    esac
    printf '%s\n' "$new" > versions.txt
fi

new=$(tr -d '[:space:]' < versions.txt)
tag="v$new"
after=$(git rev-parse HEAD)

if [ "$new" = "$old" ] && [ "$after" = "$before" ] && [ "$rc" -ne 0 ]; then
    echo "$runner $target failed" >&2
    exit "$rc"
fi
if [ "$new" = "$old" ]; then
    echo "versions.txt was not changed (still $old)" >&2
    exit 1
fi

if [ "$after" = "$before" ]; then
    echo "the target only edited versions.txt - committing, tagging and pushing"
    if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
        git checkout -- versions.txt
        echo "tag $tag already exists" >&2
        exit 1
    fi
    git commit -q -m "bump $level" -- versions.txt
    git tag "$tag"
    rc=0
    for r in $(git remote); do
        git push -q "$r" HEAD --tags || { echo "push to $r failed" >&2; rc=1; }
    done
fi

git rev-parse -q --verify "refs/tags/$tag" >/dev/null || { echo "tag $tag was not created" >&2; exit 1; }

echo
echo "version: $old -> $new"
echo "commit:  $(git log -1 --format='%h %s')"
echo "tag:     $tag -> $(git rev-parse --short "$tag^{commit}")"
branch=$(git symbolic-ref --short HEAD)
for r in $(git remote); do
    if git ls-remote --exit-code --tags "$r" "refs/tags/$tag" >/dev/null 2>&1; then
        t=ok
    else
        t=MISSING
        rc=1
    fi
    if [ "$(git ls-remote "$r" "refs/heads/$branch" 2>/dev/null | cut -f1)" = "$(git rev-parse HEAD)" ]; then
        b=ok
    else
        b=MISSING
        rc=1
    fi
    echo "remote $r: tag $t, branch $branch $b ($(git remote get-url "$r"))"
done
exit "$rc"
