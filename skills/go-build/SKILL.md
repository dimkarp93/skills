---
name: go-build
description: Собрать Go-программу по конвенциям репозитория dimkarp93/install (~/tools/install/CONVENTIONS.md) — версия из versions.txt, origin/upstream/commit/channel через ldflags, бинарь с именем репозитория в корне, рецепты build/test/bump-* в Justfile. Используй, когда просят собрать Go-тулзу, подготовить её к установке через local_install.sh / github_install.sh / gitea_install.sh, завести Justfile или релизный workflow, добавить --version / --origin / --buildinfo, поднять версию (bump) или обновить Go до последнего стабильного релиза.
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

## Шаг 2. Как собирать: just, не make

По умолчанию — `just` и `Justfile` (строчный `justfile` тоже годится, следуй тому, что уже есть
в репозитории).

`make` используй **только** если:
- в репозитории уже есть `Makefile` и нет `Justfile`, либо
- пользователь явно попросил make.

Если в репозитории нет ни того, ни другого — заводи `justfile` из
`~/.claude/skills/go-build/templates/justfile` (замени `bin := "NAME"` на имя репозитория,
поправь путь к `package main`: `./cmd/<name>` или `.`).

Если `just` не установлен, скажи об этом и предложи поставить (`cargo install just`,
`apt install just`, `go install github.com/casey/just@latest` — по обстоятельствам),
а не молча переходи на make.

## Шаг 3. Чек-лист соответствия конвенциям

Перед сборкой проверь и при необходимости заведи:

- `versions.txt` в корне — semver без `v` (`0.4.0`).
- Имя бинаря = имя каталога/репозитория; после `just build` он лежит **в корне** репозитория.
- `package main` в `cmd/<name>/` (нужно для `go install`), либо в корне — как fallback.
- В `go.mod` путь модуля — сетевой адрес (`github.com/<owner>/<name>`), для major ≥ 2 с суффиксом
  `/v2`.
- `upstream.txt` — опционален, нужен только зеркалам.
- Флаги `--version`, `--origin`, `--buildinfo` реализованы. Проще всего подключить
  `github.com/dimkarp93/install-libs/buildinfo` — шаблон
  `~/.claude/skills/go-build/templates/buildinfo.go.txt`. Формат вывода:
  `--version` → голый semver; `--origin` → канонический URL; `--buildinfo` → строки
  `origin/upstream/version/commit/channel` в этом порядке.
- Сборка статическая: `CGO_ENABLED=0`, `-trimpath`,
  `-ldflags="-s -w -X main.version=... -X main.origin=... -X main.upstream=... -X main.commit=... -X main.channel=local"`.
- Origin нормализован: `https`, без userinfo, без `.git`, без хвостового `/`.
  Готовый case-блок уже в шаблоне justfile — не переписывай его руками.
- Рецепты `bump-patch` / `bump-minor` / `bump-major` есть и правят только `versions.txt`.
- Релизные workflow, если нужны релизы: `~/tools/install/workflows/release.yml` →
  `.github/workflows/release.yml`, `release-gitea.yml` → `.gitea/workflows/release.yml`.

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

Затем `just check` (vet + test), если такой рецепт есть.

## Что не делать

- Не коммить и не пушить без просьбы; `bump-*` — только по явному запросу.
- Не вшивать `git remote get-url origin` в бинарь без нормализации: в remote бывает токен.
- Не менять формат вывода флагов «для красоты» — это контракт с инсталляторами.
- Не добавлять комментарии в justfile, скрипты и Go-код.
