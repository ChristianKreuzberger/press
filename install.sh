#!/usr/bin/env bash
# install.sh — install the latest press binary from GitHub Releases
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/ChristianKreuzberger/press/main/install.sh | bash
#
# The script will:
#   1. Detect your OS and architecture.
#   2. Fetch the latest release tag from the GitHub API.
#   3. Download and extract the correct tarball.
#   4. Install the binary to /usr/local/bin (falls back to ~/.local/bin).

set -euo pipefail

REPO="ChristianKreuzberger/press"
BINARY="press"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"

# ── helpers ──────────────────────────────────────────────────────────────────

info()  { printf '\033[1;34m==> \033[0m%s\n' "$*"; }
ok()    { printf '\033[1;32m✓   \033[0m%s\n' "$*"; }
die()   { printf '\033[1;31mERROR: \033[0m%s\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "required tool not found: $1"
}

# ── detect OS ────────────────────────────────────────────────────────────────

detect_os() {
  local os
  os="$(uname -s)"
  case "${os}" in
    Linux*)   echo "linux"  ;;
    Darwin*)  echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) die "unsupported OS: ${os}" ;;
  esac
}

# ── detect architecture ───────────────────────────────────────────────────────

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "${arch}" in
    x86_64|amd64)   echo "amd64" ;;
    aarch64|arm64)  echo "arm64" ;;
    *) die "unsupported architecture: ${arch}" ;;
  esac
}

# ── fetch latest version ──────────────────────────────────────────────────────

fetch_latest_version() {
  local version
  if command -v curl >/dev/null 2>&1; then
    version="$(curl -fsSL "${GITHUB_API}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  elif command -v wget >/dev/null 2>&1; then
    version="$(wget -qO- "${GITHUB_API}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  else
    die "neither curl nor wget is available"
  fi
  [ -n "${version}" ] || die "could not determine latest release version"
  echo "${version}"
}

# ── download ──────────────────────────────────────────────────────────────────

download() {
  local url="$1" dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "${url}" -o "${dest}"
  else
    wget -qO "${dest}" "${url}"
  fi
}

# ── install ───────────────────────────────────────────────────────────────────

main() {
  need uname

  local os arch version archive ext install_dir tmp_dir

  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(fetch_latest_version)"
  # Strip leading 'v' if present for the archive name
  local ver_num="${version#v}"

  info "Installing ${BINARY} ${version} (${os}/${arch})"

  # Goreleaser names archives: press_<version>_<os>_<arch>.tar.gz (zip on Windows)
  if [ "${os}" = "windows" ]; then
    ext="zip"
    need unzip
  else
    ext="tar.gz"
    need tar
  fi

  local archive_name="${BINARY}_${ver_num}_${os}_${arch}.${ext}"
  local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "${tmp_dir}"' EXIT

  info "Downloading ${download_url}"
  download "${download_url}" "${tmp_dir}/${archive_name}"

  info "Extracting archive"
  if [ "${ext}" = "zip" ]; then
    unzip -q "${tmp_dir}/${archive_name}" -d "${tmp_dir}"
  else
    tar -xzf "${tmp_dir}/${archive_name}" -C "${tmp_dir}"
  fi

  local bin_src="${tmp_dir}/${BINARY}"
  [ "${os}" = "windows" ] && bin_src="${tmp_dir}/${BINARY}.exe"
  [ -f "${bin_src}" ] || die "binary not found in archive (expected: ${bin_src})"

  # Choose install directory
  if [ -w "/usr/local/bin" ]; then
    install_dir="/usr/local/bin"
  elif [ "$(id -u)" -eq 0 ]; then
    install_dir="/usr/local/bin"
    mkdir -p "${install_dir}"
  else
    install_dir="${HOME}/.local/bin"
    mkdir -p "${install_dir}"
  fi

  info "Installing to ${install_dir}/${BINARY}"
  install -m 755 "${bin_src}" "${install_dir}/${BINARY}"

  ok "${BINARY} ${version} installed to ${install_dir}/${BINARY}"

  # Warn if the install directory is not in PATH
  case ":${PATH}:" in
    *:"${install_dir}":*) ;;
    *)
      printf '\n\033[1;33mNote:\033[0m %s is not in your PATH.\n' "${install_dir}"
      printf 'Add the following line to your shell profile (e.g. ~/.bashrc or ~/.zshrc):\n'
      printf '\n  export PATH="%s:$PATH"\n\n' "${install_dir}"
      ;;
  esac
}

main "$@"
