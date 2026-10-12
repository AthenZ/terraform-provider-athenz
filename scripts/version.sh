#!/usr/bin/env bash
# Print provider tags used by the certify workflow.
#
#   scripts/version.sh show         latest stable tag (vX.Y.Z)
#   scripts/version.sh prerelease   vX.Y.Z-<shortsha> for HEAD, without bumping patch
#   scripts/version.sh patch        next stable tag after the latest stable tag
set -euo pipefail

usage() {
  echo "usage: $0 show|prerelease|patch" >&2
  exit 2
}

latest_stable() {
  git tag -l 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 || true
}

[[ $# -eq 1 ]] || usage
cd "$(git rev-parse --show-toplevel)"

stable="$(latest_stable)"
if [[ -z "${stable}" ]]; then
  echo "no stable vX.Y.Z tags found" >&2
  exit 1
fi

case "$1" in
  show)
    printf '%s\n' "${stable}"
    ;;
  prerelease)
    sha="$(git rev-parse --short=7 HEAD)"
    printf '%s-%s\n' "${stable}" "${sha}"
    ;;
  patch)
    ver="${stable#v}"
    major="${ver%%.*}"
    rest="${ver#*.}"
    minor="${rest%%.*}"
    patch="${rest#*.}"
    printf 'v%s.%s.%s\n' "${major}" "${minor}" "$((patch + 1))"
    ;;
  *)
    usage
    ;;
esac
