#!/usr/bin/env sh
set -eu

usage() {
    echo "usage: rr-regroup.sh PLAN_FILE" >&2
    exit 2
}

[ $# -eq 1 ] && [ -f "$1" ] || usage
plan=$(readlink -f "$1")
here=$(dirname "$(readlink -f "$0")")
root=$(git rev-parse --show-toplevel)
cd "$root"

base=""
source_ref=HEAD
build="go build -o /dev/null ./..."
branch_name=""

while IFS= read -r line; do
    case "$line" in
        base\ *) base=${line#base } ;;
        source\ *) source_ref=${line#source } ;;
        build\ *) build=${line#build } ;;
        branch\ *) branch_name=${line#branch } ;;
        step\ *) break ;;
    esac
done < "$plan"

[ -n "$base" ] || { echo "plan: 'base REV' is required" >&2; exit 1; }
git rev-parse -q --verify "$base^{commit}" >/dev/null || { echo "unknown base: $base" >&2; exit 1; }
git rev-parse -q --verify "$source_ref^{commit}" >/dev/null || { echo "unknown source: $source_ref" >&2; exit 1; }
source_sha=$(git rev-parse "$source_ref")
if [ -z "$branch_name" ]; then
    cur=$(git symbolic-ref -q --short HEAD || echo detached)
    branch_name="review/$cur"
fi
if git rev-parse -q --verify "refs/heads/$branch_name" >/dev/null; then
    echo "branch $branch_name already exists: pick another name with 'branch NAME' in the plan" >&2
    exit 1
fi

ck=$("$here/ensure-commitkind.sh")
tags=$("$ck" tags | tr ' ' '|')
re="\\[[0-9a-z-]+\\] \\[($tags)\\] [^ ]"

wt=$(mktemp -d)
rmdir "$wt"
git worktree add -q -b "$branch_name" "$wt" "$base"
echo "worktree: $wt"
echo "branch:   $branch_name"

msg=""
body=""
open=0
count=0

fail() {
    echo "FAILED: $*" >&2
    echo "worktree kept at $wt (branch $branch_name); remove with: git worktree remove --force $wt && git branch -D $branch_name" >&2
    exit 1
}

finish() {
    [ "$open" -eq 1 ] || return 0
    open=0
    printf '%s\n' "$msg" | grep -Eq "$re" || fail "message needs [feature] [kind] text (feature: lowercase a-z, 0-9, dash): $msg"
    git -C "$wt" diff --cached --quiet && fail "nothing staged for: $msg"
    (cd "$wt" && sh -c "$build") >/dev/null 2>&1 || {
        (cd "$wt" && sh -c "$build") 2>&1 | tail -n 20 >&2
        fail "build failed for: $msg"
    }
    if [ -n "$body" ]; then
        git -C "$wt" commit -q -m "$msg" -m "$body" || fail "commit rejected (hook?) for: $msg"
    else
        git -C "$wt" commit -q -m "$msg" || fail "commit rejected (hook?) for: $msg"
    fi
    subject=$(git -C "$wt" log -1 --format=%s)
    printf '%s\n' "$subject" | grep -Eq "$re" || fail "hooks removed the tag from: $subject"
    count=$((count + 1))
    git -C "$wt" log -1 --format='  %h %s'
    body=""
}

while IFS= read -r line; do
    case "$line" in
        ""|base\ *|source\ *|build\ *|branch\ *|\#*) ;;
        step\ *)
            finish
            msg=${line#step }
            open=1
            ;;
        body\ *) body=${line#body } ;;
        file\ *)
            rest=${line#file }
            rev=${rest%% *}
            path=${rest#* }
            [ "$rev" != "-" ] || rev=$source_sha
            git -C "$wt" checkout "$rev" -- "$path" || fail "file $rev $path"
            ;;
        gen\ *)
            rest=${line#gen }
            rev=${rest%% *}
            rest=${rest#* }
            path=${rest%% *}
            spec=""
            case "$rest" in *\ *) spec=${rest#* } ;; esac
            [ "$rev" != "-" ] || rev=$source_sha
            case "$spec" in
                keep=\*) stage_arg="--all" ;;
                keep=*) stage_arg="--keep=${spec#keep=}" ;;
                "") stage_arg="--keep=" ;;
                *) fail "bad gen spec: $line" ;;
            esac
            mkdir -p "$(dirname "$wt/$path")"
            git -C "$wt" show "$rev:$path" | "$ck" stage "$stage_arg" > "$wt/$path" || fail "$line"
            git -C "$wt" add -- "$path" || fail "$line"
            ;;
        delete\ *) git -C "$wt" rm -q -r -- "${line#delete }" || fail "$line" ;;
        patch\ *) git -C "$wt" apply --index "${line#patch }" || fail "$line" ;;
        *) fail "unknown plan line: $line" ;;
    esac
done < "$plan"
finish

[ "$count" -gt 0 ] || fail "plan has no steps"

new_tree=$(git rev-parse "$branch_name^{tree}")
src_tree=$(git rev-parse "$source_sha^{tree}")
rc=0
if [ "$new_tree" = "$src_tree" ]; then
    echo "tree: identical to $source_ref"
else
    echo "tree: DIFFERS from $source_ref" >&2
    git diff --stat "$source_sha" "$branch_name" >&2
    rc=1
fi
"$ck" check "$base..$branch_name" || rc=1
git worktree remove --force "$wt"
echo
"$ck" overview "$base..$branch_name" --top=5 || true
echo
echo "branch $branch_name: $count commits on top of $(git rev-parse --short "$base")"
exit "$rc"
