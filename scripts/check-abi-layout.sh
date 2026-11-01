#!/usr/bin/env bash
# check-abi-layout.sh — compile csrc/test/abi_layout.c against an upstream tree
# and compare what it prints with pkg/sd/testdata/abi_layout.txt, the file the Go
# tests hold the binding's structs and enums against. A mismatch means the header
# changed and the binding has not followed it. Needs only a C11 compiler.
#
# Layouts are the same on every 64-bit target the libraries are built for, so the
# output is stable across linux/amd64, linux/arm64 and darwin/arm64.
#
# Usage:
#   scripts/check-abi-layout.sh <upstream-dir> [--update]
set -euo pipefail

upstream="${1:?usage: check-abi-layout.sh <upstream-dir> [--update]}"
mode="${2:-}"
here="$(cd "$(dirname "$0")/.." && pwd)"
golden="$here/pkg/sd/testdata/abi_layout.txt"
cc="${CC:-cc}"
out="$(mktemp -d)"
trap 'rm -rf "$out"' EXIT

"$cc" -std=c11 -I"$upstream/include" -o "$out/abi_layout" "$here/csrc/test/abi_layout.c"
"$out/abi_layout" > "$out/abi_layout.txt"

if [ "$mode" = "--update" ]; then
  mkdir -p "$(dirname "$golden")"
  cp "$out/abi_layout.txt" "$golden"
  echo "check-abi-layout: updated $golden"
  exit 0
fi

if ! diff -u "$golden" "$out/abi_layout.txt"; then
  echo "check-abi-layout: the header no longer matches pkg/sd/testdata/abi_layout.txt" >&2
  echo "  (port the binding, then rerun with --update)" >&2
  exit 1
fi
echo "check-abi-layout: OK"
