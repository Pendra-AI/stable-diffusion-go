#!/usr/bin/env bash
# inject-devmem-wrapper.sh — copy Pendra's sd_backend_dev_* device-memory wrapper
# into a freshly cloned upstream stable-diffusion.cpp tree so it compiles into
# libstable-diffusion.
#
# Upstream globs its library sources with `file(GLOB SD_LIB_SOURCES
# CONFIGURE_DEPENDS "src/*.cpp" ...)`, so dropping the file into src/ is enough —
# no CMakeLists edit, and it survives upstream refactors/target renames. Run
# right after scripts/clone-upstream.sh, before the cmake configure. Idempotent.
#
# Usage:
#   scripts/inject-devmem-wrapper.sh <upstream-dir>
set -euo pipefail

upstream="${1:?usage: inject-devmem-wrapper.sh <upstream-dir>}"
here="$(cd "$(dirname "$0")/.." && pwd)"
src="$here/csrc/sd_devmem.cpp"
dest="$upstream/src/sd_devmem.cpp"

if [ ! -f "$src" ]; then
  echo "inject-devmem-wrapper: wrapper source not found: $src" >&2
  exit 1
fi
if [ ! -d "$upstream/src" ]; then
  echo "inject-devmem-wrapper: no src/ directory in upstream tree: $upstream" >&2
  echo "  (has the upstream layout changed? the src/*.cpp glob may need revisiting)" >&2
  exit 1
fi

cp "$src" "$dest"
echo "inject-devmem-wrapper: installed $dest"
