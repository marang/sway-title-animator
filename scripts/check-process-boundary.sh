#!/bin/sh
set -eu

animator=./cmd/sway-title-animator
module=github.com/marang/sway-title-animator

for removed in cmd/sway-session internal/session internal/sessionrequest internal/codexreport internal/statefile internal/herdrinit internal/diagnostic; do
	if [ -e "$removed" ]; then
		echo "session-only component remains in animator repository: $removed" >&2
		exit 1
	fi
done

for package in internal/session internal/sessionrequest internal/codexreport internal/statefile internal/herdrinit internal/diagnostic; do
	if go list -deps "$animator" | grep -Fqx "$module/$package"; then
		echo "animator depends on removed session package: $package" >&2
		exit 1
	fi
done

if go list -m all | awk '$1 == "github.com/marang/sway-session" || $1 == "modernc.org/sqlite" { found = 1 } END { exit(found ? 0 : 1) }'; then
	echo 'animator module graph retains sway-session or modernc.org/sqlite' >&2
	exit 1
fi

if find "$animator" -type f -name '*.go' ! -name '*_test.go' -print0 |
	xargs -0 grep -En 'state\.sqlite3|contexts\.json|layout\.json|session-start\.sock|codex-report\.sock|internal/(session|sessionrequest|codexreport|statefile|herdrinit|diagnostic)|github\.com/marang/sway-session|modernc\.org/sqlite' >/dev/null; then
	echo 'animator production code names a session state file, socket, or removed package' >&2
	exit 1
fi
