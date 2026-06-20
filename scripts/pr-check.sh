#!/bin/sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

fail() {
  printf 'pr-check: %s\n' "$1" >&2
  exit 1
}

info() {
  printf '==> %s\n' "$1"
}

check_required_layout() {
  info "checking project layout"

  [ -f AGENTS.md ] || fail "missing AGENTS.md"

  if [ -d templates ] && [ ! -d templates/partials ]; then
    printf 'pr-check: warning: templates/ exists without templates/partials/\n' >&2
  fi

  if [ -d web ] && [ ! -d templates ]; then
    fail "web/ exists but templates/ is missing"
  fi
}

check_no_generated_noise() {
  info "checking for local noise"

  if find . \
    -path ./.git -prune -o \
    \( -name '.DS_Store' -o -name '*.tmp' -o -name '*.bak' -o -name '*.orig' \) \
    -print | grep -q .; then
    fail "found local noise files such as .DS_Store, *.tmp, *.bak, or *.orig"
  fi
}

check_go() {
  if [ ! -f go.mod ]; then
    info "go.mod not found; skipping Go checks"
    return
  fi

  info "checking gofmt"
  go_files="$(find . -path ./.git -prune -o -name '*.go' -print)"
  if [ -n "$go_files" ]; then
    unformatted="$(gofmt -l $go_files)"
    [ -z "$unformatted" ] || fail "gofmt needed for: $unformatted"
  fi

  info "running go test"
  go test ./...
}

check_templates() {
  if [ ! -d templates ]; then
    info "templates/ not found; skipping template checks"
    return
  fi

  info "checking templates"

  find templates -type f | while IFS= read -r file; do
    case "$file" in
      *.html) ;;
      *) fail "non-html file found in templates/: $file" ;;
    esac
  done

  if grep -R "SmartRecruiters API" templates >/dev/null 2>&1; then
    fail "templates should not contain scraper/API implementation details"
  fi
}

check_scripts() {
  info "checking scripts"

  [ -f scripts/pr-check.sh ] || fail "missing scripts/pr-check.sh"

  if command -v shellcheck >/dev/null 2>&1; then
    shellcheck scripts/pr-check.sh
  else
    printf 'pr-check: shellcheck not installed; skipping shell lint\n' >&2
  fi
}

check_required_layout
check_no_generated_noise
check_go
check_templates
check_scripts

info "all checks passed"
