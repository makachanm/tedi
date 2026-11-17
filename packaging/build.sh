#!/bin/sh
set -eu

VERSION=${VERSION:-0.0.0}
ARCHES=${ARCHES:-"amd64 arm64"}
NFPM=${NFPM:-nfpm}

ROOT=$(cd "$(dirname "$0")/.." && pwd)
DIST="$ROOT/dist"
export VERSION

rm -rf "$DIST"
mkdir -p "$DIST"

for arch in $ARCHES; do
	case $arch in
	amd64) rpmarch=x86_64 ;;
	arm64) rpmarch=aarch64 ;;
	*) echo "unsupported arch: $arch" >&2; exit 1 ;;
	esac

	name="tedi_linux_$arch"
	work="$DIST/$name"
	mkdir -p "$work"

	CGO_ENABLED=0 GOOS=linux GOARCH=$arch go build -C "$ROOT" -trimpath \
		-ldflags "-s -w -X main.version=$VERSION" -o "$work/tedi" .
	cp "$ROOT/README.md" "$ROOT/LICENSE" "$work/"
	tar -C "$DIST" -czf "$DIST/$name.tar.gz" "$name"

	(
		cd "$work"
		GOARCH=$arch "$NFPM" package -f "$ROOT/packaging/nfpm.yaml" -p deb -t "$DIST/tedi_$arch.deb"
		GOARCH=$arch "$NFPM" package -f "$ROOT/packaging/nfpm.yaml" -p rpm -t "$DIST/tedi_$rpmarch.rpm"
	)

	rm -rf "$work"
done

cd "$DIST"
sha256sum ./*.tar.gz ./*.deb ./*.rpm | sed 's# \./# #' > checksums.txt
