#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GWT_BIN="$ROOT_DIR/gwt"
WORK_DIR="$(mktemp -d /tmp/gwt-smoke-XXXXXX)"
UPSTREAM_DIR="$WORK_DIR/upstream-src"
BARE_REMOTE="$WORK_DIR/remote.git"

cleanup() {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

log() {
  printf '[smoke] %s\n' "$*"
}

fail() {
  printf '[smoke] FAIL: %s\n' "$*" >&2
  exit 1
}

assert_exists() {
  local path="$1"
  [[ -e "$path" ]] || fail "Expected path to exist: $path"
}

assert_not_exists() {
  local path="$1"
  [[ ! -e "$path" ]] || fail "Expected path to be missing: $path"
}

assert_contains() {
  local needle="$1"
  local haystack="$2"
  [[ "$haystack" == *"$needle"* ]] || fail "Expected output to contain '$needle'"
}

log "Building gwt binary"
(
  cd "$ROOT_DIR"
  GOCACHE=/tmp/gocache-gwt-smoke go build -o "$GWT_BIN" .
)

log "Creating local upstream repo"
mkdir -p "$UPSTREAM_DIR"
(
  cd "$UPSTREAM_DIR"
  git init -b main >/dev/null
  git config user.email "smoke@example.com"
  git config user.name "Smoke Test"
  echo "hello" > README.md
  git add README.md
  git commit -m "initial commit" >/dev/null
  git checkout -b feature/existing >/dev/null
  echo "feature" > feature.txt
  git add feature.txt
  git commit -m "existing feature" >/dev/null
  git checkout main >/dev/null
)

git clone --bare "$UPSTREAM_DIR" "$BARE_REMOTE" >/dev/null

CLONE_TARGET="$WORK_DIR/workspace"
log "Running gwt clone"
"$GWT_BIN" clone "$BARE_REMOTE" "$CLONE_TARGET" --branch main || fail "gwt clone failed"

assert_exists "$CLONE_TARGET/.bare"
assert_exists "$CLONE_TARGET/.git"
assert_exists "$CLONE_TARGET/main"

log "Running gwt checkout for new branch"
(
  cd "$CLONE_TARGET"
  "$GWT_BIN" checkout feature/smoke >/dev/null
)
assert_exists "$CLONE_TARGET/feature/smoke"

log "Running gwt list"
LIST_OUTPUT="$(cd "$CLONE_TARGET" && "$GWT_BIN" list 2>&1)"
assert_contains "feature/smoke" "$LIST_OUTPUT"
assert_contains "main" "$LIST_OUTPUT"

log "Running gwt remove"
(
  cd "$CLONE_TARGET"
  "$GWT_BIN" remove feature/smoke --force >/dev/null
)
assert_not_exists "$CLONE_TARGET/feature/smoke"

log "Smoke test passed"
