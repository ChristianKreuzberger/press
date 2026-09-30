#!/usr/bin/env bash
# Tests install.sh against a local fake release directory (file:// URLs).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

case "$(uname -s)" in Darwin*) os=darwin ;; *) os=linux ;; esac
case "$(uname -m)" in aarch64|arm64) arch=arm64 ;; *) arch=amd64 ;; esac
name="press_1.2.3_${os}_${arch}.tar.gz"

rel="${work}/release/v1.2.3"
mkdir -p "${rel}" "${work}/pkg"
printf '#!/bin/sh\necho fake-press\n' > "${work}/pkg/press"
chmod +x "${work}/pkg/press"
tar -czf "${rel}/${name}" -C "${work}/pkg" press
if command -v sha256sum >/dev/null 2>&1; then
  (cd "${rel}" && sha256sum "${name}" > checksums.txt)
else
  (cd "${rel}" && shasum -a 256 "${name}" > checksums.txt)
fi

run_install() { # $1 = install dir
  VERSION=v1.2.3 PRESS_RELEASE_BASE="file://${work}/release" INSTALL_DIR="$1" \
    bash "${root}/install.sh"
}

fail() { echo "FAIL: $*" >&2; exit 1; }

# 1. valid release installs
run_install "${work}/ok" >/dev/null || fail "valid release was rejected"
[ -x "${work}/ok/press" ] || fail "binary not installed"

# 2. tampered tarball is rejected and nothing is installed
printf 'tampered' >> "${rel}/${name}"
if run_install "${work}/bad" >/dev/null 2>&1; then fail "tampered tarball was accepted"; fi
[ ! -e "${work}/bad/press" ] || fail "tampered binary was installed"

# 3. missing checksums.txt is rejected
rm "${rel}/checksums.txt"
if run_install "${work}/nocs" >/dev/null 2>&1; then fail "missing checksums.txt was accepted"; fi
[ ! -e "${work}/nocs/press" ] || fail "binary installed without checksum"

echo "PASS"
