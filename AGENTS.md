# Repository guidance

## Project

`sway-title-animator` is a Linux-only Go program that renders animated Unicode
art in Sway title formats. It communicates directly with the Sway/i3 IPC socket
and remains usable without a registry or a running session manager.

## Required checks

Run the shared local/CI gate before handing off code changes:

```sh
make verify
```

This checks formatting, unit tests, race tests, `go vet`, `staticcheck`, the
`CGO_ENABLED=0` build, and whitespace errors. Use `make fmt` or `gofmt` on
changed Go files. The module requires Go 1.26; the preferred patched toolchain
is Go 1.26.5 as declared in `go.mod`.

## Workflow

`docs/workflow_conventions.md` is the canonical planning, branch, PR, review,
and cleanup workflow. Normal work belongs to the Linear `Lab` team's Sway Title
Animator P001 project and has the `Codebase` → `Sway Title Animator` label.

The repository split is an explicit cross-repository exception: LAB-119 is
coordinated in the new Sway Session project because it moves ownership between
both repositories. Record shared release and upgrade decisions in both
repositories until the cutover is complete.

Run the `code-review` skill before opening or finalizing a PR. Run
`review-codebase` after every fifth substantive branch or at a meaningful
project checkpoint.

## Architecture

- `cmd/sway-title-animator/main.go`: CLI parsing and process startup only.
- `daemon.go`: animation-only event subscription and frame loop.
- `animator.go`: title calculation, caching, and Sway title updates.
- `instance_lock.go`: single-instance lock and safe replacement.
- `config.go` / `model.go`: configuration and shared data types.
- `animations*.go`: pure animation rendering and deterministic seeded motion.
- `audio_meter.go`: optional `parec` capture and spectral analysis.
- `preview.go`: terminal preview and terminal-width handling.
- `internal/swayipc`: bounded i3/Sway IPC framing and reconnect behavior.
- `internal/titleindicator`: versioned presentation-only Sway-mark protocol;
  it contains no registry, session state, or restore logic.

Keep new responsibilities in the matching module instead of growing `main.go`.
Do not add dependencies on session state, session sockets, SQLite, or the
separate `sway-session` repository.

## Animation invariants

- Every animation returns exactly the requested terminal width after truncation.
- Widths at and near zero must not panic.
- Organic movement may vary between launches, but a fixed `animationSeed` must
  remain deterministic for tests.
- `square` uses connected scan-line glyphs, never Braille, and builds in place.
- New presets participate in the all-pairs visual-similarity guard; intentional
  relationships need a documented allowlist entry.
- Sound variants use `<base>_sound`, preserve the base preset's choreography,
  fall back safely when capture is unavailable, and remain opt-in.

## Audio and safety

Audio capture starts only for an active sound preset. `parec` failures must
degrade safely, produce one actionable diagnostic, avoid busy retries, and
stop promptly on cancellation. Keep FFT ranges ordered and non-overlapping;
never store captured audio.

- Never trust IPC payload lengths without a fixed upper bound.
- Never cache a title-format update that Sway rejected or failed to acknowledge.
- Never signal a PID from the instance file without validating executable and
  process start time.
- Preserve the exclusive instance lock for the full daemon lifetime.
