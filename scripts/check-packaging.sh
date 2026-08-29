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
require_fixed .goreleaser.yaml 'contrib/codex/hooks.json'
require_fixed .goreleaser.yaml 'contrib/codex/hooks-system.json'
require_fixed PKGBUILD 'contrib/codex/hooks-system.json'
require_fixed .goreleaser.yaml 'mode: 0755'
require_fixed PKGBUILD 'install -Dm755 scripts/verify-codex-boundary.sh'

for asset in \
	contrib/sway/45-title-animator.conf \
	contrib/herdr/config.toml \
	contrib/apparmor/codex-home-guard \
	scripts/verify-codex-boundary.sh
do
	require_fixed .goreleaser.yaml "$asset"
	require_fixed PKGBUILD "$asset"
done
