#!/usr/bin/env bash
set -euo pipefail

MODE="install"
GOROOT_TARGET="${GO_INSTALL_ROOT:-}"

usage() {
    cat <<USAGE
Usage: ensure-go.sh [--check] [--root DIR]

  --check      only report current/latest Go, exit 1 if outdated
  --root DIR   install prefix for GOROOT (default: current GOROOT or /usr/local/go)
USAGE
}

while [ $# -gt 0 ]; do
    case "$1" in
        --check) MODE="check"; shift ;;
        --root) GOROOT_TARGET="$2"; shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
    esac
done

need() {
    command -v "$1" >/dev/null 2>&1 || { echo "required tool not found: $1" >&2; exit 1; }
}

need curl
need tar

current_version() {
    if command -v go >/dev/null 2>&1; then
        go env GOVERSION 2>/dev/null || true
    fi
}

latest_json() {
    curl -fsSL --max-time 20 'https://go.dev/dl/?mode=json'
}

os_arch() {
    local os arch
    case "$(uname -s)" in
        Linux) os=linux ;;
        Darwin) os=darwin ;;
        *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64) arch=amd64 ;;
        aarch64|arm64) arch=arm64 ;;
        *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
    esac
    printf '%s %s\n' "$os" "$arch"
}

JSON="$(latest_json)"
read -r OS ARCH <<<"$(os_arch)"

if command -v jq >/dev/null 2>&1; then
    LATEST="$(printf '%s' "$JSON" | jq -r '[.[] | select(.stable)][0].version')"
    FILENAME="$(printf '%s' "$JSON" | jq -r --arg v "$LATEST" --arg os "$OS" --arg arch "$ARCH" \
        '[.[] | select(.version == $v)][0].files[] | select(.os == $os and .arch == $arch and .kind == "archive") | .filename')"
    SHA="$(printf '%s' "$JSON" | jq -r --arg f "$FILENAME" '.[].files[] | select(.filename == $f) | .sha256' | head -n1)"
else
    LATEST="$(printf '%s' "$JSON" | grep -o '"version": *"go[0-9][^"]*"' | head -n1 | sed 's/.*"\(go[0-9][^"]*\)"/\1/')"
    FILENAME="${LATEST}.${OS}-${ARCH}.tar.gz"
    SHA=""
fi

[ -n "$LATEST" ] || { echo "cannot determine latest stable Go version" >&2; exit 1; }
[ -n "$FILENAME" ] || FILENAME="${LATEST}.${OS}-${ARCH}.tar.gz"

CURRENT="$(current_version)"
echo "current: ${CURRENT:-none}"
echo "latest:  $LATEST"

if [ "$CURRENT" = "$LATEST" ]; then
    echo "up to date"
    exit 0
fi

if [ "$MODE" = "check" ]; then
    echo "outdated"
    exit 1
fi

if [ -z "$GOROOT_TARGET" ]; then
    if command -v go >/dev/null 2>&1; then
        GOROOT_TARGET="$(go env GOROOT)"
    else
        GOROOT_TARGET="/usr/local/go"
    fi
fi

PARENT="$(dirname "$GOROOT_TARGET")"
SUDO=""
if [ ! -w "$PARENT" ]; then
    command -v sudo >/dev/null 2>&1 || { echo "no write access to $PARENT and sudo not available" >&2; exit 1; }
    SUDO="sudo"
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

URL="https://go.dev/dl/${FILENAME}"
echo "downloading $URL"
curl -fsSL --max-time 600 -o "$TMP/$FILENAME" "$URL"

if [ -n "$SHA" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
        ACTUAL="$(sha256sum "$TMP/$FILENAME" | awk '{print $1}')"
    else
        ACTUAL="$(shasum -a 256 "$TMP/$FILENAME" | awk '{print $1}')"
    fi
    [ "$ACTUAL" = "$SHA" ] || { echo "sha256 mismatch: $ACTUAL != $SHA" >&2; exit 1; }
    echo "sha256 ok"
else
    echo "sha256 unavailable, skipping verification" >&2
fi

echo "installing into $GOROOT_TARGET"
$SUDO rm -rf "${GOROOT_TARGET}.old"
if [ -d "$GOROOT_TARGET" ]; then
    $SUDO mv "$GOROOT_TARGET" "${GOROOT_TARGET}.old"
fi
$SUDO tar -C "$TMP" -xzf "$TMP/$FILENAME"
$SUDO mv "$TMP/go" "$GOROOT_TARGET"
$SUDO rm -rf "${GOROOT_TARGET}.old"

"$GOROOT_TARGET/bin/go" version
