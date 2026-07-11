#!/bin/bash
# Pre-push verification for Relay
# Run this before every push to catch CI failures locally.
# Usage: ./scripts/pre-push.sh
# Or: bash scripts/pre-push.sh
#
# Checks: gofmt, go vet, go test, unused imports, undefined refs

set -e

cd "$(dirname "$0")/.."

export CGO_ENABLED=0
export GOCACHE="${GOCACHE:-$HOME/.cache/go-build}"
export TMPDIR="${TMPDIR:-$HOME/.cache}"
mkdir -p "$GOCACHE" "$TMPDIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

fail_count=0

check() {
    local name="$1"
    local cmd="$2"
    echo "[*] $name"
    if eval "$cmd"; then
        echo "  -> pass"
    else
        echo "  -> FAIL"
        fail_count=$((fail_count + 1))
    fi
}

echo "=== Relay pre-push checks ==="
echo ""

# 1. gofmt - check all .go files
check "gofmt" \
    'out=$(gofmt -l .); if [ -n "$out" ]; then echo "  unformatted files:"; echo "$out"; false; fi'

# 2. go vet - catches undefined refs, unused imports, suspicious constructs
check "go vet" \
    'go vet ./...'

# 3. go test - run all tests
check "go test" \
    'go test -count=1 -timeout 120s ./...'

# 4. go build - make sure binary compiles
check "go build" \
    'go build -o /dev/null .'

# 5. Check for unused imports (go vet catches most, but double check)
check "unused imports" \
    'if go vet ./... 2>&1 | grep -q "imported and not used"; then false; else true; fi'

# 6. Check that test assertions reference real symbols
# (catches the doctorMarker issue where test called a standalone func but it was a method)
check "test symbol refs" \
    'if go vet ./... 2>&1 | grep -q "undefined"; then false; else true; fi'

echo ""
if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}All checks passed. Safe to push.${NC}"
    exit 0
else
    echo -e "${RED}$fail_count check(s) failed. Fix before pushing.${NC}"
    exit 1
fi
