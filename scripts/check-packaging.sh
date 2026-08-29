#!/bin/sh

set -eu

require_fixed() {
	file=$1
	value=$2
	if ! grep -F -- "$value" "$file" >/dev/null; then
		echo "$file is missing required packaging entry: $value" >&2
		exit 1
	fi
}

require_fixed .goreleaser.yaml 'id: sway-session'
require_fixed .goreleaser.yaml 'main: ./cmd/sway-session'
require_fixed .goreleaser.yaml 'binary: sway-session'
require_fixed .goreleaser.yaml 'ids: [sway-title-animator, sway-session]'
require_fixed PKGBUILD 'go build -trimpath -ldflags="-s -w" -o sway-session ./cmd/sway-session'
require_fixed PKGBUILD 'install -Dm755 sway-session "$pkgdir/usr/bin/sway-session"'

for asset in \
	contrib/sway/45-title-animator.conf \
	contrib/herdr/config.toml \
	contrib/codex/hooks.json \
	contrib/apparmor/codex-home-guard
do
	require_fixed .goreleaser.yaml "$asset"
	require_fixed PKGBUILD "$asset"
done
