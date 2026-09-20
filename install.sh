#!/bin/sh
set -eu

# qualtrics-cli installer
# Usage: curl -fsSL https://raw.githubusercontent.com/thedavidweng/qualtrics-cli/main/install.sh | sh

REPO="thedavidweng/qualtrics-cli"
BINARY="qualtrics"

step()  { printf '==> %s\n' "$1"; }
die()   { printf 'ERROR: %s\n' "$1" >&2; exit 1; }

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) platform="darwin" ;;
  Linux)  platform="linux"  ;;
  *)      die "Unsupported OS: $os. Use install.ps1 on Windows." ;;
esac

case "$arch" in
  x86_64|amd64)  goarch="x86_64" ;;
  arm64|aarch64) goarch="arm64" ;;
  *)             die "Unsupported architecture: $arch" ;;
esac

platform_label="$platform/$goarch"

resolve_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/'
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O - "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/'
  else
    die "curl or wget is required."
  fi
}

download() {
  url="$1"
  output="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$output" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$output" "$url"
  fi
}

tag="${QUALTRICS_VERSION:-$(resolve_version)}"
[ -n "$tag" ] || die "Could not determine the latest release version."

version="${tag#v}"

archive_name="${BINARY}_${platform}_${goarch}.tar.gz"
archive_url="https://github.com/$REPO/releases/download/$tag/$archive_name"

step "Installing $BINARY $tag ($platform_label)..."

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

step "Downloading $archive_url"
download "$archive_url" "$tmp_dir/$archive_name" || die "Failed to download $archive_url"

step "Extracting..."
tar -xzf "$tmp_dir/$archive_name" -C "$tmp_dir" || die "Failed to extract archive"

install_dir="${QUALTRICS_INSTALL_DIR:-}"
if [ -z "$install_dir" ]; then
  if [ -w "/usr/local/bin" ]; then
    install_dir="/usr/local/bin"
  else
    install_dir="$HOME/.local/bin"
    mkdir -p "$install_dir"
  fi
fi

step "Installing to $install_dir/$BINARY"
mv "$tmp_dir/$BINARY" "$install_dir/$BINARY"
chmod +x "$install_dir/$BINARY"

step "Verifying installation..."
"$install_dir/$BINARY" version || true

printf '\nqualtrics installed successfully to %s/%s\n' "$install_dir" "$BINARY"
