#!/usr/bin/env bash
# mentions-tiers.sh — (re)generates THIRD-PARTY-NOTICES.md: the third-party
# components shipped in MailStone Verifier (Go modules compiled into the four
# release binaries, Go standard library), their licence, their source and the
# text of each licence.
#
#   scripts/mentions-tiers.sh             regenerates THIRD-PARTY-NOTICES.md
#   scripts/mentions-tiers.sh --verifier  checks the file is up to date (exit code 1 otherwise)
#
# The module list is the set of packages actually compiled for each release
# platform, with the Wails build tags and cgo setting of release.yml
# (`go list -deps`): a module present in go.mod but never compiled is not
# distributed. Fails when a module has no recognised licence file, or carries a
# strong copyleft licence (GPL, LGPL, AGPL).
#
# Requires Go and access to the Go module proxy (module sources are
# downloaded). Neither GTK nor Wails: nothing is compiled.
set -euo pipefail
export LC_COLLATE=C LC_ALL=C   # stable order from one machine to another
cd "$(dirname "$0")/.."
OUT=THIRD-PARTY-NOTICES.md
VERIFY=0
[ "${1:-}" = "--verifier" ] && VERIFY=1

T=$(mktemp -d); trap 'rm -rf "$T"' EXIT

echo "== Compiled modules (Linux, Windows, macOS)" >&2
# cgo as in the real build: on for Linux (WebKitGTK) and macOS (Cocoa), off for the cross-compiled Windows binary.
# Left unset, go list turns cgo off by itself for a cross target and misses the modules of Wails' cgo files.
for target in "linux/amd64 desktop,production,webkit2_41 1" "windows/amd64 desktop,production 0" \
              "darwin/amd64 desktop,production 1" "darwin/arm64 desktop,production 1"; do
  set -- $target
  CGO_ENABLED=$3 GOOS=${1%/*} GOARCH=${1#*/} GOFLAGS="-mod=mod -buildvcs=false" go list -tags "$2" -deps \
    -f '{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}}{{end}}{{end}}' . \
    | sort -u | sed "s|\$| $1|"
done | sort | awk '{ k=$1" "$2; if (k in p) p[k]=p[k]", "$3; else p[k]=$3 } END { for (k in p) print k, p[k] }' \
  | sort > "$T/mods"

licence_of() {   # $1 = file → SPDX identifier, or "?"
  local f=$1
  if   grep -qi "Mozilla Public License" "$f"; then echo MPL-2.0
  elif grep -qiE "GNU (AFFERO|LESSER|LIBRARY)? ?GENERAL PUBLIC LICENSE" "$f"; then echo GPL
  elif grep -qi "Apache License" "$f"; then echo Apache-2.0
  elif grep -qi "Permission is hereby granted, free of charge" "$f"; then echo MIT
  elif grep -qi "Neither the name" "$f" || grep -qiE "names of (its|the) contributors may be used" "$f"; then echo BSD-3-Clause
  elif grep -qi "Redistribution and use in source and binary forms" "$f"; then echo BSD-2-Clause
  elif grep -qiE "Permission to use, copy, modify, and(/or)? distribute" "$f"; then echo ISC
  elif grep -qiE "dedicated to the public domain|Is Public Domain" "$f"; then echo Public-Domain
  else echo "?"; fi
}
escape() { echo "$1" | sed 's/[A-Z]/!&/g' | tr 'A-Z' 'a-z'; }   # Go proxy paths
block() {   # $1 = file: its text in a code block
  echo; echo "\`$(basename "$1")\`:"; echo; echo '````text'; sed 's/[[:space:]]*$//' "$1"; echo '````'
}

echo "== Licences of $(wc -l < "$T/mods") modules" >&2
ERRORS=0
: > "$T/table"; : > "$T/texts"
while read -r path ver platforms; do
  platforms=$(echo "$platforms" | sed 's/,$//')
  dir=$(go mod download -json "$path@$ver" | sed -n 's/^[[:space:]]*"Dir": "\(.*\)",$/\1/p')
  [ -n "$dir" ] && [ -d "$dir" ] || { echo "❌ $path $ver: sources not found" >&2; ERRORS=1; continue; }
  mapfile -t files < <(find "$dir" -maxdepth 1 -type f \( -iname 'LICEN[CS]E*' -o -iname 'COPYING*' -o -iname 'NOTICE*' -o -iname 'COPYRIGHT*' -o -iname 'PATENTS*' \) | sort)
  [ ${#files[@]} -gt 0 ] || { echo "❌ $path $ver: no licence file" >&2; ERRORS=1; continue; }
  spdx=""
  for f in "${files[@]}"; do
    l=$(licence_of "$f")
    [ "$l" = GPL ] && { echo "❌ $path $ver: strong copyleft licence ($(basename "$f"))" >&2; ERRORS=1; }
    # NOTICE, COPYRIGHT and PATENTS accompany a licence without being one: reproduced, not classified.
    case $(basename "$f") in NOTICE*|notice*|COPYRIGHT*|copyright*|PATENTS*|LICENSE-LOGO*|LICENSE-3RD-PARTY*) [ "$l" = "?" ] && continue ;; esac
    [ "$l" = "?" ] && { echo "❌ $path $ver: licence not recognised ($(basename "$f"))" >&2; ERRORS=1; }
    case " $spdx " in *" $l "*) ;; *) spdx="${spdx:+$spdx }$l" ;; esac
  done
  [ -n "$spdx" ] || { echo "❌ $path $ver: no recognised licence" >&2; ERRORS=1; }
  spdx=$(echo "$spdx" | sed 's/ /, /g')
  echo "| \`$path\` | $ver | $spdx | $platforms |" >> "$T/table"
  {
    echo; echo "### \`$path\` $ver"; echo
    echo "Licence(s): $spdx. Platforms: $platforms. Source: <https://proxy.golang.org/$(escape "$path")/@v/$ver.zip>"
    case "$spdx" in *MPL-2.0*)
      echo; echo "Files under the Mozilla Public License 2.0 are distributed **unmodified**:"
      echo "their source form is available at the address above." ;;
    esac
    for f in "${files[@]}"; do block "$f"; done
  } >> "$T/texts"
done < "$T/mods"

[ "$ERRORS" = 0 ] || { echo "❌ notices not generated: fix the issues above" >&2; exit 1; }

{
  cat <<'EOF'
# Third-party components — MailStone Verifier

Generated by `scripts/mentions-tiers.sh`: do not edit by hand.

MailStone Verifier is distributed under the MIT licence (see `LICENSE`). It
includes the third-party components below, each under its own licence, which
this notice does not restrict in any way. This file is embedded in the
application (About, "Licence and third-party components") and attached to
each release.

## Go modules

Only the modules actually compiled into the release binaries are listed
(Linux amd64, Windows amd64, macOS amd64 and arm64).

| Module | Version | Licence(s) | Platforms |
|---|---|---|---|
EOF
  cat "$T/table"
  cat <<'EOF'

The Go standard library and toolchain are under the BSD-3-Clause licence
(text reproduced below).

## System libraries, not shipped

On Linux the interface is rendered by WebKitGTK and GTK, installed by the
distribution; on Windows by the Microsoft Edge WebView2 runtime, a system
component; on macOS by the system WebKit. MailStone Verifier does not
redistribute them.

## Licence texts

### Go standard library
EOF
  block "$(go env GOROOT)/LICENSE"
  cat "$T/texts"
} > "$T/out"

if [ "$VERIFY" = 1 ]; then
  if cmp -s "$T/out" "$OUT"; then echo "✅ $OUT up to date ($(wc -l < "$T/mods") modules)"; exit 0; fi
  echo "❌ $OUT is out of date: run scripts/mentions-tiers.sh" >&2
  diff <(grep '^| ' "$OUT" 2>/dev/null) <(grep '^| ' "$T/out") >&2 || true
  exit 1
fi
cp "$T/out" "$OUT"
echo "✅ $OUT regenerated ($(wc -l < "$T/mods") modules)"
