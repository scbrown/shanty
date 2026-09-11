#!/usr/bin/env bash
# Refuse a feat/fix that does not touch CHANGELOG.md.
#
# WHY THIS IS RANGE-SCOPED AND NOT PULL-REQUEST-SCOPED. The obvious shape for
# this guard is "on pull_request, did the PR touch CHANGELOG.md". In this repo
# that guard would almost never run: the last 30 commits contain ONE merge, so
# work lands by pushing to main. A guard wired only to pull_request would read as
# wired and fire on nothing -- which is how three released-to-the-fleet commits
# reached main unchangelogged in the first place.
#
# So it takes a COMMIT RANGE and CI passes it the push range as well as the PR
# range.
#
# The opt-out is a trailer, not a scope allowlist. A list of "scopes that are not
# user visible" rots the moment somebody invents a scope, and a guard that cannot
# be satisfied honestly gets deleted rather than obeyed. `Changelog: skip` in the
# commit body is explicit, greppable, and shows up in review.
set -uo pipefail

usage() { echo "usage: $0 <base> <head>   |   $0 --selftest" >&2; exit 2; }

check() {
  local base="$1" head="$2" subjects touched offenders=()
  # An unresolvable base (first push, force push, shallow clone) must not be read
  # as an empty range -- that would pass everything. Fall back to the head commit
  # alone, which is strictly safer than checking nothing.
  if ! git rev-parse --verify -q "$base^{commit}" >/dev/null 2>&1; then
    base="$head~1"
    git rev-parse --verify -q "$base^{commit}" >/dev/null 2>&1 || base="$head"
  fi
  local range="$base..$head"
  [ "$base" = "$head" ] && range="$head -1"

  touched="$(git log --format=%H --name-only $range -- CHANGELOG.md 2>/dev/null | head -1)"

  while IFS= read -r sha; do
    [ -n "$sha" ] || continue
    local subj body
    subj="$(git log -1 --format=%s "$sha")"
    body="$(git log -1 --format=%B "$sha")"
    printf '%s' "$subj" | grep -qE '^(feat|fix)(\([^)]*\))?!?:' || continue
    printf '%s' "$body" | grep -qiE '^Changelog:[[:space:]]*skip' && continue
    offenders+=("$sha ${subj}")
  done < <(git log --format=%H $range 2>/dev/null)

  if [ "${#offenders[@]}" -eq 0 ]; then
    echo "ok: no feat/fix commit in $range needs a changelog entry"
    return 0
  fi
  if [ -n "$touched" ]; then
    echo "ok: ${#offenders[@]} feat/fix commit(s) in $range, and CHANGELOG.md was updated"
    return 0
  fi
  echo "FAIL: feat/fix landed with no CHANGELOG.md entry in $range:"
  printf '  %s\n' "${offenders[@]}"
  echo
  echo "This repo logs features and fixes under '## [Unreleased]' as they land, and"
  echo "the release commit rolls that section into the new version. Add an entry, or"
  echo "put 'Changelog: skip' in the commit body if the change is genuinely not"
  echo "user-visible."
  return 1
}

selftest() {
  local t rc fails=0
  t="$(mktemp -d)"; trap 'rm -rf "$t"' RETURN
  git init -q "$t"; cd "$t" || return 1
  git config user.email t@t; git config user.name t
  echo "# Changelog" > CHANGELOG.md; echo x > f
  git add -A; git commit -qm "chore: base"
  local base; base=$(git rev-parse HEAD)

  _arm() { # name expected_rc
    local name="$1" want="$2"; shift 2
    check "$base" "$(git rev-parse HEAD)" >/dev/null 2>&1; rc=$?
    if [ "$rc" = "$want" ]; then echo "  ok   $name (rc=$rc)"; else echo "  FAIL $name: want rc=$want got $rc"; fails=$((fails+1)); fi
  }

  echo y > f; git commit -qam "feat(x): a user-visible thing"
  _arm "feat WITHOUT changelog must FAIL" 1

  echo "- entry" >> CHANGELOG.md; git add CHANGELOG.md; git commit -qm "docs: note it"
  _arm "same range, changelog now touched, must PASS" 0

  git checkout -q -b t2 "$base"
  echo z > f; git commit -qam "docs(readme): wording only"
  _arm "docs-only must PASS" 0

  git checkout -q -b t3 "$base"
  echo w > f; git commit -qam "fix(ci): not user visible

Changelog: skip"
  _arm "fix with 'Changelog: skip' trailer must PASS" 0

  git checkout -q -b t4 "$base"
  echo v > f; git commit -qam "fix(bar): a real user-visible fix"
  _arm "fix WITHOUT changelog must FAIL" 1

  # An unresolvable base must not silently pass everything.
  check "0000000000000000000000000000000000000000" "$(git rev-parse HEAD)" >/dev/null 2>&1; rc=$?
  if [ "$rc" = 1 ]; then echo "  ok   unresolvable base falls back to HEAD and still FAILS (rc=1)"; else echo "  FAIL unresolvable base: want rc=1 got $rc"; fails=$((fails+1)); fi

  echo "---"
  [ "$fails" -eq 0 ] && { echo "PASS: all arms"; return 0; } || { echo "FAILED: $fails arm(s)"; return 1; }
}

case "${1:-}" in
  --selftest) selftest ;;
  "" ) usage ;;
  * ) [ $# -eq 2 ] || usage; check "$1" "$2" ;;
esac
