# skills

Долговременное хранилище пользовательских скиллов Claude Code.

Рабочая копия скиллов живёт в `~/.claude/skills` — этот каталог не под версионным контролем
и не переносится между машинами. Здесь лежит их эталон, а `skills-sync` копирует их в обе
стороны.

Встроенные скиллы Claude Code и скиллы из плагинов сюда не кладутся — они ставятся вместе с CLI
и через маркетплейс.

## Установка

```sh
./install.sh
```

Кладёт симлинки `~/.local/bin/skills-sync` и `~/.local/bin/config-sync` на скрипты из
репозитория.

## Использование

```sh
skills-sync list                 # что есть с обеих сторон и что разошлось
skills-sync status go-build      # какие файлы разошлись и в какую сторону
skills-sync diff go-build        # полный diff
skills-sync push go-build        # ~/.claude/skills -> репозиторий
skills-sync pull                 # репозиторий -> ~/.claude/skills (все скиллы)
skills-sync index                # таблица для этого README
```

Флаги: `-n` — только показать изменения, `-y` — не спрашивать перед перезаписью и удалением,
`--dest DIR` — работать с другим каталогом скиллов.

Порядок работы: поправил скилл в `~/.claude/skills` — сделай `push` и закоммить. На новой машине —
`./install.sh && skills-sync pull && config-sync pull`.

Скилл, которого нет в источнике, никогда не удаляется на приёмнике; удаления происходят только
внутри каталога того скилла, который синхронизируется. Перед записью скрипт всегда печатает
список изменений и спрашивает подтверждение, если что-то будет перезаписано или удалено.

## Сохранение всего разом

```sh
./save.sh        # skills-sync push + config-sync push, затем git status
./save.sh -n     # только показать, что уедет в репозиторий
./save.sh -y     # не спрашивать перед перезаписью и удалением
```

Забирает текущее состояние машины в
репозиторий. Коммит не делает — печатает `git status --short` и команду для коммита.

## Конфигурация

В `config/` лежит эталон глобальной конфигурации Claude Code:

| Файл | Что это |
| --- | --- |
| `config/CLAUDE.md` | глобальный системный промпт, `~/.claude/CLAUDE.md` |
| `config/settings.json` | настройки и хуки, `~/.claude/settings.json` |
| `config/scripts/` | скрипты, на которые ссылаются команды из `settings.json` |

`~/.claude/settings.local.json` — помашинный файл, он намеренно не синхронизируется.

```sh
config-sync list                 # все цели и их статус
config-sync status               # что и в какую сторону разошлось
config-sync diff claude-md       # полный diff
config-sync scripts              # какие скрипты найдены в settings.json
config-sync push                 # ~/.claude -> репозиторий
config-sync pull settings        # репозиторий -> ~/.claude, только settings.json
```

Цели: `claude-md`, `settings` и `scripts`; без имени цели команда работает со всеми. Флаги те же,
что у `skills-sync`: `-n`, `-y`, `--dest DIR`. Перед перезаписью скрипт печатает diff и спрашивает
подтверждение, а `settings.json` дополнительно проверяется на валидность JSON.

### Скрипты из settings.json

Цель `scripts` ничего не настраивает вручную: `config-sync` читает `settings.json`, берёт из
команд (`statusLine`, хуки и прочее) все пути, ведущие внутрь `~/.claude`, и хранит эти файлы в
`config/scripts/` с тем же относительным путём и правом на исполнение. Добавил новый хук со
скриптом — достаточно `config-sync push`, отдельный список вести не нужно.

```
~/.claude/statusline-command.sh            <-> config/scripts/statusline-command.sh
~/.claude/hooks/block-claude-attribution.sh <-> config/scripts/hooks/block-claude-attribution.sh
```

Путь, на который ссылается `settings.json`, но которого нет ни в репозитории, ни на машине,
`config-sync status` показывает отдельной строкой — так видно битые ссылки в конфиге.

При `pull` цели идут в порядке `claude-md`, `settings`, `scripts`: сначала приезжает
`settings.json`, и уже по нему определяется список скриптов. Пути в `settings.json` абсолютные,
поэтому на чужой машине они разбираются по части после `/.claude/`.

## Скиллы

<!-- skills:begin -->
| Skill | Description |
| --- | --- |
| go-build | Собрать Go-программу по конвенциям репозитория dimkarp93/install (~/tools/install/CONVENTIONS.md) — версия из versions.txt, origin/upstream/commit/channel через ldflags, бинарь с именем репозитория в корне, рецепты build/test/bump-* в Justfile. Используй, когда просят собрать Go-тулзу, подготовить её к установке через local_install.sh / github_install.sh / gitea_install.sh, завести Justfile или релизный workflow, добавить --version / --origin / --buildinfo, поднять версию (bump) или обновить Go до последнего стабильного релиза. |
<!-- skills:end -->
