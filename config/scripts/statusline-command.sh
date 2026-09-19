#!/bin/sh
input=$(cat)
cwd=$(echo "$input" | jq -r '.workspace.current_dir // .cwd // empty')
dir=$(basename "$cwd")

# Git branch info (skip locks to avoid blocking)
branch=$(git -C "$cwd" --no-optional-locks symbolic-ref --short HEAD 2>/dev/null)
git_status=""
if [ -n "$branch" ]; then
    if git -C "$cwd" --no-optional-locks status --porcelain 2>/dev/null | grep -q .; then
        git_status=$(printf '\033[1;34mgit:(\033[0;31m%s\033[1;34m) \033[0;33m✗\033[0m' "$branch")
    else
        git_status=$(printf '\033[1;34mgit:(\033[0;31m%s\033[1;34m)\033[0m' "$branch")
    fi
fi

printf '\033[1;32m➜\033[0m  \033[0;36m%s\033[0m' "$dir"
if [ -n "$git_status" ]; then
    printf ' %s' "$git_status"
fi
printf '\n'
