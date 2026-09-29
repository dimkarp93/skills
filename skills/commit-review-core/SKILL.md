---
name: commit-review-core
description: Общие правила и инструменты для скиллов review-write и review-regroup — формат темы [фича] [вид] и словарь видов ([rename], [add], [logic], …), порядок коммитов, совместимость с хуками и утилита commitkind (только Go). Сам не запускается; подгружается из review-write и review-regroup.
user-invocable: false
---

# Общая часть скиллов review-write / review-regroup

Не выполняй ничего самостоятельно: этот скилл — справочник и набор скриптов для
`review-write` и `review-regroup`. Правила — в `RULES.md` рядом, прочитай его целиком.

Скрипты (`~/.claude/skills/commit-review-core/scripts/`, POSIX sh):

| Скрипт | Что делает |
|---|---|
| `ensure-commitkind.sh` | собирает Go-утилиту `commitkind` в `~/.cache/commit-review/`, печатает путь |
| `rr-plan.sh base\|plan [BASE]` | база ветки, черновой план по `commitkind hint` (для новых файлов — `commitkind funcs` и `order`) |
| `rr-regroup.sh PLAN` | перегруппировка: строит новую ветку `review/<ветка>` в отдельном worktree по файлу плана, ничего не сбрасывает и не перезаписывает в рабочей копии |
| `rr-commit.sh -m MSG [-b BUILD] [--body T] -- PATH..` | коммит файлов в текущей ветке: сборка индекса во временном каталоге, коммит через хуки, проверка темы |
| `rr-verify.sh [-b BUILD] [-t TREE_REF] BASE [END]` | каждый коммит собирается (из `git archive`), дерево совпадает с `TREE_REF`, `commitkind check` |

Скрипты не используют `git reset`, `git stash`, `git checkout -f` и не переписывают файлы
рабочей копии, поэтому их можно запускать из субагента. Если система прав всё же блокирует
вызов — остановись, ничего не обходи и сообщи пользователю, какой именно вызов заблокирован и
что состояние репозитория не менялось.

Только Go. Если проект не на Go — скажи об этом и предложи обычную работу без этого режима.
