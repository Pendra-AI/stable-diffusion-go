#!/usr/bin/env bash
# test-upstream-patches.sh — compile and run the regression tests for the
# carried upstream patches (csrc/test/*_test.cpp) against an upstream tree that
# scripts/apply-upstream-patches.sh has already patched. Needs only a C++17
# compiler: the tests link the patched upstream sources they cover directly, not
# ggml or a GPU backend, and use no model files.
#
# Usage:
#   scripts/test-upstream-patches.sh <upstream-dir>
set -euo pipefail

upstream="${1:?usage: test-upstream-patches.sh <upstream-dir>}"
here="$(cd "$(dirname "$0")/.." && pwd)"
cxx="${CXX:-c++}"
out="$(mktemp -d)"
trap 'rm -rf "$out"' EXIT

includes=(-I"$upstream/src" -I"$upstream/include" -I"$upstream/ggml/include")

"$cxx" -std=c++17 "${includes[@]}" -o "$out/name_conversion_test" \
  "$here/csrc/test/name_conversion_test.cpp" "$upstream/src/name_conversion.cpp"
"$out/name_conversion_test"
