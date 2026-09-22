#!/usr/bin/env bash
# apply-upstream-patches.sh — apply the local fixes in patches/*.patch (in
# filename order) to a freshly cloned upstream stable-diffusion.cpp tree.
#
# These are small bug fixes to upstream carried until they land there. Each is a
# `git format-patch` against the commit pinned in lib/version.txt. Fail-closed:
# if a patch no longer applies (e.g. after a pin bump), the build stops so the
# patch is re-ported or dropped rather than silently skipped. Run right after
# scripts/clone-upstream.sh, before the cmake configure. Idempotent: a patch that
# is already applied is skipped.
#
# Usage:
#   scripts/apply-upstream-patches.sh <upstream-dir>
set -euo pipefail

upstream="${1:?usage: apply-upstream-patches.sh <upstream-dir>}"
here="$(cd "$(dirname "$0")/.." && pwd)"

shopt -s nullglob
patches=("$here"/patches/*.patch)
if [ ${#patches[@]} -eq 0 ]; then
  echo "apply-upstream-patches: no patches to apply"
  exit 0
fi

for p in "${patches[@]}"; do
  name="$(basename "$p")"
  if git -C "$upstream" apply --reverse --check "$p" 2>/dev/null; then
    echo "apply-upstream-patches: $name already applied, skipping"
    continue
  fi
  if ! git -C "$upstream" apply --check "$p"; then
    echo "apply-upstream-patches: $name does not apply to $(git -C "$upstream" rev-parse --short HEAD)" >&2
    echo "  (re-port it against the pinned upstream commit, or drop it if upstream has fixed the issue)" >&2
    exit 1
  fi
  git -C "$upstream" apply "$p"
  echo "apply-upstream-patches: applied $name"
done
