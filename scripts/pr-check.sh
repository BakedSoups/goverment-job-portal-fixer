#!/bin/sh
set -eu

CDPATH=''
repo_root="$(cd -- "$(dirname -- "$0")/.." && pwd)"
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
  [ -d scripts ] || fail "missing scripts/"
  [ -d .githooks ] || fail "missing .githooks/"
  [ -f .githooks/pre-commit ] || fail "missing .githooks/pre-commit"

  if [ -d templates ] && [ ! -d templates/partials ]; then
    printf 'pr-check: warning: templates/ exists without templates/partials/\n' >&2
  fi

  if [ -d web ] && [ ! -d templates ]; then
    fail "web/ exists but templates/ is missing"
  fi
}

check_git_setup() {
  info "checking git setup"

  git rev-parse --is-inside-work-tree >/dev/null 2>&1 || fail "not inside a git work tree"

  if [ "${GITHUB_ACTIONS:-}" != "true" ]; then
    hooks_path="$(git config --get core.hooksPath || true)"
    if [ "$hooks_path" != ".githooks" ]; then
      info "configuring git hooks"
      git config core.hooksPath .githooks
    fi

    chmod +x .githooks/pre-commit scripts/pr-check.sh
  else
    info "running in GitHub Actions"
  fi

  branch="$(git branch --show-current)"
  if [ "${GITHUB_ACTIONS:-}" != "true" ]; then
    [ -n "$branch" ] || fail "detached HEAD; create or checkout a branch before opening a PR"
  fi

  remote_url="$(git remote get-url origin 2>/dev/null || true)"
  [ -n "$remote_url" ] || fail "missing origin remote"

  if [ "${GITHUB_ACTIONS:-}" != "true" ]; then
    upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)"
    [ -n "$upstream" ] || fail "current branch has no upstream; run: git push -u origin $branch"
  fi

  [ -x .githooks/pre-commit ] || fail ".githooks/pre-commit is not executable"
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
  unformatted="$(find . -path ./.git -prune -o -name '*.go' -exec gofmt -l {} +)"
  [ -z "$unformatted" ] || fail "gofmt needed for: $unformatted"

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
  [ -x scripts/pr-check.sh ] || fail "scripts/pr-check.sh is not executable"

  if command -v shellcheck >/dev/null 2>&1; then
    shellcheck scripts/pr-check.sh .githooks/pre-commit
  else
    printf 'pr-check: shellcheck not installed; skipping shell lint\n' >&2
  fi

  if [ -d deploy/zero ]; then
    command -v node >/dev/null 2>&1 || fail "deploy/zero requires Node.js for validation"
    node --input-type=module --check <deploy/zero/worker.js
    node --check deploy/zero/build-payload.mjs
  fi
}

check_required_layout
check_git_setup
check_no_generated_noise
check_go
check_templates
check_scripts

info "all checks passed"
