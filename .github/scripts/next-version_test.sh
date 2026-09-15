#!/usr/bin/env bash
# Table test for next-version.sh. Run: bash .github/scripts/next-version_test.sh
set -euo pipefail

script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/next-version.sh"
failures=0

# check <name> <tags, newline-separated> <bump> <want output> <want exit code>
check() {
  local name="$1" tags="$2" bump="$3" want="$4" want_code="$5"
  local got code=0
  got=$(printf '%s' "$tags" | bash "$script" "$bump" 2>/dev/null) || code=$?
  if [ "$got" != "$want" ] || [ "$code" != "$want_code" ]; then
    echo "FAIL $name: got '$got' (exit $code), want '$want' (exit $want_code)"
    failures=$((failures + 1))
  else
    echo "ok   $name"
  fi
}

check "patch bump" $'v0.1.12\nv0.1.13' patch v0.1.14 0
check "minor bump resets patch" $'v0.1.12\nv0.1.13' minor v0.2.0 0
check "numeric not lexical order" $'v0.1.9\nv0.1.13\nv0.1.10' patch v0.1.14 0
check "major order" $'v1.0.0\nv0.9.9' minor v1.1.0 0
check "pre-release tags ignored" $'v0.1.13\nv0.2.0-rc.1' patch v0.1.14 0
check "non-semver tags ignored" $'latest\nv1\nv0.1.13\nrelease-2026' patch v0.1.14 0
check "no tags starts at v0.1.0" '' patch v0.1.0 0
check "no stable tags starts at v0.1.0" $'v0.1.0-rc.1' minor v0.1.0 0
check "unknown bump fails" $'v0.1.13' major '' 2
check "missing bump fails" $'v0.1.13' '' '' 2

if [ "$failures" -gt 0 ]; then
  echo "$failures case(s) failed"
  exit 1
fi
echo "all cases passed"
