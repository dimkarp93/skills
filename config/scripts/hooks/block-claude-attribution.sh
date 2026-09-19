#!/bin/sh
cmd=$(jq -r '.tool_input.command // empty')

case "$cmd" in
  *git*commit*|*gh*pr*|*git*tag*|*gh*release*) ;;
  *) exit 0 ;;
esac

if printf '%s' "$cmd" | grep -qiE 'co-authored-by:[[:space:]]*claude|noreply@anthropic\.com|generated with .{0,3}claude code|claude\.com/claude-code|claude\.ai/code|Claude-Session:'; then
  echo "Заблокировано: запрещено указывать Claude/Claude Code как автора или соавтора в коммитах, тегах и PR (см. ~/.claude/CLAUDE.md). Удали строки атрибуции и повтори." >&2
  exit 2
fi

exit 0
