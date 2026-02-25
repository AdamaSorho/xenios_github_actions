#!/usr/bin/env bash
# =============================================================================
# pre-pr-check.sh — Run this before opening a PR
#
# Mirrors the CI checks in deploy-backend.yml, deploy-web.yml, deploy-mobile.yml
# and tdd-gate.yml so failures are caught locally before push.
#
# Usage:
#   ./scripts/pre-pr-check.sh              # Auto-detect changed apps
#   ./scripts/pre-pr-check.sh --all        # Check all apps
#   ./scripts/pre-pr-check.sh --backend    # Check backend only
#   ./scripts/pre-pr-check.sh --web        # Check web only
#   ./scripts/pre-pr-check.sh --mobile     # Check mobile only
# =============================================================================

set -euo pipefail

# ── Colors ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────────────────────
pass() { echo -e "  ${GREEN}✓${RESET} $1"; }
fail() { echo -e "  ${RED}✗${RESET} $1"; FAILURES=$((FAILURES + 1)); }
info() { echo -e "  ${BLUE}→${RESET} $1"; }
header() { echo -e "\n${BOLD}${BLUE}══ $1 ══${RESET}"; }
warn() { echo -e "  ${YELLOW}⚠${RESET}  $1"; }

FAILURES=0
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# ── Argument Parsing ──────────────────────────────────────────────────────────
CHECK_BACKEND=false
CHECK_WEB=false
CHECK_MOBILE=false
EXPLICIT=false

for arg in "$@"; do
  case $arg in
    --all)     CHECK_BACKEND=true; CHECK_WEB=true; CHECK_MOBILE=true; EXPLICIT=true ;;
    --backend) CHECK_BACKEND=true; EXPLICIT=true ;;
    --web)     CHECK_WEB=true; EXPLICIT=true ;;
    --mobile)  CHECK_MOBILE=true; EXPLICIT=true ;;
    *) echo "Unknown argument: $arg"; exit 1 ;;
  esac
done

# ── Auto-detect changed apps from git diff ────────────────────────────────────
if [ "$EXPLICIT" = false ]; then
  BASE_BRANCH="${BASE_BRANCH:-main}"
  CHANGED=$(git diff --name-only "origin/$BASE_BRANCH" 2>/dev/null || git diff --name-only HEAD~1 2>/dev/null || echo "")

  if echo "$CHANGED" | grep -q "^apps/backend/\|^packages/"; then
    CHECK_BACKEND=true
  fi
  if echo "$CHANGED" | grep -q "^apps/web/\|^packages/"; then
    CHECK_WEB=true
  fi
  if echo "$CHANGED" | grep -q "^apps/mobile/\|^packages/"; then
    CHECK_MOBILE=true
  fi

  # If nothing detected (e.g. new branch with no commits), check all
  if [ "$CHECK_BACKEND" = false ] && [ "$CHECK_WEB" = false ] && [ "$CHECK_MOBILE" = false ]; then
    warn "Could not detect changed apps. Running all checks."
    CHECK_BACKEND=true; CHECK_WEB=true; CHECK_MOBILE=true
  fi
fi

echo -e "${BOLD}╔══════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}║        Xenios Pre-PR Validation          ║${RESET}"
echo -e "${BOLD}╚══════════════════════════════════════════╝${RESET}"
echo ""
info "Checking: $([ "$CHECK_BACKEND" = true ] && echo 'backend ') $([ "$CHECK_WEB" = true ] && echo 'web ') $([ "$CHECK_MOBILE" = true ] && echo 'mobile')"

# =============================================================================
# BACKEND CHECKS (Go)
# =============================================================================
check_backend() {
  header "Backend (Go)"
  cd "$ROOT/apps/backend"

  # 1. Compilation
  info "Building..."
  if go build ./... 2>&1; then
    pass "Compilation"
  else
    fail "Compilation — fix build errors before continuing"
    return
  fi

  # 2. Clean Architecture checks (mirrors tdd-gate.yml)
  info "Checking Clean Architecture..."
  ARCH_VIOLATIONS=0

  if grep -r "infrastructure\|adapter" internal/domain/ 2>/dev/null | grep -v "_test.go" | grep -q .; then
    fail "Architecture: domain/ imports from outer layers"
    ARCH_VIOLATIONS=$((ARCH_VIOLATIONS + 1))
  fi

  if grep -r "infrastructure\|adapter/handler" internal/usecase/ 2>/dev/null | grep -v "_test.go" | grep -q .; then
    fail "Architecture: usecase/ imports infrastructure or handlers"
    ARCH_VIOLATIONS=$((ARCH_VIOLATIONS + 1))
  fi

  if grep -rE "github.com/jackc/pgx|database/sql" internal/domain/ internal/usecase/ 2>/dev/null | grep -v "_test.go" | grep -q .; then
    fail "Architecture: database imports found outside infrastructure layer"
    ARCH_VIOLATIONS=$((ARCH_VIOLATIONS + 1))
  fi

  if grep -rE "gorm.io|entgo.io|github.com/uptrace/bun|volatiletech/sqlboiler" --include="*.go" . 2>/dev/null | grep -q .; then
    fail "Architecture: ORM detected — use raw SQL with pgx only"
    ARCH_VIOLATIONS=$((ARCH_VIOLATIONS + 1))
  fi

  if [ "$ARCH_VIOLATIONS" -eq 0 ]; then
    pass "Clean Architecture"
  fi

  # 3. Tests + Coverage
  info "Running tests with coverage..."
  if go test -race -coverprofile=coverage.out ./... 2>&1; then
    pass "Tests"

    # Coverage threshold check
    TOTAL=$(go tool cover -func=coverage.out | grep "^total:" | awk '{print $3}' | sed 's/%//')
    if [ -z "$TOTAL" ]; then
      warn "Could not determine coverage percentage"
    else
      THRESHOLD=80
      if awk "BEGIN {exit !($TOTAL >= $THRESHOLD)}"; then
        pass "Coverage: ${TOTAL}% (≥ ${THRESHOLD}% required)"
      else
        fail "Coverage: ${TOTAL}% — must be ≥ ${THRESHOLD}%"
        info "Run: go tool cover -html=coverage.out  (to see uncovered lines)"
      fi
    fi
  else
    fail "Tests failed"
  fi

  # 4. Lint (optional — only if golangci-lint is installed)
  if command -v golangci-lint &>/dev/null; then
    info "Running linter..."
    if golangci-lint run ./... 2>&1; then
      pass "Lint"
    else
      fail "Lint — fix lint errors before PR"
    fi
  else
    warn "golangci-lint not installed — skipping lint (CI will still run it)"
    info "Install: https://golangci-lint.run/usage/install/"
  fi

  cd "$ROOT"
}

# =============================================================================
# WEB CHECKS (Next.js)
# =============================================================================
check_web() {
  header "Web (Next.js)"
  cd "$ROOT"

  # 1. TypeScript
  info "Running typecheck..."
  if npm run typecheck --workspace=apps/web 2>&1; then
    pass "TypeScript"
  else
    fail "TypeScript — fix type errors before PR"
  fi

  # 2. Lint
  info "Running lint..."
  if npm run lint --workspace=apps/web 2>&1; then
    pass "Lint"
  else
    fail "Lint — fix lint errors before PR"
  fi

  # 3. Tests + Coverage
  info "Running tests with coverage..."
  if npm run test --workspace=apps/web -- --coverage --passWithNoTests 2>&1; then
    pass "Tests"

    # Coverage report location
    LCOV="apps/web/coverage/lcov.info"
    if [ -f "$LCOV" ]; then
      COVERED=$(grep -c "^DA:" "$LCOV" 2>/dev/null || echo "0")
      TOTAL_LINES=$(grep "^DA:" "$LCOV" 2>/dev/null | wc -l || echo "0")
      info "Coverage report: apps/web/coverage/lcov.info"
    else
      warn "No coverage report found at apps/web/coverage/lcov.info"
      warn "Make sure your test script generates coverage (--coverage flag)"
    fi
  else
    fail "Tests failed"
  fi

  # 4. No database imports check
  info "Checking for forbidden database imports..."
  FORBIDDEN='@supabase/supabase-js|prisma|typeorm|drizzle-orm|sequelize|@mikro-orm|"pg"|"mysql"|"mongodb"'
  if grep -E "$FORBIDDEN" apps/web/package.json 2>/dev/null | grep -v "//"; then
    fail "Forbidden database library detected in web app"
  else
    pass "No forbidden database imports"
  fi
}

# =============================================================================
# MOBILE CHECKS (React Native)
# =============================================================================
check_mobile() {
  header "Mobile (React Native)"
  cd "$ROOT"

  # 1. TypeScript
  info "Running typecheck..."
  if npm run typecheck --workspace=apps/mobile 2>&1; then
    pass "TypeScript"
  else
    fail "TypeScript — fix type errors before PR"
  fi

  # 2. Lint
  info "Running lint..."
  if npm run lint --workspace=apps/mobile -- || true 2>&1; then
    pass "Lint"
  else
    fail "Lint — fix lint errors before PR"
  fi

  # 3. Tests + Coverage
  info "Running tests with coverage..."
  if npm run test --workspace=apps/mobile -- --coverage --passWithNoTests 2>&1; then
    pass "Tests"

    LCOV="apps/mobile/coverage/lcov.info"
    if [ -f "$LCOV" ]; then
      info "Coverage report: apps/mobile/coverage/lcov.info"
    else
      warn "No coverage report found at apps/mobile/coverage/lcov.info"
    fi
  else
    fail "Tests failed"
  fi

  # 4. No database imports check
  info "Checking for forbidden database imports..."
  FORBIDDEN='@supabase/supabase-js|prisma|typeorm|drizzle-orm|sequelize|@mikro-orm|"pg"|"mysql"|"mongodb"'
  if grep -E "$FORBIDDEN" apps/mobile/package.json 2>/dev/null | grep -v "//"; then
    fail "Forbidden database library detected in mobile app"
  else
    pass "No forbidden database imports"
  fi
}

# =============================================================================
# Run selected checks
# =============================================================================
[ "$CHECK_BACKEND" = true ] && check_backend
[ "$CHECK_WEB" = true ]     && check_web
[ "$CHECK_MOBILE" = true ]  && check_mobile

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${BOLD}══════════════════════════════════════════${RESET}"
if [ "$FAILURES" -eq 0 ]; then
  echo -e "${GREEN}${BOLD}  ✓ All checks passed — safe to open PR${RESET}"
else
  echo -e "${RED}${BOLD}  ✗ $FAILURES check(s) failed — fix before opening PR${RESET}"
  echo ""
  echo -e "  ${YELLOW}Tip:${RESET} Fix the issues above, then re-run this script."
fi
echo -e "${BOLD}══════════════════════════════════════════${RESET}"
echo ""

exit $FAILURES
