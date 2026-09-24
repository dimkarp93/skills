---
name: go-build
description: Собрать Go-программу по конвенциям репозитория dimkarp93/install (~/tools/install/CONVENTIONS.md) — версия из versions.txt, origin/upstream/commit/channel через ldflags, бинарь с именем репозитория в корне, рецепты build/test/bump-*/release в justfile. Используй, когда просят завести новую тулзу по конвенциям (init_install.sh), собрать Go-тулзу, подготовить её к установке через local_install.sh / github_install.sh / gitea_install.sh, завести Justfile или релизный workflow, добавить --version / --origin / --buildinfo или обновить Go до последнего стабильного релиза. Поднять версию — скилл bump-version.
---

# Сборка Go-программы по CONVENTIONS.md

Источник истины — `~/tools/install/CONVENTIONS.md` (ru-версия: `CONVENTIONS.ru.md`).
Читай его, если нужна деталь, не описанная здесь.

## Шаг 1. Последний стабильный Go — обязательно

Перед любой сборкой проверь версию тулчейна:

```sh
~/.claude/skills/go-build/scripts/ensure-go.sh --check
```

Скрипт печатает `current` / `latest` (берёт `https://go.dev/dl/?mode=json`) и выходит с кодом 1,
если Go устарел или отсутствует. Тогда:

1. Скажи пользователю текущую и последнюю версии.
2. Установка идёт в существующий `GOROOT` (обычно `/usr/local/go`) и потому требует `sudo` —
   **спроси разрешение**, прежде чем запускать, и предложи выполнить самому:
   `! ~/.claude/skills/go-build/scripts/ensure-go.sh`
3. Скрипт качает архив, сверяет sha256 из того же JSON, подменяет каталог `GOROOT`.
   Флаг `--root DIR` ставит в произвольный префикс (например, `~/.local/go`) без sudo.

Не собирай устаревшим тулчейном молча. Если пользователь отказался обновляться — продолжай,
но скажи об этом явно.

После обновления Go приведи `go.mod` в порядок: директива `go` не должна быть выше
установленного тулчейна; если поднимаешь её — спроси.

## Шаг 2. Новый репозиторий — `init_install.sh`, а не шаблоны руками

Если репозитория ещё нет (или в нём нет `versions.txt`, `justfile`, workflow), не собирай скелет
вручную — зови скаффолдер из `~/tools/install`:

```sh
~/tools/install/init_install.sh --lang go --owner <owner> [--ci both] [--layout root] <name|path>
```

Он создаёт `versions.txt`, `justfile` (`build`, `check`, `bump-*`, `release`, `install`),
`.gitignore`, `cmd/<name>/main.go` с подключённым `github.com/dimkarp93/install-libs/buildinfo`,
`go.mod` с сетевым путём модуля, workflow релиза, делает `git init` и первый коммит, а в конце сам
прогоняет `check_install.sh`. Существующие файлы он не перезаписывает, так что поверх
наполовину готового репозитория запускай с `--force`.

Для shell-тулзы — `--lang sh` (скелет `<name>.sh` с подстановкой version/origin при сборке).

Существующий репозиторий, где чего-то не хватает, чинится не руками:

```sh
~/tools/install/check_install.sh --fix <path>
```

`--fix` дописывает `versions.txt`, `.gitignore`, рецепты `bump-*` и workflow релиза. `Makefile` он
не правит — там цели добавляй сам.

## Шаг 3. Чем собирать: just, не make

По умолчанию — `just` и `justfile` (заглавный `Justfile` тоже годится, следуй тому, что уже есть
в репозитории).

`make` используй **только** если:
- пользователь явно попросил make, либо
- в проекте уже есть сборка на `Makefile` (и нет `Justfile`), и переделать её на `justfile`
  не просили.

Во втором случае не заменяй `Makefile` на `justfile` из шаблона, даже если так проще, а дописывай
в него недостающие цели по образцам для `Makefile` из `~/tools/install/CONVENTIONS.md`.

Если `just` не установлен, скажи об этом и предложи поставить (`cargo install just`,
`apt install just`, `go install github.com/casey/just@latest` — по обстоятельствам),
а не молча переходи на make.

Проверить соответствие конвенциям целиком:

```sh
~/tools/install/check_install.sh --build <path>
```

Он покрывает весь чек-лист: `versions.txt`, имя бинаря, цель сборки и рецепты `bump-*`, workflow
релиза с `SHA256SUMS`, формат тегов, `.gitignore`, подстановку origin, требования `go install`
(сетевой путь модуля, `cmd/<name>`, отсутствие `replace`, суффикс `/vN`) и вывод
`--version` / `--origin` / `--buildinfo`. Не пересказывай чек-лист по памяти — запусти скрипт.

## Шаг 4. Сборка и проверка

```sh
just build
./<name> --version
./<name> --buildinfo
```

Что проверить в выводе:
- `--version` печатает ровно содержимое `versions.txt`, без `v` и без лишнего текста;
- `origin` в `--buildinfo` — канонический https-URL или `local`, **без токена и без userinfo**;
- `channel=local` для локальной сборки.

Затем `just check` (vet + test) и `just check-conventions`, если такие рецепты есть.

## Что не делать

- Не коммить и не пушить без просьбы; поднять версию — только по явному запросу и через скилл
  `bump-version` (`bump-*` сам коммитит, ставит тег и пушит во все remote).
- Не вшивать `git remote get-url origin` в бинарь без нормализации: в remote бывает токен.
- Не копировать шаблоны `justfile` и workflow руками: единственный их источник —
  `init_install.sh` (`--emit` печатает любой шаблон).
- Не менять формат вывода флагов «для красоты» — это контракт с инсталляторами.
- Не добавлять комментарии в justfile, скрипты и Go-код.
