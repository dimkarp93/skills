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

Добавить или перезаписать один скилл на машине: `skills-sync add review-write` (то же, что
`pull review-write`; остальные скиллы не затрагиваются). Если в frontmatter скилла есть
`requires: a b`, эти скиллы синхронизируются вместе с ним; `--no-deps` отключает.

Скилл, которого нет в источнике, никогда не удаляется на приёмнике; удаления происходят только
внутри каталога того скилла, который синхронизируется. Перед записью скрипт всегда печатает
список изменений и спрашивает подтверждение, если что-то будет перезаписано или удалено.

## Скиллы для ревью коммитов

`review-write` («решай задачу в удобном для ревью виде») и `review-regroup` («подготовь к
ревью») раскладывают изменения по фичам и коммитам с темой `[фича] [вид]` (`[classify] [rename]`,
`[classify] [logic]`…): техническое (интерфейс) отдельно от логики `[logic]`, фичи идут по очереди,
обзор из топ-5 фич согласуется до начала. Только Go. Общее — в скилле `commit-review-core` (правила `RULES.md`,
sh-скрипты и Go-утилита `commitkind`, собирается при первом вызове в `~/.cache/commit-review`).
Префиксы хуков (`[PROJ-123]`) сохраняются, хуки не обходятся.

```sh
skills-sync add review-write review-regroup   # + commit-review-core по requires
cd skills/commit-review-core/commitkind && go test ./...
```

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
| bump-version | Поднять версию (patch / minor / major) собственной CLI-тулзы или библиотеки пользователя, у которой версия лежит в versions.txt, и опубликовать её — коммит versions.txt с сообщением «bump <level>», тег vX.Y.Z, push ветки и тегов во все remote. Сам определяет, чем запускать (just или make) и как называется цель (bump-patch или patch). Используй, когда просят поднять / бампнуть / повысить версию, «сделать patch/minor/major», выпустить новую версию или релиз тулзы, утилиты, CLI, библиотеки из ~/tools или ~/program, «just patch», «make bump-minor» и т.п. — даже если just/make и теги не названы. |
| commit-review-core | Общие правила и инструменты для скиллов review-write и review-regroup — формат темы [фича] [вид] и словарь видов ([rename], [add], [logic], …), порядок коммитов, совместимость с хуками и утилита commitkind (только Go). Сам не запускается; подгружается из review-write и review-regroup. |
| go-build | Собрать Go-программу по конвенциям репозитория dimkarp93/install (~/tools/install/CONVENTIONS.md) — версия из versions.txt, origin/upstream/commit/channel через ldflags, бинарь с именем репозитория в корне, рецепты build/test/bump-*/release в justfile. Используй, когда просят завести новую тулзу по конвенциям (init_install.sh), собрать Go-тулзу, подготовить её к установке через local_install.sh / github_install.sh / gitea_install.sh, завести Justfile или релизный workflow, добавить --version / --origin / --buildinfo или обновить Go до последнего стабильного релиза. Поднять версию — скилл bump-version. |
| install-tool | Создать новую устанавливаемую программу (CLI-тулзу) на Go или shell по конвенциям dimkarp93/install — скелет через init_install.sh, проверка через check_install.sh, установка через local_install.sh / github_install.sh / gitea_install.sh / go_install.sh. Используй, когда просят завести новую тулзу, утилиту, CLI, скрипт «как обычно» / «по конвенциям», сделать программу устанавливаемой, добавить versions.txt / justfile / workflow релиза / флаги --version, --origin, --buildinfo, починить существующий репозиторий под конвенции (check_install.sh --fix) или разобраться, почему check_install.sh ругается. Триггерься даже если init_install.sh и конвенции не названы явно — достаточно того, что человек заводит новую утилиту, которую потом будет ставить себе в PATH. |
| review-regroup | Подготовить ветку к ревью — уже сделанные коммиты перегруппировать в маленькие однотипные по фичам по очереди с темой [фича] [вид]: технические [rename] [add] [split] … отдельно от логики [logic] (обзор топ-5 фич до начала, каждый коммит собирается, итоговое дерево не меняется). Только Go. Используй ТОЛЬКО когда пользователь явно просит «подготовь к ревью», «перегруппируй коммиты для ревью» или вызывает /review-regroup; иначе не включай. |
| review-write | Решить задачу так, чтобы её было удобно ревьюить — правки сразу раскладываются по маленьким коммитам по фичам по очереди с темой [фича] [вид]: технические [rename] [add] [split] … отдельно от логики [logic] (обзор топ-5 фич до начала, каждый коммит собирается). Только Go. Используй ТОЛЬКО когда пользователь явно просит «решай задачу в удобном для ревью виде», «делай по коммитам для ревью» или вызывает /review-write; иначе не включай. |
<!-- skills:end -->
