#!/bin/sh
set -eu

REPO="makachanm/termfedi"
BASE_URL=${TEDI_BASE_URL:-"https://github.com/$REPO/releases/download/latest"}

fail() {
	echo "tedi install: $*" >&2
	exit 1
}

[ "$(uname -s)" = Linux ] || fail "only Linux is supported"

case $(uname -m) in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) fail "unsupported architecture: $(uname -m)" ;;
esac

fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$2" "$1"
	else
		fail "curl or wget is required"
	fi
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

archive="tedi_linux_$arch.tar.gz"
echo "Downloading $archive ..."
fetch "$BASE_URL/$archive" "$tmp/$archive"
fetch "$BASE_URL/checksums.txt" "$tmp/checksums.txt"

(cd "$tmp" && grep "  $archive\$" checksums.txt | sha256sum -c - >/dev/null) || fail "checksum mismatch"
tar -C "$tmp" -xzf "$tmp/$archive"

dir=${TEDI_INSTALL_DIR:-}
sudo=""
if [ -z "$dir" ]; then
	if [ -w /usr/local/bin ]; then
		dir=/usr/local/bin
	elif command -v sudo >/dev/null 2>&1; then
		dir=/usr/local/bin
		sudo=sudo
	else
		dir="$HOME/.local/bin"
	fi
fi

$sudo mkdir -p "$dir"
$sudo install -m 755 "$tmp/tedi_linux_$arch/tedi" "$dir/tedi"
echo "Installed tedi $("$dir/tedi" version) to $dir/tedi"

case ":$PATH:" in
*":$dir:"*) ;;
*) echo "Note: $dir is not in your PATH." ;;
esac
