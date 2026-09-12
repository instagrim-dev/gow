#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/build-paper.sh [--outdir DIR] [--no-install]

Build paper/geometry-of-work.tex with Tectonic.

If tectonic is not already on PATH, this script installs it into the
repo-local .cache/tools/tectonic directory using Tectonic's official Unix
installer. Set TECTONIC_TOOL_DIR to override that location. Use --no-install
to require an already-installed tectonic binary.
USAGE
}

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
outdir="${PAPER_OUTDIR:-paper}"
allow_install=1

while [ "$#" -gt 0 ]; do
  case "$1" in
    --outdir)
      if [ "$#" -lt 2 ]; then
        echo "--outdir requires a directory" >&2
        exit 2
      fi
      outdir="$2"
      shift 2
      ;;
    --no-install)
      allow_install=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$outdir" in
  /*) ;;
  *) outdir="$root/$outdir" ;;
esac

tool_dir="${TECTONIC_TOOL_DIR:-$root/.cache/tools/tectonic}"
cached_tectonic="$tool_dir/tectonic"

find_tectonic() {
  if command -v tectonic >/dev/null 2>&1; then
    command -v tectonic
    return
  fi
  if [ -x "$cached_tectonic" ]; then
    printf '%s\n' "$cached_tectonic"
    return
  fi
  return 1
}

install_tectonic() {
  if [ "$allow_install" -ne 1 ]; then
    cat >&2 <<'MSG'
tectonic is required to build paper/geometry-of-work.tex, but it is not on PATH.
Install tectonic or rerun without --no-install to allow repo-local installation.
MSG
    return 1
  fi
  if ! command -v curl >/dev/null 2>&1; then
    cat >&2 <<'MSG'
tectonic is required to build paper/geometry-of-work.tex, and curl is unavailable.
Install tectonic manually, or run this target in CI/a dev image that has curl.
MSG
    return 1
  fi

  mkdir -p "$(dirname "$cached_tectonic")"
  for attempt in 1 2 3; do
    if (
      cd "$(dirname "$cached_tectonic")"
      curl --proto '=https' --tlsv1.2 -fsSL https://drop-sh.fullyjustified.net | sh
    ); then
      break
    fi
    if [ "$attempt" -eq 3 ]; then
      echo "failed to install tectonic after 3 attempts" >&2
      return 1
    fi
    echo "tectonic install attempt $attempt failed; retrying" >&2
    sleep $((attempt * 2))
  done
  chmod +x "$cached_tectonic"
}

tectonic_bin="$(find_tectonic || true)"
if [ -z "$tectonic_bin" ]; then
  install_tectonic
  tectonic_bin="$(find_tectonic)"
fi

mkdir -p "$outdir"
export XDG_CACHE_HOME="${XDG_CACHE_HOME:-$root/.cache}"

(
  cd "$root/paper"
  "$tectonic_bin" --keep-logs --outdir "$outdir" geometry-of-work.tex
)
