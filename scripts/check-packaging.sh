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

require_count() {
	file=$1
	value=$2
	want=$3
	got=$(grep -F -c -- "$value" "$file" || true)
	if [ "$got" -ne "$want" ]; then
		echo "$file has $got copies of required packaging entry (want $want): $value" >&2
		exit 1
	fi
}

require_sequence() {
	file=$1
	first=$2
	second=$3
	third=$4
	if ! awk -v first="$first" -v second="$second" -v third="$third" '
		$0 == first {
			if ((getline next_line) > 0 && next_line == second &&
			    (getline final_line) > 0 && final_line == third) {
				found = 1
			}
		}
		END { exit(found ? 0 : 1) }
	' "$file"; then
		echo "$file is missing required associated packaging block: $first / $second / $third" >&2
		exit 1
	fi
}

reject_fixed() {
	file=$1
	value=$2
	if grep -F -- "$value" "$file" >/dev/null; then
		echo "$file contains forbidden packaging entry: $value" >&2
		exit 1
	fi
}

reject_regex() {
	file=$1
	pattern=$2
	if grep -E -- "$pattern" "$file" >/dev/null; then
		echo "$file contains forbidden packaging line: $pattern" >&2
		exit 1
	fi
}

require_fixed .goreleaser.yaml 'id: sway-title-animator'
require_fixed .goreleaser.yaml 'main: ./cmd/sway-title-animator'
require_fixed .goreleaser.yaml 'binary: sway-title-animator'
require_fixed .goreleaser.yaml 'ids: [sway-title-animator]'
require_count .goreleaser.yaml 'dependencies:' 1
require_sequence .goreleaser.yaml \
	'    dependencies:' \
	'      - sway' \
	'    contents:'
reject_fixed .goreleaser.yaml 'sway-session'
reject_fixed .goreleaser.yaml 'sway-herdr-init'
reject_fixed .goreleaser.yaml 'sqlite'
reject_regex .goreleaser.yaml '^[[:space:]]+(recommends|suggests):'
reject_fixed .goreleaser.yaml '      - golang'

require_fixed PKGBUILD '# Maintainer: marang <1550038+marang@users.noreply.github.com>'
require_fixed PKGBUILD '# Release template:'
require_fixed PKGBUILD "depends=('sway')"
require_fixed PKGBUILD "makedepends=('go>=1.26.5')"
require_fixed PKGBUILD "options=('!debug')"
require_fixed PKGBUILD '_go_build_flags=(-buildmode=pie -trimpath -buildvcs=false -mod=readonly -modcacherw)'
require_fixed PKGBUILD '_go_ldflags=(-s -w -buildid=)'
require_fixed PKGBUILD 'export GOTOOLCHAIN=local'
require_fixed PKGBUILD 'CGO_ENABLED=0 go build "${_go_build_flags[@]}" -ldflags="${_go_ldflags[*]}" -o sway-title-animator ./cmd/sway-title-animator'
require_fixed PKGBUILD 'CGO_ENABLED=0 go test "${_go_build_flags[@]}" -count=1 ./...'
require_fixed PKGBUILD 'install -Dm755 sway-title-animator "$pkgdir/usr/bin/sway-title-animator"'
require_fixed PKGBUILD 'install -Dm644 contrib/sway/45-title-animator.conf "$pkgdir/usr/share/doc/$pkgname/45-title-animator.conf"'
reject_regex PKGBUILD '^[[:space:]]*optdepends[[:space:]]*(\+)?='
reject_regex PKGBUILD '^[[:space:]]*depends\+'
reject_regex PKGBUILD '^depends=.*(alacritty|apparmor|flatpak|foot|glib2|go|herdr|libpulse|noto|python|sqlite)'
reject_fixed PKGBUILD '-o sway-session'
reject_fixed PKGBUILD 'install -Dm755 sway-session'
require_fixed PKGBUILD 'if [[ -d cmd/sway-session ]]; then'
reject_fixed PKGBUILD 'sway-herdr-init'
reject_fixed PKGBUILD 'sqlite:'
reject_fixed PKGBUILD 'noto-fonts:'

# .SRCINFO stays at the last released archive until the release workflow writes
# exact v0.10.0 metadata; it must still express the animator's runtime contract.
require_fixed .SRCINFO 'depends = sway'
require_fixed .SRCINFO 'makedepends = go>=1.26.5'
reject_regex .SRCINFO '^[[:space:]]*optdepends[[:space:]]*='
reject_fixed .SRCINFO 'depends = sqlite'
reject_fixed .SRCINFO 'optdepends = noto-fonts:'

require_count contrib/sway/45-title-animator.conf 'exec_always --no-startup-id /usr/bin/sway-title-animator --replace --fps 25' 1
reject_fixed contrib/sway/45-title-animator.conf 'sway-session'
require_fixed .goreleaser.yaml '      - src: ./contrib/sway/45-title-animator.conf'
require_fixed .goreleaser.yaml '        dst: /usr/share/doc/sway-title-animator/45-title-animator.conf'

# Keep the AUR and release-sync hardening checks added by LAB-111/PR #64.
require_fixed .github/workflows/aur.yml "grep -F 'SKIP' PKGBUILD"
require_fixed .github/workflows/aur.yml 'sed -i "s/^pkgrel=.*/pkgrel=1/" PKGBUILD'
require_fixed .github/workflows/aur.yml "grep -Fx 'pkgrel=1' PKGBUILD"
require_fixed .github/workflows/aur.yml 'makepkg --syncdeps --cleanbuild --clean --noconfirm'
require_fixed .github/workflows/aur.yml 'workflow_dispatch:'
require_fixed .github/workflows/aur.yml '          - verify-sync-token'
require_fixed .github/workflows/aur.yml "if: \${{ github.event_name == 'push' || inputs.operation == 'publish-release' }}"
require_fixed .github/workflows/aur.yml "if: \${{ github.event_name == 'workflow_dispatch' && inputs.operation == 'verify-sync-token' }}"
require_fixed .github/workflows/aur.yml 'VERSION: ${{ github.event.inputs.version || github.ref_name }}'
require_fixed .github/workflows/aur.yml 'ref: ${{ github.event.inputs.version || github.ref_name }}'
require_fixed .github/workflows/aur.yml 'show-ref --verify --quiet "refs/tags/$VERSION"'
require_fixed .github/workflows/aur.yml 'git -c safe.directory="$GITHUB_WORKSPACE"'
require_fixed .github/workflows/aur.yml 'merge-base --is-ancestor "$tag_commit" origin/main'
require_fixed .github/workflows/release.yml 'git merge-base --is-ancestor "$GITHUB_SHA" origin/main'
require_fixed .github/workflows/aur.yml 'actions/upload-artifact@v4'
require_fixed .github/workflows/aur.yml 'actions/download-artifact@v4'
require_fixed .github/workflows/aur.yml 'group: aur-release'
require_fixed .github/workflows/aur.yml 'persist-credentials: false'
require_fixed .github/workflows/aur.yml 'needs: publish-aur'
require_fixed .github/workflows/aur.yml 'RELEASE_SYNC_TOKEN'
require_fixed .github/workflows/aur.yml 'gh pr create'
require_fixed .github/workflows/aur.yml 'branch="automation/release-sync-check-${GITHUB_RUN_ID}-${GITHUB_RUN_ATTEMPT}"'
require_fixed .github/workflows/aur.yml 'gh pr close "$pr_url" --repo "$GITHUB_REPOSITORY"'
require_fixed .github/workflows/aur.yml 'printf '\''probe_sha=%s\n'\'' "$(git rev-parse HEAD)" >> "$GITHUB_OUTPUT"'
require_fixed .github/workflows/aur.yml 'if ! remote_ref=$(git ls-remote --heads origin'
require_fixed .github/workflows/aur.yml 'if [ -n "$remote_sha" ] && [ "$remote_sha" != "$PROBE_SHA" ]; then'
require_fixed .github/workflows/aur.yml '--force-with-lease="refs/heads/${PROBE_BRANCH}:${PROBE_SHA}"'
require_fixed .github/workflows/aur.yml 'origin ":refs/heads/${PROBE_BRANCH}"; then'
require_fixed .github/workflows/aur.yml 'GH_TOKEN: ${{ github.token }}'
require_fixed .github/workflows/aur.yml 'timeout 10m gh run watch "$run_id"'
require_fixed .github/workflows/aur.yml 'cp release-metadata/SRCINFO .SRCINFO'
require_fixed .github/workflows/aur.yml 'git checkout -B "$branch" origin/main'
require_fixed .github/workflows/aur.yml '--force-with-lease="refs/heads/${branch}:${remote_sha}"'
require_fixed .github/workflows/aur.yml 'aur.archlinux.org ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEuBKrPzbawxA/k2g6NcyV5jmqwJ2s+zpgZGZ7tpLIcN'
require_fixed .github/workflows/aur.yml 'SHA256:RFzBCUItH9LZS0cKB5UE6ceAYhBD5C8GeOBip8Z11+4'
require_fixed .github/workflows/aur.yml 'StrictHostKeyChecking=yes'
reject_fixed .github/workflows/aur.yml 'ssh-keyscan'
reject_fixed .github/workflows/aur.yml 'StrictHostKeyChecking=accept-new'

if git check-ignore --no-index --quiet .SRCINFO; then
	echo '.SRCINFO must remain trackable for release metadata sync PRs.' >&2
	exit 1
fi
